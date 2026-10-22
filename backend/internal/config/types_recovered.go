// Reconstructed declarations from Go runtime metadata; method bodies are not recovered.
package config

// Recovered from runtime type 0x130a740; original size 40 bytes.
type CloneConfig struct {
	CanonicalDir string `yaml:"canonical_dir"`
	OutputDir    string `yaml:"output_dir"`
	DiffTTLSec   int    `yaml:"diff_ttl_sec"`
}

// Recovered from runtime type 0x1365ca0; original size 256 bytes.
type Config struct {
	Installed bool          `yaml:"installed"`
	Server    ServerConfig  `yaml:"server"`
	MySQL     MySQLConfig   `yaml:"mysql"`
	Redis     RedisConfig   `yaml:"redis"`
	Clone     CloneConfig   `yaml:"clone"`
	Upload    UploadConfig  `yaml:"upload"`
	Session   SessionConfig `yaml:"session"`
	Log       LogConfig     `yaml:"log"`
	Web       WebConfig     `yaml:"web"`
	Geo       GeoConfig     `yaml:"geo"`
}

type GeoConfig struct {
	LocalDir string `yaml:"local_dir"`
}

// Recovered from runtime type 0x12e83a0; original size 32 bytes.
type ServerConfig struct {
	Addr string `yaml:"addr"`
	Mode string `yaml:"mode"`
}

// Recovered from runtime type 0x133c5a0; original size 56 bytes.
type MySQLConfig struct {
	DSN          string `yaml:"dsn"`
	MaxOpen      int    `yaml:"max_open"`
	MaxIdle      int    `yaml:"max_idle"`
	ConnLifetime int    `yaml:"conn_lifetime"`
	TablePrefix  string `yaml:"table_prefix"`
}

// Recovered from runtime type 0x130a680; original size 40 bytes.
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	DB       int    `yaml:"db"`
	Password string `yaml:"password"`
}

// Recovered from runtime type 0x12e8440; original size 32 bytes.
type UploadConfig struct {
	Root      string `yaml:"root"`
	URLPrefix string `yaml:"url_prefix"`
}

// Recovered from runtime type 0x12e84e0; original size 16 bytes.
type SessionConfig struct {
	TTLSec         int `yaml:"ttl_sec"`
	AbsoluteTTLSec int `yaml:"absolute_ttl_sec"`
}

// Recovered from runtime type 0x12b86e0; original size 16 bytes.
type LogConfig struct {
	Level string `yaml:"level"`
}

// Recovered from runtime type 0x12b8760; original size 16 bytes.
type WebConfig struct {
	Dist string `yaml:"dist"`
}
