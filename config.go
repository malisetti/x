package main

type config struct {
	HTTPPort                     string `json:"HTTPPort"`
	EncryptKey                   string `json:"encKey"`
	IndexTemplatePath            string `json:"indexTemplatePath"`
	AppDatabasePath              string `json:"appDatabasePath"`
	StaticResourcesDirectoryPath string `json:"staticResourcesDirPath"`
	RobotsTextFilePath           string `json:"robotsTxtPath"`

	RateLimit string `json:"rateLimit"`

	TweetItems bool `json:"tweetItems"`
	PingGoogle bool `json:"pingGoogle"`

	TwitterAccessToken       string `json:"twitterAccessToken"`
	TwitterAccessTokenSecret string `json:"twitterAccessTokenSecret"`
	TwitterConsumerAPIKey    string `json:"twitterConsumerAPIKey"`
	TwitterConsumerSecretKey string `json:"twitterConsumerSecretKey"`
}
