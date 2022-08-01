package initialize

import (
    "errors"
    "fmt"
    "github.com/go-redis/redis"
    "github.com/ulule/deepcopier"
    "gorm.io/gorm"
    "monitoring/config"
    "monitoring/global"
    "monitoring/utils"
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
