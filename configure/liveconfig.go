package configure

import (
	"flag"

	"github.com/kr/pretty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ServerCfg is the configuration of server
type ServerCfg struct {
	Level         string `mapstructure:"level"`
	FLVDir        string `mapstructure:"flv_dir"`
	RTMPAddr      string `mapstructure:"rtmp_addr"`
	HTTPFLVAddr   string `mapstructure:"httpflv_addr"`
	ReadTimeout   int    `mapstructure:"read_timeout"`
	WriteTimeout  int    `mapstructure:"write_timeout"`
	GopNum        int    `mapstructure:"gop_num"`
	ServerName    string `mapstructure:"server_name"`
	ServerChannel string `mapstructure:"server_channel"`
	ServerKey     string `mapstructure:"server_key"`
}

// Config is the configuration of this livego
var Config = viper.New()

func initLog() {
	if l, err := log.ParseLevel(Config.GetString("level")); err == nil {
		log.SetLevel(l)
		log.SetReportCaller(l == log.DebugLevel)
	}
}

var (
	configFileFlag = flag.String("config", "livego.yaml", "config file")
)

func init() {
	// Set default values
	Config.SetDefault("level", "info")
	Config.SetDefault("flv_dir", "./tmp/app")
	Config.SetDefault("rtmp_addr", ":1935")
	Config.SetDefault("httpflv_addr", ":7001")
	Config.SetDefault("read_timeout", 10)
	Config.SetDefault("write_timeout", 10)
	Config.SetDefault("gop_num", 1)
	Config.SetDefault("hls_keep_after_end", false)
	Config.SetDefault("server.name", "live")
	Config.SetDefault("server.channel", "movie")
	Config.SetDefault("server.key", "123456")

	flag.Parse()

	Config.SetConfigFile(*configFileFlag)
	Config.AddConfigPath("/livego")
	Config.AddConfigPath("./")
	Config.ReadInConfig()
	// Log
	initLog()

	// Print final config
	c := ServerCfg{}
	Config.Unmarshal(&c)
	log.Debugf("Current configurations: \n%# v", pretty.Formatter(c))
}

// CheckAppName check the appname is still live
func CheckAppName(name string) bool {
	appName := Config.GetString("server.name")
	if appName == name {
		return true
	}
	return false
}

// GetChannel check if key is valid and returns channel name
func GetChannel(key string) string {
	appKey := Config.GetString("server.key")
	if appKey == key {
		return Config.GetString("server.channel")
	}
	return ""
}
