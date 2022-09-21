package common

type Param struct {
	RootPath       *string `json:"rootPath"`
	ConfigFileName *string `json:"configFileName"`
	ModuleName     *string `json:"moduleName"`
	Debug          *bool   `json:"debug"`
	Port           *uint64 `json:"port"`
}
