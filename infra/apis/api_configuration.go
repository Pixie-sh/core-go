package apis

type RestAPIConfiguration struct {
	APIConfiguration

	Host string `json:"host"`
}

type APIConfiguration struct {
	ApiKey  string `json:"api_key"`
	Timeout string `json:"timeout"`
}
