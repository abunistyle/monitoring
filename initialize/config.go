package initialize

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/ulule/deepcopier"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"io/ioutil"
	"monitoring/config"
	"monitoring/global"
	"monitoring/model/common"
	"monitoring/utils"
	"os"
)

func getDBConfig(myconfig *config.Config, pModuleName string, pGroupName string) (mysqlConfig config.Db, err error) {
	moduleExist := false
	for name, module := range myconfig.ModuleMap {
		if pModuleName != name {
			continue
		}
		mysqlConfig = config.Db{}
		err := deepcopier.Copy(module.Db).To(&mysqlConfig)
		if err != nil {
			panic(err)
		}
		moduleExist = true
		if module.GroupMap == nil {
			continue
		}
		for gName, group := range module.GroupMap {
			if pGroupName != gName {
				continue
			}

			if group.Db.Host != "" {
				mysqlConfig.Host = group.Db.Host
			}
			if group.Db.Port != 0 {
				mysqlConfig.Port = group.Db.Port
			}
			if group.Db.DbName != "" {
				mysqlConfig.DbName = group.Db.DbName
			}
			if group.Db.Username != "" {
				mysqlConfig.Username = group.Db.Username
			}
			if group.Db.Password != "" {
				mysqlConfig.Password = group.Db.Password
			}
			if group.Db.Config != "" {
				mysqlConfig.Config = group.Db.Config
			}
		}
	}
	if !moduleExist {
		err = errors.New(fmt.Sprintf("module %s not exist!\n", pModuleName))
	}
	return mysqlConfig, err
}

func getDBConfigList(myconfig *config.Config) (map[string]string, map[string]config.Db) {
	dbConfigIndexMap := make(map[string]string)
	dbConfigMap := make(map[string]config.Db)
	for moduleName, module := range myconfig.ModuleMap {
		//默认DB
		dbConfig, err := getDBConfig(myconfig, moduleName, "")
		if err != nil {
			fmt.Println(err.Error())
		} else {
			key1 := utils.GenDbIndexKey(moduleName, "")
			key2 := utils.MD5(fmt.Sprintf("%s|%d|%s|%s|%s", dbConfig.Host, dbConfig.Port, dbConfig.DbName, dbConfig.Config, dbConfig.Username))
			_, key2Exist := dbConfigMap[key2]
			if !key2Exist {
				dbConfigMap[key2] = dbConfig
			}
			dbConfigIndexMap[key1] = key2
		}
		for groupName := range module.GroupMap {
			dbConfig, err := getDBConfig(myconfig, moduleName, groupName)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			key1 := utils.GenDbIndexKey(moduleName, groupName)
			key2 := utils.MD5(fmt.Sprintf("%s|%d|%s|%s", dbConfig.Host, dbConfig.Port, dbConfig.DbName, dbConfig.Config, dbConfig.Username))
			_, key2Exist := dbConfigMap[key2]
			if !key2Exist {
				dbConfigMap[key2] = dbConfig
			}
			dbConfigIndexMap[key1] = key2
		}
	}

	return dbConfigIndexMap, dbConfigMap
}

func InitDBList(myconfig *config.Config) {
	dbConfigIndexMap, dbConfigMap := getDBConfigList(myconfig)
	dbMap := make(map[string]*gorm.DB)
	if global.DBMap == nil {
		global.DBMap = make(map[string]*gorm.DB)
	}
	for key, dbConfig := range dbConfigMap {
		db := GormMysqlByConfig(dbConfig)
		global.DBList = append(global.DBList, db)
		dbMap[key] = db
	}
	for key, indexKey := range dbConfigIndexMap {
		_, indexKeyExist := dbMap[indexKey]
		if !indexKeyExist {
			continue
		}
		global.SetDBByKey(key, dbMap[indexKey])
	}
}

func InitRedis(myconfig *config.Config) {
	client := redis.NewClient(&redis.Options{
		Addr:     myconfig.Redis.Addr,
		Password: myconfig.Redis.Password,
		DB:       myconfig.Redis.Db,
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err.Error())
	}
	global.RedisClient = client
}

// InitParam 初始化命令参数
func InitParam() common.Param {
	param := common.Param{}
	param.RootPath = flag.String("r", "./", "请输入程序根路径")
	param.ConfigFileName = flag.String("c", "./config.yaml", "请输入配置文件路径")
	param.ModuleName = flag.String("m", "web", "请输入模块名，比如：web、fdweb")
	param.Debug = flag.Bool("debug", false, "是否debug，比如：true、false")
	param.Port = flag.Uint64("p", 0, "请输入端口号，比如：8080")
	flag.Parse()
	file, err := ioutil.ReadFile(*param.ConfigFileName)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
		return param
	}
	//yaml文件内容影射到结构体中
	var myConfig config.Config
	err1 := yaml.Unmarshal(file, &myConfig)
	if err1 != nil {
		fmt.Println(err1)
		os.Exit(1)
		return param
	}
	*param.Port = (func() uint64 {
		if *param.Port > 0 {
			return *param.Port
		} else {
			return myConfig.Application.Port
		}
	})()
	InitDBList(&myConfig)
	return param
}

// InitDB 连接数据库
func InitDB(moduleName string, paramDb **gorm.DB) bool {
	db, dbExist := global.GetDBByName(moduleName, "")
	*paramDb = db
	if !dbExist {
		err := errors.New(fmt.Sprintf("module:%s, db not exist!", moduleName))
		panic(err)
		return false
	}
	return true
}
