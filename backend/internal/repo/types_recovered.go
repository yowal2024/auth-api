// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package repo

import (
	"github.com/jmoiron/sqlx"
	"sync/atomic"
)

// Recovered from runtime type 0x134a680; original size 64 bytes.
type APICode struct {
	ID           int64  `db:"id"            json:"id"`
	Code         int    `db:"code"          json:"code"`
	OverrideCode int    `db:"override_code" json:"overrideCode"`
	Message      string `db:"message"       json:"message"`
	Remark       string `db:"remark"        json:"remark"`
	Rewritten    int    `db:"rewritten"     json:"rewritten"`
}

// Recovered from runtime type 0x134a780; original size 72 bytes.
type APIRoute struct {
	ID        int64  `db:"id"            json:"id"`
	APIKey    string `db:"api_key"       json:"apiKey"`
	Name      string `db:"name"          json:"name"`
	Override  string `db:"override_path" json:"overridePath"`
	Rewritten int    `db:"rewritten"     json:"rewritten"`
	SortOrder int    `db:"sort_order"    json:"sortOrder"`
}

// Recovered from runtime type 0x13604c0; original size 112 bytes.
type Admin struct {
	ID        int64  `db:"id"           json:"id"`
	Username  string `db:"username"     json:"username"`
	Nickname  string `db:"nickname"     json:"nickname"`
	Avatar    string `db:"avatar"       json:"avatar"`
	Password  string `db:"password"     json:"-"`
	Role      string `db:"role"         json:"role"`
	LastLogin int64  `db:"last_login"   json:"lastLogin"`
	LastIP    string `db:"last_ip"      json:"lastIp"`
}

// Recovered from runtime type 0x1395be0; original size 192 bytes.
type App struct {
	ID            int64  `db:"id"             json:"id"`
	Name          string `db:"name"           json:"name"`
	Type          string `db:"type"           json:"type"`
	Mode          string `db:"mode"           json:"mode"`
	Logo          string `db:"logo"           json:"logo"`
	Pkg           string `db:"pkg"            json:"pkg"`
	TokenKey      string `db:"token_key"      json:"tokenKey"`
	BindDevice    bool   `db:"bind_device"    json:"bindDevice"`
	BindRegion    bool   `db:"bind_region"    json:"bindRegion"`
	DiffUpdate    bool   `db:"diff_update"    json:"diffUpdate"`
	RandomPackage bool   `db:"random_package" json:"randomPackage"`
	Encryption    int64  `db:"encryption"     json:"encryption"`
	UserTokenTTL  int    `db:"user_token_ttl" json:"userTokenTtl"`
	GeoProvider   string `db:"geo_provider"   json:"geoProvider"`
	GeoKey        string `db:"geo_key"        json:"geoKey"`
	IPDistanceMax int    `db:"ip_distance_max" json:"ipDistanceMax"`
	BruteForce    bool   `db:"brute_force_enabled" json:"bruteForceEnabled"`
	MtlsEnabled   bool   `db:"mtls_enabled"   json:"mtlsEnabled"`
	Status        int    `db:"status"         json:"status"`
	CreateTime    int64  `db:"create_time"    json:"createTime"`
}

// Recovered from runtime type 0x12e8da0; original size 24 bytes.
type AppCA struct {
	ID     int64  `db:"id"`
	CACert string `db:"mtls_ca_cert"`
}

// Recovered from runtime type 0x134a880; original size 72 bytes.
type Block struct {
	ID         int64  `db:"id"          json:"id"`
	AppID      int64  `db:"appid"       json:"appid"`
	Type       string `db:"type"        json:"type"`
	Value      string `db:"value"       json:"value"`
	Reason     string `db:"reason"      json:"reason"`
	CreateTime int64  `db:"create_time" json:"createTime"`
}

