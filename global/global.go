package global

import (
    "gorm.io/gorm"
    "monitoring/utils"
)

var (
    DBList []*gorm.DB
    DBMap  map[string]*gorm.DB
)

func GetDBByName(moduleName string, groupName string) (*gorm.DB, bool) {
    key := utils.GenDbIndexKey(moduleName, groupName)
    db, keyExist := DBMap[key]
    return db, keyExist
}

func GetDBByKey(key string) (*gorm.DB, bool) {
    db, keyExist := DBMap[key]
    return db, keyExist
}

func SetDBByName(moduleName string, groupName string, DB *gorm.DB) {
    if DBMap == nil {
        DBMap = make(map[string]*gorm.DB)
    }
    key := utils.GenDbIndexKey(moduleName, groupName)
    DBMap[key] = DB
}

func SetDBByKey(key string, DB *gorm.DB) {
    if DBMap == nil {
        DBMap = make(map[string]*gorm.DB)
    }
    DBMap[key] = DB
}
