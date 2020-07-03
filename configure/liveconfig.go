package configure

import (
	"flag"

	"github.com/kr/pretty"
	log "github.com/sirupsen/logrus"
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
	Appname string `mapstructure:"appname"`
	Channel string `mapstructure:"channel"`
	Key     string `mapstructure:"key"`
}

// Applications is a collection of Application
// type Applications []Application

// JWT is used fro jwt authentication
// type JWT struct {
// 	Secret    string `mapstructure:"secret"`
// 	Algorithm string `mapstructure:"algorithm"`
// }

// ServerCfg is the configuration of server
type ServerCfg struct {
	Level        string      `mapstructure:"level"`
	FLVDir       string      `mapstructure:"flv_dir"`
	RTMPAddr     string      `mapstructure:"rtmp_addr"`
	HTTPFLVAddr  string      `mapstructure:"httpflv_addr"`
	ReadTimeout  int         `mapstructure:"read_timeout"`
	WriteTimeout int         `mapstructure:"write_timeout"`
	GopNum       int         `mapstructure:"gop_num"`
	Server       Application `mapstructure:"server"`
}

// // defaultConfig is the default configuration
// var defaultConf = ServerCfg{
// 	Level:        "info",
// 	FLVDir:       "/tmp/app",
// 	RTMPAddr:     ":1935",
// 	HTTPFLVAddr:  ":7001",
// 	WriteTimeout: 10,
// 	ReadTimeout:  10,
// 	GopNum:       1,
// 	App: Application{
// 		Appname: "live",
// 		Channel: "movie",
// 		Key:     "123456",
// 	},
// }

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
	// Default config
	// b, _ := json.Marshal(defaultConf)
	// defaultConfig := bytes.NewReader(b)
	// Config.ReadConfig(defaultConfig)
	// log.Info(Config.GetString("config_file"))

	// // log.Info(Config.GetString("config_file"))
	// // File
	// Config.SetConfigFile("livego.yaml")
	// Config.AddConfigPath("/livego")
	// Config.AddConfigPath(".")
	// err := Config.ReadInConfig()
	// if err != nil {
	// 	log.Warning(err)
	// 	log.Info("Using default config")
	// } else {
	// 	Config.MergeInConfig()
	// }
	// log.Info(Config.GetString("config_file"))

	// // Environment
	// replacer := strings.NewReplacer(".", "_")
	// Config.SetEnvKeyReplacer(replacer)
	// Config.AllowEmptyEnv(true)
	// Config.AutomaticEnv()

	// Set default values
	Config.SetDefault("level", "info")
	Config.SetDefault("flv_dir", "/tmp/app")
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
func CheckAppName(appname string) bool {
	app := Application{}
	Config.UnmarshalKey("server", &app)
	if app.Appname == appname {
		return true
	}
	return false
}

// GetChannel check if key is valid and returns channel name
func GetChannel(key string) string {
	app := Application{}
	Config.UnmarshalKey("server", &app)
	if app.Key == key {
		return app.Channel
	}
	return ""
}
