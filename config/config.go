package config

import "strconv"

type Config struct {
    Application Application       `yaml:"application"`
    Redis       Redis             `yaml:"redis"`
    ModuleMap   map[string]Module `yaml:"modules"`
}

type Application struct {
    Port uint64 `yaml:"port"`
}

type Redis struct {
    Addr     string `yaml:"addr"`
    Password string `yaml:"password"`
    Db       int    `yaml:"db"`
}

type Module struct {
    Db       Db               `yaml:"db"`
    GroupMap map[string]Group `yaml:"groupMap"`
}

type Group struct {
    Name string `yaml:"name"`
    Db   Db     `yaml:"db"`
}

type Db struct {
    Host         string `yaml:"host"`
    Port         uint64 `yaml:"port"`
    DbName       string `yaml:"dbName"`
    Username     string `yaml:"username"`
    Password     string `yaml:"password"`
    Config       string `yaml:"config"`
    MaxOpenConns int    `yaml:"maxOpenConns"`
    MaxIdleConns int    `yaml:"maxIdleConns"`
}

func (m *Db) Dsn() string {
    return m.Username + ":" + m.Password + "@tcp(" + m.Host + ":" + strconv.FormatUint(m.Port, 10) + ")/" + m.DbName + "?" + m.Config
}
