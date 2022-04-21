package initialize

import (
    "errors"
    "fmt"
    "github.com/ulule/deepcopier"
    "monitoring/config"
)

func GetMysqlConfig(myconfig *config.Config, pModuleName string, pGroupName string) (mysqlConfig config.Db, err error) {
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
