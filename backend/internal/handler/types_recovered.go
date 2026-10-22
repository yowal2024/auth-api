// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package handler

import (
	"auth-api/internal/auth"
	"auth-api/internal/buildpool"
	"auth-api/internal/geo"
	"auth-api/internal/repo"
	"auth-api/internal/tlsmgr"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Recovered from runtime type 0x12e8a80; original size 16 bytes.
type AdminAdminHandler struct {
	admins *repo.AdminRepo
	store  *auth.SessionStore
}

// Recovered from runtime type 0x134a080; original size 80 bytes.
type AdminAppHandler struct {
	LocalGeo  *geo.LocalDB
	apps      *repo.AppRepo
	rdb       *redis.Client
	mgr       *tlsmgr.Manager
	masterKey []uint8
	logoDir   string
	logoURL   string
}

// Recovered from runtime type 0x130b1c0; original size 24 bytes.
type AdminAuthHandler struct {
	admins  *repo.AdminRepo
	store   *auth.SessionStore
	captcha *auth.CaptchaService
}

// Recovered from runtime type 0x12b8fe0; original size 8 bytes.
type AdminBlockHandler struct {
	blocks *repo.BlockRepo
}

// Recovered from runtime type 0x134a180; original size 48 bytes.
type AdminDashboardHandler struct {
	users *repo.UserRepo
	apps  *repo.AppRepo
	kamis *repo.KamiRepo
	logs  *repo.LogRepo
	uLogs *repo.UserLogRepo
	rdb   *redis.Client
}

// Recovered from runtime type 0x12b9060; original size 8 bytes.
type AdminEncryptHandler struct {
	encrypts *repo.EncryptRepo
}

// Recovered from runtime type 0x130b280; original size 40 bytes.
type AdminExtendFileHandler struct {
	files     *repo.ExtendFileRepo
	uploadDir string
	uploadURL string
}

// Recovered from runtime type 0x12b9160; original size 8 bytes.
type AdminExtendHandler struct {
	extends *repo.ExtendRepo
}

// Recovered from runtime type 0x130b340; original size 24 bytes.
type AdminFunctionHandler struct {
	fns  *repo.FunctionRepo
	apps *repo.AppRepo
	db   *sqlx.DB
}

// Recovered from runtime type 0x133c840; original size 40 bytes.
type AdminKamiHandler struct {
	groups  *repo.KamiGroupRepo
	kamis   *repo.KamiRepo
	devices *repo.UserDeviceRepo
	logs    *repo.UserLogRepo
	rdb     *redis.Client
}

// Recovered from runtime type 0x12e8b20; original size 16 bytes.
type AdminLogHandler struct {
	logs     *repo.LogRepo
	userLogs *repo.UserLogRepo
}

// Recovered from runtime type 0x13298e0; original size 48 bytes.
type AdminMeHandler struct {
	admins    *repo.AdminRepo
	store     *auth.SessionStore
	uploadDir string
	uploadURL string
}

// Recovered from runtime type 0x12b92e0; original size 8 bytes.
type AdminNoticeHandler struct {
	notices *repo.NoticeRepo
}

// Recovered from runtime type 0x134a280; original size 64 bytes.
type AdminSystemHandler struct {
	settings           *repo.SettingsRepo
	routes             *repo.APIRouteRepo
	codes              *repo.APICodeRepo
	masterKey          []uint8
	OnAdminEntryChange func() error
	OnHTTPSChange      func() error
}

// Recovered from runtime type 0x12b93e0; original size 8 bytes.
type AdminUserGroupHandler struct {
	groups *repo.UserGroupRepo
}

// Recovered from runtime type 0x134a380; original size 48 bytes.
type AdminUserHandler struct {
	users   *repo.UserRepo
	groups  *repo.UserGroupRepo
	extends *repo.UserExtendRepo
	devices *repo.UserDeviceRepo
	logs    *repo.UserLogRepo
	rdb     *redis.Client
}

// Recovered from runtime type 0x13299a0; original size 48 bytes.
type AdminVersionHandler struct {
	versions     *repo.VersionRepo
	canonicalDir string
	canonicalURL string
	diffStore    *buildpool.DiffStore
}

// Recovered from runtime type 0x12b9460; original size 8 bytes.
type SDKExtendFileHandler struct {
	files *repo.ExtendFileRepo
}

// Recovered from runtime type 0x12e8c60; original size 16 bytes.
type SDKFunctionHandler struct {
	fns *repo.FunctionRepo
	db  *sqlx.DB
}

// Recovered from runtime type 0x12e8d00; original size 16 bytes.
type SDKHandshakeHandler struct {
	sessStore   *auth.SDKSessionStore
	encryptRepo *repo.EncryptRepo
}

// Recovered from runtime type 0x133c920; original size 40 bytes.
type SDKIniHandler struct {
	notices   *repo.NoticeRepo
	extends   *repo.ExtendRepo
	versions  *repo.VersionRepo
	apps      *repo.AppRepo
	diffStore *buildpool.DiffStore
}

// Recovered from runtime type 0x1356f20; original size 56 bytes.
type SDKKamiHandler struct {
	kamis      *repo.KamiRepo
	kamiGroups *repo.KamiGroupRepo
	users      *repo.UserRepo
	userLogs   *repo.UserLogRepo
	settings   *repo.SettingsRepo
	fns        *repo.FunctionRepo
	db         *sqlx.DB
}

// Recovered from runtime type 0x138d600; original size 144 bytes.
type SDKUserHandler struct {
	users       *repo.UserRepo
	userDevices *repo.UserDeviceRepo
	blocks      *repo.BlockRepo
	userLogs    *repo.UserLogRepo
	sessions    *auth.UserSessionStore
	rdb         *redis.Client
	geo         *geo.Checker
	settings    *repo.SettingsRepo
	brute       *auth.BruteForce
	kamis       *repo.KamiRepo
	kamiGroups  *repo.KamiGroupRepo
	userExtends *repo.UserExtendRepo
	fns         *repo.FunctionRepo
	db          *sqlx.DB
	avatarDir   string
	uploadURL   string
}

// Recovered from runtime type 0x133ca00; original size 40 bytes.
type SDKVersionHandler struct {
	versions  *repo.VersionRepo
	users     *repo.UserRepo
	diffStore *buildpool.DiffStore
	userLogs  *repo.UserLogRepo
	settings  *repo.SettingsRepo
}

// Recovered from runtime type 0x13290a0; original size 56 bytes.
type adminReq struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// Recovered from runtime type 0x1391340; original size 160 bytes.
type appReq struct {
	InheritFrom   int64   `json:"inheritFrom"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Mode          string  `json:"mode"`
	Logo          string  `json:"logo"`
	Pkg           string  `json:"pkg"`
	TokenKey      string  `json:"tokenKey"`
	BindDevice    *bool   `json:"bindDevice"`
	BindRegion    *bool   `json:"bindRegion"`
	DiffUpdate    *bool   `json:"diffUpdate"`
	RandomPackage *bool   `json:"randomPackage"`
	Encryption    *int64  `json:"encryption"`
	UserTokenTTL  *int    `json:"userTokenTtl"`
	GeoProvider   *string `json:"geoProvider"`
	GeoKey        *string `json:"geoKey"`
	IPDistanceMax *int    `json:"ipDistanceMax"`
	BruteForce    *bool   `json:"bruteForceEnabled"`
	MtlsEnabled   *bool   `json:"mtlsEnabled"`
	Status        *int    `json:"status"`
}

// Recovered from runtime type 0x12b8960; original size 24 bytes.
type batchDeleteReq struct {
	IDs []int64 `json:"ids"`
}

// Recovered from runtime type 0x130ae00; original size 48 bytes.
type blockAddReq struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Reason string `json:"reason"`
}

// Recovered from runtime type 0x12e86c0; original size 16 bytes.
type codeOverrideReq struct {
	OverrideCode int   `json:"overrideCode"`
	Rewritten    *bool `json:"rewritten"`
}

// Recovered from runtime type 0x130a8c0; original size 24 bytes.
type codeSaveItem struct {
	ID           int64 `json:"id"`
	OverrideCode int   `json:"overrideCode"`
	Rewritten    bool  `json:"rewritten"`
}

// Recovered from runtime type 0x12b89e0; original size 24 bytes.
type codeSaveReq struct {
	Items []codeSaveItem `json:"items"`
}

// Recovered from runtime type 0x12b8a60; original size 1 bytes.
type codeToggleReq struct {
	Rewritten bool `json:"rewritten"`
}

// Recovered from runtime type 0x130a980; original size 48 bytes.
type dashboardRecent struct {
	Text string `json:"text"`
	Time string `json:"time"`
	Type string `json:"type"`
}

// Recovered from runtime type 0x1371300; original size 160 bytes.
type dashboardResp struct {
	Users           int               `json:"users"`
	UsersToday      int               `json:"usersToday"`
	QueryToday      int64             `json:"queryToday"`
	APITotal        int64             `json:"apiTotal"`
	KamiAvail       int               `json:"kamiAvail"`
	KamiUsedToday   int               `json:"kamiUsedToday"`
	OnlineNow       int               `json:"onlineNow"`
	OnlinePeakToday int               `json:"onlinePeakToday"`
	Trend           dashboardTrend    `json:"trend"`
	Recent          []dashboardRecent `json:"recent"`
}

// Recovered from runtime type 0x130aa40; original size 72 bytes.
type dashboardTrend struct {
	Dates  []string `json:"dates"`
	Logins []int    `json:"logins"`
	Regs   []int    `json:"regs"`
}

// Recovered from runtime type 0x1376f40; original size 160 bytes.
type encryptReq struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Algorithm        string `json:"algorithm"`
	Encoding         string `json:"encoding"`
	SymKey           string `json:"symKey"`
	TimeSkew         int    `json:"timeSkew"`
	SignMode         string `json:"signMode"`
	ClientPrivateKey string `json:"clientPrivateKey"`
	ClientPublicKey  string `json:"clientPublicKey"`
	ServerPrivateKey string `json:"serverPrivateKey"`
	ServerPublicKey  string `json:"serverPublicKey"`
}

// Recovered from runtime type 0x130ab00; original size 40 bytes.
type extendItem struct {
	ID    int64  `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Recovered from runtime type 0x1329160; original size 56 bytes.
type extendSaveReq struct {
	ID     int64  `json:"id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	Remark string `json:"remark"`
}

// Recovered from runtime type 0x1329220; original size 64 bytes.
type extendsSaveReq struct {
	List  []extendItem `json:"list"`
	ID    int64        `json:"id"`
	Key   string       `json:"key"`
	Value string       `json:"value"`
}

// Recovered from runtime type 0x1356bc0; original size 96 bytes.
type fileInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Filesize int64  `json:"filesize"`
	Filetype string `json:"filetype"`
	Category string `json:"category"`
	Version  string `json:"version"`
	Remark   string `json:"remark"`
}

// Recovered from runtime type 0x1356ce0; original size 96 bytes.
type fnReq struct {
	ID      int64       `json:"id"`
	Name    string      `json:"name"`
	Trigger string      `json:"trigger"`
	Code    string      `json:"code"`
	Remark  string      `json:"remark"`
	Timeout int         `json:"timeout"`
	Enabled interface{} `json:"enabled"`
}

// Recovered from runtime type 0x130aec0; original size 40 bytes.
type fnRunReq struct {
	Code    string      `json:"code"`
	Timeout int         `json:"timeout"`
	Payload interface{} `json:"payload"`
}

// Recovered from runtime type 0x12b8ae0; original size 16 bytes.
type handshakeRequest struct {
	Key string `json:"key"`
}

// Recovered from runtime type 0x12b8b60; original size 16 bytes.
type httpsGenReq struct {
	Host string `json:"host"`
}

// Recovered from runtime type 0x13292e0; original size 40 bytes.
type httpsSaveReq struct {
	Enabled *bool   `json:"enabled"`
	Addr    string  `json:"addr"`
	Cert    *string `json:"cert"`
	Key     *string `json:"key"`
}

// Recovered from runtime type 0x133c680; original size 56 bytes.
type kamiGenerateReq struct {
	GroupID int64  `json:"groupId"`
	Count   int    `json:"count"`
	Prefix  string `json:"prefix"`
	Length  int    `json:"length"`
	Remark  string `json:"remark"`
}

// Recovered from runtime type 0x133c760; original size 64 bytes.
type kamiGroupReq struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Value  int64  `json:"value"`
	Remark string `json:"remark"`
}

// Recovered from runtime type 0x12e8580; original size 16 bytes.
type kamiListItem struct {
	*repo.Kami
	Online bool `json:"online"`
}

// Recovered from runtime type 0x12b8be0; original size 8 bytes.
type kamiRewardReq struct {
	Seconds int64 `json:"seconds"`
}

// Recovered from runtime type 0x1356e00; original size 56 bytes.
type kamiUserUpdateReq struct {
	Frozen      *bool   `json:"frozen"`
	Nickname    *string `json:"nickname"`
	Vip         *int64  `json:"vip"`
	BanExpire   *int64  `json:"banExpire"`
	BindLimit   *int    `json:"bindLimit"`
	RegionLimit *int    `json:"regionLimit"`
	Mi          *string `json:"mi"`
	Remark      *string `json:"remark"`
}

// Recovered from runtime type 0x13293a0; original size 64 bytes.
type logClearReq struct {
	IDs      []int64 `json:"ids"`
	Category string  `json:"category"`
	Type     string  `json:"type"`
	Before   int64   `json:"before"`
}

// Recovered from runtime type 0x1329460; original size 64 bytes.
type loginReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaId   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

// Recovered from runtime type 0x12e8620; original size 24 bytes.
type loginResp struct {
	Token   string                 `json:"token"`
	Profile map[string]interface{} `json:"profile"`
}

// Recovered from runtime type 0x1329520; original size 48 bytes.
type noticeSaveReq struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Enabled *int   `json:"enabled"`
}

// Recovered from runtime type 0x12e8760; original size 32 bytes.
type passwordReq struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// Recovered from runtime type 0x130af80; original size 48 bytes.
type profileReq struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// Recovered from runtime type 0x130abc0; original size 40 bytes.
type recentItem struct {
	text string
	ts   int64
	kind string
}

// Recovered from runtime type 0x12b8c60; original size 16 bytes.
type resetPwdReq struct {
	Password string `json:"password"`
}

// Recovered from runtime type 0x13295e0; original size 64 bytes.
type rewardReq struct {
	Value   int64   `json:"value"`
	Unit    string  `json:"unit"`
	Target  string  `json:"target"`
	UserIDs []int64 `json:"userIds"`
}

// Recovered from runtime type 0x12e8800; original size 24 bytes.
type routeOverrideReq struct {
	OverridePath string `json:"overridePath"`
	Rewritten    *bool  `json:"rewritten"`
}

// Recovered from runtime type 0x130ac80; original size 32 bytes.
type routeSaveItem struct {
	ID           int64  `json:"id"`
	OverridePath string `json:"overridePath"`
	Rewritten    bool   `json:"rewritten"`
}

// Recovered from runtime type 0x12b8ce0; original size 24 bytes.
type routeSaveReq struct {
	Items []routeSaveItem `json:"items"`
}

// Recovered from runtime type 0x12b8d60; original size 1 bytes.
type routeToggleReq struct {
	Rewritten bool `json:"rewritten"`
}

// Recovered from runtime type 0x12e88a0; original size 32 bytes.
type sdkCloudFunctionReq struct {
	Name  string      `json:"name"`
	Param interface{} `json:"param"`
}

// Recovered from runtime type 0x12e8940; original size 32 bytes.
type sdkIniReq struct {
	VersionKey string `json:"versionKey"`
	VersionNo  string `json:"versionNo"`
}

// Recovered from runtime type 0x130b040; original size 48 bytes.
type sdkKamiLogonReq struct {
	Card     string `json:"card"`
	Username string `json:"username"`
	Mi       string `json:"mi"`
}

// Recovered from runtime type 0x12b8de0; original size 16 bytes.
type sdkKamiReq struct {
	Card string `json:"card"`
}

// Recovered from runtime type 0x130b100; original size 48 bytes.
type sdkLogonReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Mi       string `json:"mi"`
}

// Recovered from runtime type 0x13296a0; original size 64 bytes.
type sdkRegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Mi       string `json:"mi"`
}

// Recovered from runtime type 0x12e89e0; original size 16 bytes.
type sdkUpdateInfoReq struct {
	Nickname *string `json:"nickname"`
	Avatar   *string `json:"avatar"`
}

// Recovered from runtime type 0x12b8e60; original size 8 bytes.
type sdkUserExtendsReq struct {
	Set map[string]string `json:"set"`
}

// Recovered from runtime type 0x1371460; original size 104 bytes.
type settingsReq struct {
	WebURL           string `json:"webUrl"`
	UploadLimit      int    `json:"uploadLimit"`
	AdminTokenExpire int    `json:"adminTokenExpire"`
	AdminEntry       string `json:"adminEntry"`
	MysqlMaxConn     int    `json:"mysqlMaxConn"`
	MysqlMaxIdle     int    `json:"mysqlMaxIdle"`
	CacheEnabled     *bool  `json:"cacheEnabled"`
	UserLogEnabled   *bool  `json:"userLogEnabled"`
	AdminLogEnabled  *bool  `json:"adminLogEnabled"`
	LogLevel         string `json:"logLevel"`
}

// Recovered from runtime type 0x12e8bc0; original size 24 bytes.
type triggerReject struct {
	Code int
	Msg  string
}

// Recovered from runtime type 0x1329760; original size 56 bytes.
type userCreateReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	GroupID  int64  `json:"groupId"`
}

// Recovered from runtime type 0x1329820; original size 48 bytes.
type userGroupReq struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Level  int    `json:"level"`
	Remark string `json:"remark"`
}

// Recovered from runtime type 0x130ad40; original size 24 bytes.
type userListItem struct {
	*repo.User
	Online bool `json:"online"`
	Status int  `json:"status"`
}

// Recovered from runtime type 0x13715c0; original size 80 bytes.
type userUpdateReq struct {
	Vip         *int64  `json:"vip"`
	BindLimit   *int    `json:"bindLimit"`
	RegionLimit *int    `json:"regionLimit"`
	ClearIPList *bool   `json:"clearIpList"`
	IPList      *string `json:"ipList"`
	BanExpire   *int64  `json:"banExpire"`
	Password    *string `json:"password"`
	Nickname    *string `json:"nickname"`
	Avatar      *string `json:"avatar"`
	GroupID     *int64  `json:"groupId"`
}

// Recovered from runtime type 0x1385bc0; original size 184 bytes.
type versionSaveReq struct {
	ID              int64  `json:"id"`
	VersionName     string `json:"versionName"`
	VersionKey      string `json:"versionKey"`
	VersionNo       string `json:"versionNo"`
	Status          *int   `json:"status"`
	URL             string `json:"url"`
	ApkName         string `json:"apkName"`
	ApkSize         int64  `json:"apkSize"`
	ApkURL          string `json:"apkUrl"`
	IsCanonical     int    `json:"isCanonical"`
	CanonicalPath   string `json:"canonicalPath"`
	CloneDiffEnable int    `json:"cloneDiffEnable"`
	OriginalPackage string `json:"originalPackage"`
	Changelog       string `json:"changelog"`
}
