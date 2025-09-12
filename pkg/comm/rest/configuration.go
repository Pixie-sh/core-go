package rest

// ClientConfiguration rest client config
type ClientConfiguration struct {
	HeaderAPIKey string            `json:"header_api_key" `
	APIKeys      map[string]string `json:"api_keys" `
	Timeout      int               `json:"timeout" `
}
