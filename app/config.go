package app

// Config is required configuration for app
type Config struct {
	HTTPPort                     string `json:"HTTPPort"`
	EncryptKey                   string `json:"encKey"`
	IndexTemplatePath            string `json:"indexTemplatePath"`
	AppDatabasePath              string `json:"appDatabasePath"`
	StaticResourcesDirectoryPath string `json:"staticResourcesDirPath"`

	RateLimit          string `json:"rateLimit"`
	RobotsTextFilePath string `json:"robotsTxtPath"`

	EnableCors    bool `json:"enableCors"`
	HaveRobotsTxt bool `json:"haveRobotsTxt"`
	PingGoogle    bool `json:"pingGoogle"`
	FetchPreviews bool `json:"fetchPreviews"`
}
