// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package auth

import (
	"github.com/redis/go-redis/v9"
	"time"
)

// Recovered from runtime type 0x12b85e0; original size 8 bytes.
type BruteForce struct {
	rdb *redis.Client
}

// Recovered from runtime type 0x12e8120; original size 16 bytes.
type CaptchaService struct {
	rdb   *redis.Client
	click *ClickCaptchaService
}

// Recovered from runtime type 0x12b8660; original size 8 bytes.
type ClickCaptchaService struct {
	rdb *redis.Client
}

// Recovered from runtime type 0x1328f20; original size 64 bytes.
type GenerateResult struct {
	ID     string
	Image  string
	Type   string
	Prompt string
}

// Recovered from runtime type 0x130a440; original size 32 bytes.
type SDKSession struct {
	AESKey     string `json:"aesKey"`
	AppID      int64  `json:"appid"`
	CreateTime int64  `json:"createTime"`
}

// Recovered from runtime type 0x12e81c0; original size 16 bytes.
type SDKSessionStore struct {
	rdb *redis.Client
	ttl time.Duration
}

// Recovered from runtime type 0x1349c80; original size 80 bytes.
type Session struct {
	Credential string `json:"credential,omitempty"`
	AdminID    int64  `json:"adminId"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Role       string `json:"role"`
	IssuedAt   int64  `json:"issuedAt"`
}

// Recovered from runtime type 0x130a5c0; original size 24 bytes.
type SessionStore struct {
	TTLProvider func() time.Duration
	rdb         *redis.Client
	ttl         time.Duration
	absoluteTTL time.Duration
}

// Recovered from runtime type 0x1365b60; original size 104 bytes.
type UserSession struct {
	Revision        string `json:"revision,omitempty"`
	LastActiveAt    int64  `json:"lastActiveAt"`
	LastHeartbeatAt int64  `json:"lastHeartbeatAt"`
	LoginIP         string `json:"loginIp"`
	DeviceID        int64  `json:"deviceId,omitempty"`
	Credential      string `json:"credential,omitempty"`
	UID             int64  `json:"uid"`
	AppID           int64  `json:"appid"`
	Username        string `json:"username"`
	Mi              string `json:"mi"`
	LastIP          string `json:"lastIp"`
	Region          string `json:"region"`
	IssuedAt        int64  `json:"issuedAt"`
	TTLSec          int    `json:"ttlSec,omitempty"`
	IsKami          bool   `json:"isKami,omitempty"`
}

// Recovered from runtime type 0x12e8260; original size 16 bytes.
type UserSessionStore struct {
	rdb *redis.Client
	ttl time.Duration
}
