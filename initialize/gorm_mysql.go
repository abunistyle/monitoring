package initialize

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "monitoring/config"
)

// GormMysqlByConfig 初始化Mysql数据库用过传入配置
func GormMysqlByConfig(dbConfig config.Db) *gorm.DB {
    if dbConfig.DbName == "" {
        return nil
    }
    mysqlConfig := mysql.Config{
        DSN:                       dbConfig.Dsn(), // DSN data source name
        DefaultStringSize:         191,            // string 类型字段的默认长度
        SkipInitializeWithVersion: false,          // 根据版本自动配置
    }
    if db, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{}); err != nil {
        panic(err)
    } else {
        sqlDB, _ := db.DB()
        sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
        sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
        return db
    }
}
