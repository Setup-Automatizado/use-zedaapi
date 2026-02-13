package webshare

// ProxyListResponse is the paginated response from Webshare proxy list API.
type ProxyListResponse struct {
	Count    int         `json:"count"`
	Next     *string     `json:"next"`
	Previous *string     `json:"previous"`
	Results  []ProxyItem `json:"results"`
}

// ProxyItem represents a single proxy from Webshare API.
type ProxyItem struct {
	ID                    string  `json:"id"`
	Username              string  `json:"username"`
	Password              string  `json:"password"`
	ProxyAddress          string  `json:"proxy_address"`
	Port                  int     `json:"port"`
	Valid                 bool    `json:"valid"`
	LastVerification      *string `json:"last_verification"`
	CountryCode           string  `json:"country_code"`
	CityName              string  `json:"city_name"`
	ASNName               string  `json:"asn_name"`
	ASNNumber             int     `json:"asn_number"`
	HighCountryConfidence bool    `json:"high_country_confidence"`
	CreatedAt             string  `json:"created_at"`
}

// ReplaceRequest is the body for the proxy replace API (direct mode).
type ReplaceRequest struct {
	ProxyAddress string `json:"proxy_address"`
	Port         int    `json:"port"`
}

// ReplaceByIDRequest is the body for the proxy replace API (backbone mode).
type ReplaceByIDRequest struct {
	ProxyID string `json:"id"`
}

// ReplaceResponse is the response from the proxy replace API.
type ReplaceResponse struct {
	ID           string `json:"id"`
	ProxyAddress string `json:"proxy_address"`
	Port         int    `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CountryCode  string `json:"country_code"`
	CityName     string `json:"city_name"`
	Valid        bool   `json:"valid"`
}

// APIError represents a Webshare API error response.
type APIError struct {
	Detail string `json:"detail"`
	Code   string `json:"code"`
}
