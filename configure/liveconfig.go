package configure

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/kr/pretty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

/*
{
  "server": [
    {
      "appname": "live",
      "live": true,
	  "hls": true,
	  "static_push": []
    }
  ]
}
*/

// Application is application, the basic unit of push and pull
type Application struct {
	Appname    string   `mapstructure:"appname"`
	Live       bool     `mapstructure:"live"`
	Hls        bool     `mapstructure:"hls"`
	StaticPush []string `mapstructure:"static_push"`
}

// Applications is a collection of Application
type Applications []Application

// JWT is used fro jwt authentication
type JWT struct {
	Secret    string `mapstructure:"secret"`
	Algorithm string `mapstructure:"algorithm"`
}

// ServerCfg is the configuration of server
type ServerCfg struct {
	Level           string       `mapstructure:"level"`
	ConfigFile      string       `mapstructure:"config_file"`
	FLVDir          string       `mapstructure:"flv_dir"`
	RTMPAddr        string       `mapstructure:"rtmp_addr"`
	HTTPFLVAddr     string       `mapstructure:"httpflv_addr"`
	HLSAddr         string       `mapstructure:"hls_addr"`
	HLSKeepAfterEnd bool         `mapstructure:"hls_keep_after_end"`
	APIAddr         string       `mapstructure:"api_addr"`
	RedisAddr       string       `mapstructure:"redis_addr"`
	RedisPwd        string       `mapstructure:"redis_pwd"`
	ReadTimeout     int          `mapstructure:"read_timeout"`
	WriteTimeout    int          `mapstructure:"write_timeout"`
	GopNum          int          `mapstructure:"gop_num"`
	JWT             JWT          `mapstructure:"jwt"`
	Server          Applications `mapstructure:"server"`
}

// defaultConfig is the default configuration
var defaultConf = ServerCfg{
	ConfigFile:      "livego.yaml",
	RTMPAddr:        ":1935",
	HTTPFLVAddr:     ":7001",
	HLSAddr:         ":7002",
	HLSKeepAfterEnd: false,
	APIAddr:         ":8090",
	WriteTimeout:    10,
	ReadTimeout:     10,
	GopNum:          1,
	Server: Applications{{
		Appname:    "live",
		Live:       true,
		Hls:        true,
		StaticPush: nil,
	}},
}

// Config is the configuration of this livego
var Config = viper.New()

func initLog() {
	if l, err := log.ParseLevel(Config.GetString("level")); err == nil {
		log.SetLevel(l)
		log.SetReportCaller(l == log.DebugLevel)
	}
}

func init() {
	// Default config
	b, _ := json.Marshal(defaultConf)
	defaultConfig := bytes.NewReader(b)
	viper.SetConfigType("json")
	viper.ReadConfig(defaultConfig)
	Config.MergeConfigMap(viper.AllSettings())

	// Flags
	pflag.String("rtmp_addr", ":1935", "RTMP server listen address")
	pflag.String("httpflv_addr", ":7001", "HTTP-FLV server listen address")
	pflag.String("hls_addr", ":7002", "HLS server listen address")
	pflag.String("api_addr", ":8090", "HTTP manage interface server listen address")
	pflag.String("config_file", "livego.yaml", "configure filename")
	pflag.String("level", "info", "Log level")
	pflag.Bool("hls_keep_after_end", false, "Maintains the HLS after the stream ends")
	pflag.String("flv_dir", "tmp", "output flv file at flvDir/APP/KEY_TIME.flv")
	pflag.Int("read_timeout", 10, "read time out")
	pflag.Int("write_timeout", 10, "write time out")
	pflag.Int("gop_num", 1, "gop num")
	pflag.Parse()
	Config.BindPFlags(pflag.CommandLine)

	// File
	Config.SetConfigFile(Config.GetString("config_file"))
	Config.AddConfigPath(".")
	err := Config.ReadInConfig()
	if err != nil {
		log.Warning(err)
		log.Info("Using default config")
	} else {
		Config.MergeInConfig()
	}

	// Environment
	replacer := strings.NewReplacer(".", "_")
	Config.SetEnvKeyReplacer(replacer)
	Config.AllowEmptyEnv(true)
	Config.AutomaticEnv()

	// Log
	initLog()

	// Print final config
	c := ServerCfg{}
	Config.Unmarshal(&c)
	log.Debugf("Current configurations: \n%# v", pretty.Formatter(c))
}

// CheckAppName check the appname is still live
func CheckAppName(appname string) bool {
	apps := Applications{}
	Config.UnmarshalKey("server", &apps)
	for _, app := range apps {
		if app.Appname == appname {
			return app.Live
		}
	}
	return false
}

// GetStaticPushURLList get static push url list from config
func GetStaticPushURLList(appname string) ([]string, bool) {
	apps := Applications{}
	Config.UnmarshalKey("server", &apps)
	for _, app := range apps {
		if (app.Appname == appname) && app.Live {
			if len(app.StaticPush) > 0 {
				return app.StaticPush, true
			}
			return nil, false
		}
	}
	return nil, false
}
