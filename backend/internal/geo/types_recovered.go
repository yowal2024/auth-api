// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package geo

import (
	"auth-api/internal/repo"
)

// Recovered from runtime type 0x1349e80; original size 72 bytes.
type IPEntry struct {
	IP       string  `json:"ip"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Province string  `json:"province,omitempty"`
	City     string  `json:"city,omitempty"`
	Time     int64   `json:"time"`
}

// Recovered from runtime type 0x12b87e0; original size 16 bytes.
type Amap struct {
	Key string
}

// Recovered from runtime type 0x12b88e0; original size 8 bytes.
type Checker struct {
	ipInfo *repo.IPInfoRepo
}

// Recovered from runtime type 0x1349d80; original size 80 bytes.
type Point struct {
	Lat      float64
	Lng      float64
	Nation   string
	Province string
	City     string
	District string
}

// Recovered from runtime type 0x12b8860; original size 16 bytes.
type Tencent struct {
	Key string
}

// Recovered from runtime type 0x1349f80; original size 96 bytes.
type amapResp struct {
	Status    string `json:"status"`
	Info      string `json:"info"`
	Province  string `json:"province"`
	City      string `json:"city"`
	Adcode    string `json:"adcode"`
	Rectangle string `json:"rectangle"`
}

// Recovered from runtime type 0x130a800; original size 120 bytes.
type tencentResp struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		IP       string "json:\"ip\""
		Location struct {
			Lat float64 "json:\"lat\""
			Lng float64 "json:\"lng\""
		} "json:\"location\""
		AdInfo struct {
			Nation   string "json:\"nation\""
			Province string "json:\"province\""
			City     string "json:\"city\""
			District string "json:\"district\""
		} "json:\"ad_info\""
	} `json:"result"`
}