// Recovered from runtime type 0x1380740; original size 176 bytes.
type Encrypt struct {
	ID               int64  `db:"id"                  json:"id"`
	AppID            int64  `db:"appid"               json:"appid"`
	Name             string `db:"name"                json:"name"`
	Algorithm        string `db:"algorithm"           json:"algorithm"`
	Encoding         string `db:"encoding"            json:"encoding"`
	SymKey           string `db:"sym_key"             json:"symKey"`
	TimeSkew         int    `db:"time_skew"           json:"timeSkew"`
	SignMode         string `db:"sign_mode"           json:"signMode"`
	ClientPrivateKey string `db:"client_private_key"  json:"clientPrivateKey"`
	ClientPublicKey  string `db:"client_public_key"   json:"clientPublicKey"`
	ServerPrivateKey string `db:"server_private_key"  json:"serverPrivateKey"`
	ServerPublicKey  string `db:"server_public_key"   json:"serverPublicKey"`
	CreateTime       int64  `db:"create_time"         json:"createTime"`
}

// Recovered from runtime type 0x134a980; original size 72 bytes.
type Extend struct {
	ID         int64  `db:"id"          json:"id"`
	AppID      int64  `db:"appid"       json:"appid"`
	Key        string `db:"e_key"       json:"key"`
	Value      string `db:"e_value"     json:"value"`
	Remark     string `db:"remark"      json:"remark"`
	UpdateTime int64  `db:"update_time" json:"updateTime"`
}

// Recovered from runtime type 0x1385d80; original size 176 bytes.
type ExtendFile struct {
	ID            int64  `db:"id"             json:"id"`
	AppID         int64  `db:"appid"          json:"appid"`
	Name          string `db:"name"           json:"name"`
	Filename      string `db:"filename"       json:"filename"`
	Filepath      string `db:"filepath"       json:"filepath"`
	Filesize      int64  `db:"filesize"       json:"filesize"`
	Filetype      string `db:"filetype"       json:"filetype"`
	Category      string `db:"category"       json:"category"`
	Remark        string `db:"remark"         json:"remark"`
	Version       string `db:"version"        json:"version"`
	DownloadURL   string `db:"download_url"   json:"downloadUrl"`
	DownloadCount int64  `db:"download_count" json:"downloadCount"`
	CreatedAt     int64  `db:"created_at"     json:"createdAt"`
	UpdatedAt     int64  `db:"updated_at"     json:"updatedAt"`
}

// Recovered from runtime type 0x1366060; original size 104 bytes.
type Function struct {
	ID      int64  `db:"id"       json:"id"`
	AppID   int64  `db:"appid"    json:"appid"`
	Name    string `db:"name"     json:"name"`
	Trigger string `db:"trigger"  json:"trigger"`
	Code    string `db:"code"     json:"code"`
	Remark  string `db:"remark"   json:"remark"`
	Timeout int    `db:"timeout"  json:"timeout"`
	Enabled int    `db:"enabled"  json:"enabled"`
	LastRun int64  `db:"last_run" json:"lastRun"`
}

// Recovered from runtime type 0x1398b40; original size 264 bytes.
type Kami struct {
	ID             int64  `db:"id"          json:"id"`
	AppID          int64  `db:"appid"       json:"appid"`
	GroupID        int64  `db:"group_id"    json:"groupId"`
	GroupName      string `db:"group_name"  json:"groupName"`
	Card           string `db:"card"        json:"card"`
	Status         int    `db:"status"      json:"status"`
	UseUser        string `db:"use_user"    json:"useUser"`
	UseTime        int64  `db:"use_time"    json:"useTime"`
	Remark         string `db:"remark"      json:"remark"`
	CreateTime     int64  `db:"create_time" json:"createTime"`
	Creator        string `db:"creator"     json:"creator"`
	CreateIP       string `db:"create_ip"   json:"createIp"`
	Nickname       string `db:"nickname"          json:"nickname"`
	Vip            int64  `db:"vip"               json:"vip"`
	BanExpire      int64  `db:"ban_expire"        json:"banExpire"`
	BindLimit      int    `db:"bind_limit"        json:"bindLimit"`
	RegionLimit    int    `db:"region_limit"      json:"regionLimit"`
	IPList         string `db:"ip_list"           json:"ipList"`
	Mi             string `db:"mi"                json:"mi"`
	Region         string `db:"region"            json:"region"`
	LastIP         string `db:"last_ip"           json:"lastIp"`
	LastActiveTime int64  `db:"last_active_time" json:"lastActiveTime"`
	LastLoginTime  int64  `db:"last_login_time"   json:"lastLoginTime"`
}

