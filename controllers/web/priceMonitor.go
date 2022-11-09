package web

import (
    "fmt"
    "gorm.io/gorm"
    "log"
    "monitoring/utils"
    "net/http"
    "net/url"
    "strconv"
    "strings"
)

type PriceMonitor struct {
    DB           *gorm.DB
    ProjectNames []string
    Debug        bool
}

type GoodsSizeColorProjectInventory struct {
    GoodsId      int64  `json:"goods_id"`
    ProjectName  string `json:"project_name"`
    StyleColorId int64  `json:"style_color_id"`
    StyleSizeId  int64  `json:"style_size_id"`
    Color        string `json:"color"`
    Size         string `json:"size"`
    ShopPrice    string `json:"shop_price"`
}

type SkuInfo struct {
    GoodsId int64  `json:"goods_id"`
    ColorId int64  `json:"color_id"`
    SizeId  int64  `json:"size_id"`
    Color   string `json:"color"`
    Size    string `json:"size"`
}

//创建结构体及对应的指标信息
func (p *PriceMonitor) RunMonitor() {
    log.Println("zero price monitor start!")
    for _, projectName := range p.ProjectNames {
        goodsIdList := p.GetZeroSkuPriceGoodsIds(projectName)
        goodsIdList2 := p.GetZeroGoodsPriceGoodsIds(projectName)
        allGoodsIdList := append(goodsIdList, goodsIdList2...)
        allGoodsIdList = utils.ArrayUnique(allGoodsIdList)
        log.Println(allGoodsIdList)

        if len(allGoodsIdList) == 0 {
            continue
        }

        var allGoodsIdStrList []string
        for _, goodsId := range allGoodsIdList {
            allGoodsIdStrList = append(allGoodsIdStrList, strconv.Itoa(int(goodsId)))
        }
        msg := fmt.Sprintf("[0元商品]\n组织：%s\ngoodsid：%s\n在架sku销售价为0", projectName, strings.Join(allGoodsIdStrList, ","))
        log.Print(msg)
        p.RunSendNotice(msg)
    }
    log.Println("zero price monitor end!")
}

func (p *PriceMonitor) GetZeroSkuPriceGoodsIds(projectName string) []int64 {
    var goodsIdList []int64
    skuList := p.GetZeroSkuPriceList(projectName)
    for _, sku := range skuList {
        goodsIdList = append(goodsIdList, sku.GoodsId)
    }
    goodsIdList = utils.ArrayUnique(goodsIdList)
    return goodsIdList
}

func (p *PriceMonitor) GetZeroSkuPriceList(projectName string) []GoodsSizeColorProjectInventory {
    var result []GoodsSizeColorProjectInventory
    sql := fmt.Sprintf("select\n    gscpi.goods_id,\n    gscpi.project_name,\n    gscpi.style_color_id,\n    gscpi.style_size_id,\n    gscpi.color,\n    gscpi.size,\n    gscpi.shop_price\nfrom goods_size_color_project_inventory gscpi\n     left join goods_project gp on gp.goods_id = gscpi.goods_id\n        and gp.project_name = gscpi.project_name\n        and gp.is_on_sale = 1\n     left join goods_style_black_white gsbw1 on gscpi.goods_id = gsbw1.goods_id\n        and gsbw1.style_id = gscpi.style_size_id\n        and gsbw1.style_name = 'size'\n        and gsbw1.black_white = 'white'\n     left join goods_style_black_white gsbw2 on gscpi.goods_id = gsbw2.goods_id\n        and gsbw2.style_id = gscpi.style_color_id\n        and gsbw2.style_name = 'color'\n        and gsbw2.black_white = 'white'\n     left join goods_style_off_sale gsos on gsos.goods_id = gscpi.goods_id\n        and gsos.color_style_id = gscpi.style_color_id\n        and gsos.size_style_id = gscpi.style_size_id\n        and gsos.project_name in ('default', '%s')\n        and gsos.country = 'all'\nwhere gscpi.project_name = '%s'\n  and gscpi.shop_price = '%s'\n  and gp.goods_id is not null\n  and gsbw1.bw_id is not null\n  and gsbw2.bw_id is not null\n  and gsos.id is null", projectName, projectName, "0")
    p.DB.Raw(sql).Scan(&result)
    return result
}

