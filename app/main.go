package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"bn/app/config"
	"bn/app/global"
	"bn/app/login"
	"bn/app/monitor"
	"bn/app/scheduler"

	"github.com/orzogc/acfundanmu"
)

const (
	// 刷新直播弹幕列表的时间间隔
	danmuRefreshInterval = 2 * time.Minute
	// 自动点赞的默认间隔（秒）
	defaultLikeInterval = 30
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

	// 读取自动点赞配置
	autoLike := true // 默认启用自动点赞
	if autoLikeStr := os.Getenv("ACFUN_AUTO_LIKE"); autoLikeStr != "" {
		if autoLikeVal, err := strconv.ParseBool(autoLikeStr); err == nil {
			autoLike = autoLikeVal
		}
	}
	
	likeInterval := defaultLikeInterval
	if likeIntervalStr := os.Getenv("ACFUN_LIKE_INTERVAL"); likeIntervalStr != "" {
		if interval, err := strconv.Atoi(likeIntervalStr); err == nil && interval > 0 {
			likeInterval = interval
		}
	}

	fmt.Printf("启动AcFun直播弹幕监控程序，账号: %s, 更新间隔: %d秒, 过滤列表: %s, 自动点赞: %v, 点赞间隔: %d秒\n", 
		cfg.Account, cfg.Interval, filterListPath, autoLike, likeInterval)

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

	// 获取全局存储实例
	globalStore := global.GetStore()
	
	// 打印acfun.midground.api_st (临时，用于调试)
	log.Printf("主客户端的acfun.midground.api_st: %s", globalStore.GetAPIToken())

	// 创建过滤列表
	filterList := scheduler.NewFilterList(filterListPath)

	// 使用更新后的刷新间隔（从配置获取）
	refreshInterval := time.Duration(cfg.Interval) * time.Second
	
	// 启动全平台弹幕监控
	danmuMonitor := monitor.NewDanmuMonitor(ac, filterList, refreshInterval, autoLike, likeInterval)
	err = danmuMonitor.Start()
	if err != nil {
		log.Fatalf("启动弹幕监控失败: %v", err)
	}

	// 定期打印状态信息
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			count := danmuMonitor.GetActiveListenerCount()
			fmt.Printf("\n当前正在监听 %d 个直播间的弹幕\n", count)
		}
	}()

	// 捕获Ctrl+C信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// 等待信号
	<-sigCh
	fmt.Println("\n接收到停止信号，正在关闭监控...")
	
	// 停止监控
	danmuMonitor.Stop()
	
	fmt.Println("监控已停止")
}
