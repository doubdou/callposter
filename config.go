package callposter

//https://github.com/fiorix/go-eventsocket

import (
	"flag"
)

//命令行参数全局变量,运行时绑定
var (
	confDir = flag.String("conf", "../conf", "配置文件路径,默认为../conf")
	logDir  = flag.String("log", "../log", "日志文件路径,默认为../log")
	dataDir = flag.String("data", "../data", "数据文件路径,默认为../data")
)

//ConfigRoot 定义配置文件结构
type ConfigRoot struct {
	HTTP        HTTPConfig        `xml:"http"`
	Attribution AttributionConfig `xml:"attribution"`
}

//HTTPConfig 定义http配置数据
type HTTPConfig struct {
	IP   string `xml:"ip"`   //http服务监听地址
	Port string `xml:"port"` //http服务监听端口
}

//AttributionConfig 定义归属地配置数据
type AttributionConfig struct {
	Location   string `xml:"location"` //归属地城市名，如：苏州、南京
	HitPreifx  string `xml:"hit"`      //号码命中归属地城市，需加的前缀
	MissPrefix string `xml:"miss"`     //未命中归属地城市，需加的前缀
}

var config ConfigRoot

func httpConfigGet() *HTTPConfig {
	return &config.HTTP
}

func attributionConfigGet() *AttributionConfig {
	return &config.Attribution
}
