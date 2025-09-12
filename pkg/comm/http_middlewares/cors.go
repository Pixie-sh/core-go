package http_middlewares

// CorsConfiguration configurations
type CorsConfiguration struct {
	AllowOrigins     string `json:"allow_origins"`
	AllowMethods     string `json:"allow_methods"`
	AllowHeaders     string `json:"allow_headers"`
	AllowCredentials bool   `json:"allow_credentials"`
	ExposeHeaders    string `json:"expose_headers"`
	MaxAge           int    `json:"max_age"`
}
