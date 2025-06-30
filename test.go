package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"bn/app/config"
	"bn/app/login"

	"github.com/orzogc/acfundanmu"
)

func main() {
	// 指定要监听的直播间UID
	const uid int64 = 36115445

	// 获取配置
	cfg := config.DefaultConfig
	
	fmt.Printf("正在连接主播UID: %d 的直播间\n", uid)
	
	// 登录获取 Cookies
	cookies, err := login.Login(cfg.Account, cfg.Password)
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}

	// 初始化 AcFunLive 实例
	acLive, err := acfundanmu.NewAcFunLive(acfundanmu.SetCookies(cookies))
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	// 设置要监听的主播
	acLive, err = acLive.SetLiverUID(uid)
	if err != nil {
		log.Fatalf("设置主播UID失败: %v", err)
	}

	// 设置弹幕处理函数
	acLive.OnComment(func(_ *acfundanmu.AcFunLive, comment *acfundanmu.Comment) {
		timestamp := time.Unix(comment.SendTime/1000, 0).Format("15:04:05")
		fmt.Printf("[%s] %s: %s\n", timestamp, comment.Nickname, comment.Content)
	})

	// 创建上下文用于停止弹幕监听
	fmt.Println("开始监听弹幕，按Ctrl+C停止...")
	
	// 开始获取弹幕，使用事件模式
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	
	errCh := acLive.StartDanmu(ctx, true)
	
	// 等待错误或者结束信号
	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("弹幕监听发生错误: %v", err)
		}
	case <-ctx.Done():
		log.Println("接收到停止信号，正在关闭弹幕监听...")
	}
	
	fmt.Println("弹幕监听已停止")
} 