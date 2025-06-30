package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"ac/app/config"
	"ac/app/login"
	"ac/app/scheduler"

	"github.com/orzogc/acfundanmu"
)

func main() {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())
	
	cfg := config.DefaultConfig
	
	// 从环境变量读取配置
	if account := os.Getenv("ACFUN_ACCOUNT"); account != "" {
		cfg.Account = account
	}
	
	if password := os.Getenv("ACFUN_PASSWORD"); password != "" {
		cfg.Password = password
	}
	
	if intervalStr := os.Getenv("ACFUN_INTERVAL"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil {
			cfg.Interval = interval
		}
	}
	
	// 设置过滤列表路径，默认在config文件夹中
	filterListPath := os.Getenv("ACFUN_FILTER_LIST")
	if filterListPath == "" {
		filterListPath = "app/config/filter_list.txt"
	}
	
	// 确保账号和密码不为空
	if cfg.Account == "" || cfg.Password == "" {
		panic("AcFun账号或密码未设置，请设置ACFUN_ACCOUNT和ACFUN_PASSWORD环境变量")
	}

	fmt.Printf("启动AcFun直播列表监控程序，账号: %s, 更新间隔: %d秒, 过滤列表: %s\n", 
		cfg.Account, cfg.Interval, filterListPath)

	// 登录获取 Cookies
	cookies, err := login.Login(cfg.Account, cfg.Password)
	if err != nil {
		panic(err)
	}

	// 初始化 AcFunLive 实例
	ac, err := acfundanmu.NewAcFunLive(acfundanmu.SetCookies(cookies))
	if err != nil {
		panic(err)
	}

	// 启动直播监控
	monitor := scheduler.NewLiveMonitor(ac, time.Duration(cfg.Interval)*time.Second, filterListPath)
	go monitor.Start()

	// 阻塞主线程保持运行
	select {}
}