// Recovered from runtime type 0x134aa80; original size 72 bytes.
type KamiGroup struct {
	ID     int64  `db:"id"     json:"id"`
	AppID  int64  `db:"appid"  json:"appid"`
	Name   string `db:"name"   json:"name"`
	Type   string `db:"type"   json:"type"`
	Value  int64  `db:"value"  json:"value"`
	Remark string `db:"remark" json:"remark"`
}

// Recovered from runtime type 0x13808e0; original size 184 bytes.
type Log struct {
	Summary    string `db:"-" json:"summary,omitempty"`
	ID         int64  `db:"id"          json:"id"`
	Type       string `db:"type"        json:"type"`
	Action     string `db:"action"      json:"action"`
	Data       string `db:"data"        json:"data"`
	Target     string `db:"target"      json:"target"`
	Operator   string `db:"operator"    json:"operator"`
	UID        int64  `db:"uid"         json:"uid"`
	IP         string `db:"ip"          json:"ip"`
	Region     string `db:"region"      json:"region"`
	Username   string `db:"username"    json:"username"`
	Nickname   string `db:"nickname"    json:"nickname"`
	Avatar     string `db:"avatar"      json:"avatar"`
	CreateTime int64  `db:"create_time" json:"createTime"`
}

// Recovered from runtime type 0x13661a0; original size 96 bytes.
type Notice struct {
	ID         int64  `db:"id"          json:"id"`
	AppID      int64  `db:"appid"       json:"appid"`
	Title      string `db:"title"       json:"title"`
	Content    string `db:"content"     json:"content"`
	Enabled    int    `db:"enabled"     json:"enabled"`
	AdminID    int64  `db:"admin_id"    json:"adminId"`
	Author     string `db:"author"      json:"author"`
	Views      int64  `db:"views"       json:"views"`
	UpdateTime int64  `db:"update_time" json:"updateTime"`
}

// Recovered from runtime type 0x130b640; original size 40 bytes.
type TrendBucket struct {
	Date   string `db:"d"      json:"date"`
	Action string `db:"action" json:"action"`
	N      int    `db:"n"      json:"n"`
}

// Recovered from runtime type 0x1396e80; original size 256 bytes.
type User struct {
	HasSignKey     bool   `db:"has_sign_key" json:"hasSignKey"`
	ID             int64  `db:"id"               json:"id"`
	AppID          int64  `db:"appid"            json:"appid"`
	GroupID        int64  `db:"group_id"         json:"groupId"`
	Username       string `db:"username"         json:"username"`
	Password       string `db:"password"         json:"-"`
	Nickname       string `db:"nickname"         json:"nickname"`
	Avatar         string `db:"avatar"           json:"avatar"`
	Vip            int64  `db:"vip"              json:"vip"`
	Mi             string `db:"mi"               json:"mi"`
	Region         string `db:"region"           json:"region"`
	RegisterIP     string `db:"register_ip"      json:"registerIp"`
	LastIP         string `db:"last_ip"          json:"lastIp"`
	LastRegion     string `db:"last_region"      json:"lastRegion"`
	LastActiveTime int64  `db:"last_active_time" json:"lastActiveTime"`
	LastLoginTime  int64  `db:"last_login_time"  json:"lastLoginTime"`
	BindLimit      int    `db:"bind_limit"       json:"bindLimit"`
	RegionLimit    int    `db:"region_limit"     json:"regionLimit"`
	IPList         string `db:"ip_list"          json:"ipList"`
	BanExpire      int64  `db:"ban_expire"       json:"banExpire"`
	CreateTime     int64  `db:"create_time"      json:"createTime"`
	GroupName      string `db:"group_name"  json:"groupName"`
	GroupLevel     int    `db:"group_level" json:"groupLevel"`
}

