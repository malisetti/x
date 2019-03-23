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

	HaveRobotsTxt bool `json:"haveRobotsTxt"`
	TweetItems    bool `json:"tweetItems"`
	PingGoogle    bool `json:"pingGoogle"`
	FetchPreviews bool `json:"fetchPreviews"`

	TwitterAccessToken       string `json:"twitterAccessToken"`
	TwitterAccessTokenSecret string `json:"twitterAccessTokenSecret"`
	TwitterConsumerAPIKey    string `json:"twitterConsumerAPIKey"`
	TwitterConsumerSecretKey string `json:"twitterConsumerSecretKey"`
}
