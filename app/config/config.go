package config

// Config 应用配置结构体
type Config struct {
	Account  string // AcFun账号
	Password string // AcFun密码
	Interval int    // 刷新间隔（秒）
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	Account:  "13551433233",
	Password: "19941016wang",
	Interval: 300,
} 