// Recovered from runtime type 0x13662e0; original size 112 bytes.
type UserDevice struct {
	ID       int64  `db:"id"        json:"id"`
	UID      int64  `db:"uid"       json:"-"`
	Mi       string `db:"mi"        json:"mi"`
	Device   string `db:"device"    json:"device"`
	OS       string `db:"os"        json:"os"`
	IP       string `db:"ip"        json:"ip"`
	Region   string `db:"region"    json:"region"`
	BindTime int64  `db:"bind_time" json:"bindTime"`
	LastTime int64  `db:"last_time" json:"lastTime"`
}

// Recovered from runtime type 0x1329a60; original size 48 bytes.
type UserExtend struct {
	ID    int64  `db:"id"      json:"id"`
	UID   int64  `db:"uid"     json:"-"`
	Key   string `db:"e_key"   json:"key"`
	Value string `db:"e_value" json:"value"`
}

// Recovered from runtime type 0x134ab80; original size 64 bytes.
type UserGroup struct {
	ID         int64  `db:"id"          json:"id"`
	AppID      int64  `db:"appid"       json:"appid"`
	Name       string `db:"name"        json:"name"`
	Level      int    `db:"level"       json:"level"`
	Remark     string `db:"remark"      json:"remark"`
	CreateTime int64  `db:"create_time" json:"createTime"`
}

// Recovered from runtime type 0x13605e0; original size 104 bytes.
type UserLog struct {
	ID         int64  `db:"id"          json:"id"`
	UID        int64  `db:"uid"         json:"-"`
	Type       string `db:"type"        json:"type"`
	Action     string `db:"action"      json:"action"`
	Data       string `db:"data"        json:"data"`
	IP         string `db:"ip"          json:"ip"`
	Region     string `db:"region"      json:"region"`
	CreateTime int64  `db:"create_time" json:"createTime"`
}

// Recovered from runtime type 0x1366420; original size 128 bytes.
type UserLogListItem struct {
	ID         int64  `db:"id"          json:"id"`
	Username   string `db:"username"    json:"username"`
	Nickname   string `db:"nickname"    json:"nickname"`
	Avatar     string `db:"avatar"      json:"avatar"`
	Action     string `db:"action"      json:"action"`
	Data       string `db:"data"        json:"data"`
	IP         string `db:"ip"          json:"ip"`
	Region     string `db:"region"      json:"region"`
	CreateTime int64  `db:"create_time" json:"createTime"`
}

// Recovered from runtime type 0x1357160; original size 88 bytes.
type UserLogWithUser struct {
	ID         int64  `db:"id"          json:"id"`
	UID        int64  `db:"uid"         json:"uid"`
	Username   string `db:"username"    json:"username"`
	Action     string `db:"action"      json:"action"`
	IP         string `db:"ip"          json:"ip"`
	Region     string `db:"region"      json:"region"`
	CreateTime int64  `db:"create_time" json:"createTime"`
}

