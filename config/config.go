package config

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Email struct {
	Subject  string `toml:"subject"`
	SMTP     string `toml:"smtp"`
	Port     int    `toml:"port"`
	From     string `toml:"from"`
	Password string `toml:"password"`
	To       string `toml:"to"`
}

type Config struct {
	URL      string     `toml:"url"`       // 监控链接
	Timeout  int64      `toml:"timeout"`   // 超时时间(s)
	CronTime string     `toml:"cron_time"` // 循环时间(分)
	Do       [][]string `toml:"do"`        // 需要执行的操作
	Email    Email      `toml:"email"`     //邮件
}

func NewConfig(path string) Config {
	config := Config{}
	// 解析配置文件
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Fatal("配置文件读取失败！", err)
	}
	return config
}