func (p *PriceMonitor) GetZeroGoodsPriceGoodsIds(projectName string) []int64 {
    goodsIdList := p.GetZeroGoodsPriceList(projectName)
    goodsIdList = p.FilterGiftGoods(goodsIdList)
    goodsIdList = p.FilterHasSkuPriceGoods(projectName, goodsIdList)
    return goodsIdList
}

func (p *PriceMonitor) GetZeroGoodsPriceList(projectName string) []int64 {
    var result []int64
    sql := fmt.Sprintf("select\n    gp.goods_id\nfrom goods_project as gp\nwhere gp.is_on_sale = 1\n  and gp.is_delete = 0\n  and gp.is_display = 1\n  and gp.project_name = '%s'\n  and gp.shop_price = %s", projectName, "0")
    p.DB.Raw(sql).Scan(&result)
    return result
}

func (p *PriceMonitor) FilterGiftGoods(goodsIds []int64) []int64 {
    var result []int64
    for _, goodsId := range goodsIds {
        if !p.IsGiftGoods(goodsId) {
            result = append(result, goodsId)
        }
    }
    return result
}

func (p *PriceMonitor) FilterHasSkuPriceGoods(projectName string, goodsIds []int64) []int64 {
    var result []int64
    for _, goodsId := range goodsIds {
        if p.IsNoSkuPriceGoods(projectName, goodsId) {
            result = append(result, goodsId)
        }
    }
    return result
}

func (p *PriceMonitor) IsGiftGoods(goodsId int64) bool {
    var goodsIds []int64
    sql := fmt.Sprintf("select\n    ga.goods_id\nfrom goods_attr as ga\n     join attribute a on ga.attr_id = a.attr_id\nwhere ga.goods_id = %d\n  and a.is_delete = 0\n  and ga.is_delete = 0\n  and a.attr_name = 'Is Gift'\n  and a.attr_values = 'yes'", goodsId)
    p.DB.Raw(sql).Scan(&goodsIds)
    return len(goodsIds) > 0
}

func (p *PriceMonitor) IsNoSkuPriceGoods(projectName string, goodsId int64) bool {
    var result []SkuInfo
    sql := fmt.Sprintf("select\n    g.goods_id,\n    gsbwc.style_value as color,\n    gsbws.style_value as size,\n    gsbwc.style_id as color_id,\n    gsbws.style_id as size_id\nfrom goods g\n    left join goods_style_black_white gsbwc on g.goods_id = gsbwc.goods_id\n        and gsbwc.style_name = 'color'\n        and gsbwc.black_white = 'white'\n    left join goods_style_black_white gsbws on g.goods_id = gsbws.goods_id\n        and gsbws.style_name = 'size'\n        and gsbws.black_white = 'white'\n    left join goods_style_off_sale gsos on gsos.goods_id = g.goods_id\n        and gsos.color_style_id = gsbwc.style_id\n        and gsos.size_style_id = gsbws.style_id\n        and gsos.project_name in ('default','%s')\n        and gsos.country = 'all'\n    left join goods_size_color_project_inventory gscpi on gscpi.goods_id = g.goods_id\n        and gscpi.style_color_id = gsbwc.style_id\n        and gscpi.style_size_id = gsbws.style_id\n        and gscpi.project_name = '%s'\nwhere g.goods_id = %d\n    and gsbwc.bw_id is not null\n    and gsbws.bw_id is not null\n    and gsos.id is null\n    and gscpi.inventory_id is not null", projectName, projectName, goodsId)
    p.DB.Raw(sql).Scan(&result)
    return len(result) <= 0
}

func (p *PriceMonitor) RunSendNotice(message string) {
    if p.Debug {
        return
    }
    //if p.Debug {
    //    message = "(测试中，请忽略)" + message
    //}
    go func() {
        resp, err := http.Get(fmt.Sprintf("http://voice.arch800.com/notice/singleCallByTts?system=Monitoring&errorMsg=%s", url.QueryEscape(message)))
        if err != nil {
            fmt.Println(err)
            return
        }
        fmt.Println(resp)
    }()
}