// Recovered from runtime type 0x138d7e0; original size 200 bytes.
type Version struct {
	ID              int64  `db:"id"                 json:"id"`
	AppID           int64  `db:"appid"              json:"appid"`
	VersionName     string `db:"version_name"       json:"versionName"`
	VersionKey      string `db:"version_key"        json:"versionKey"`
	VersionNo       string `db:"version_no"         json:"versionNo"`
	Status          int    `db:"status"             json:"status"`
	URL             string `db:"url"                json:"url"`
	ApkName         string `db:"apk_name"           json:"apkName"`
	ApkSize         int64  `db:"apk_size"           json:"apkSize"`
	ApkURL          string `db:"apk_url"            json:"apkUrl"`
	IsCanonical     int    `db:"is_canonical"       json:"isCanonical"`
	CanonicalPath   string `db:"canonical_path"     json:"canonicalPath"`
	CloneDiffEnable int    `db:"clone_diff_enable"  json:"cloneDiffEnable"`
	OriginalPackage string `db:"original_package"   json:"originalPackage"`
	Changelog       string `db:"changelog"          json:"changelog"`
	CreateTime      int64  `db:"create_time"        json:"createTime"`
}

// Recovered from runtime type 0x1385f40; original size 152 bytes.
type Settings struct {
	ID               int64  `db:"id"                  json:"id"`
	WebURL           string `db:"web_url"             json:"webUrl"`
	UploadLimit      int    `db:"upload_limit"        json:"uploadLimit"`
	AdminTokenExpire int    `db:"admin_token_expire"  json:"adminTokenExpire"`
	AdminEntry       string `db:"admin_entry"         json:"adminEntry"`
	MysqlMaxConn     int    `db:"mysql_max_conn"      json:"mysqlMaxConn"`
	MysqlMaxIdle     int    `db:"mysql_max_idle"      json:"mysqlMaxIdle"`
	CacheEnabled     int    `db:"cache_enabled"       json:"cacheEnabled"`
	UserLogEnabled   int    `db:"user_log_enabled"    json:"userLogEnabled"`
	AdminLogEnabled  int    `db:"admin_log_enabled"   json:"adminLogEnabled"`
	LogLevel         string `db:"log_level"           json:"logLevel"`
	HTTPSEnabled     int    `db:"https_enabled"       json:"httpsEnabled"`
	HTTPSAddr        string `db:"https_addr"          json:"httpsAddr"`
	HTTPSCert        string `db:"https_cert"          json:"httpsCert"`
}

// Recovered from runtime type 0x12e8e40; original size 16 bytes.
type SettingsRepo struct {
	db    *sqlx.DB
	cache atomic.Pointer[Settings]
}

// Recovered from runtime type 0x12b9660; original size 8 bytes.
type AppRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9960; original size 8 bytes.
type IPInfoRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b94e0; original size 8 bytes.
type AdminRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b96e0; original size 8 bytes.
type BlockRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9be0; original size 8 bytes.
type UserRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9a60; original size 8 bytes.
type KamiRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9ae0; original size 8 bytes.
type LogRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9de0; original size 8 bytes.
type UserLogRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9760; original size 8 bytes.
type EncryptRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9860; original size 8 bytes.
type ExtendFileRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b97e0; original size 8 bytes.
type ExtendRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b98e0; original size 8 bytes.
type FunctionRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b99e0; original size 8 bytes.
type KamiGroupRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9c60; original size 8 bytes.
type UserDeviceRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9b60; original size 8 bytes.
type NoticeRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b95e0; original size 8 bytes.
type APIRouteRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9560; original size 8 bytes.
type APICodeRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9d60; original size 8 bytes.
type UserGroupRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9ce0; original size 8 bytes.
type UserExtendRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x12b9e60; original size 8 bytes.
type VersionRepo struct {
	db *sqlx.DB
}

// Recovered from runtime type 0x1371720; original size 128 bytes.
type IPInfo struct {
	ID         int64   `db:"id"          json:"id"`
	IP         string  `db:"ip"          json:"ip"`
	Provider   string  `db:"provider"    json:"provider"`
	Lat        float64 `db:"lat"         json:"lat"`
	Lng        float64 `db:"lng"         json:"lng"`
	Nation     string  `db:"nation"      json:"nation"`
	Province   string  `db:"province"    json:"province"`
	City       string  `db:"city"        json:"city"`
	District   string  `db:"district"    json:"district"`
	UpdateTime int64   `db:"update_time" json:"updateTime"`
}
