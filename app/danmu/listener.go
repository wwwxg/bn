package danmu

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/orzogc/acfundanmu"
)

// CommentHandler 是处理弹幕的函数类型
type CommentHandler func(comment *acfundanmu.Comment)

// DanmuListener 弹幕监听器
type DanmuListener struct {
	client        *acfundanmu.AcFunLive
	liveID        string
	handlers      []CommentHandler
	isRunning     bool
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewDanmuListener 创建新的弹幕监听器
func NewDanmuListener(client *acfundanmu.AcFunLive) *DanmuListener {
	ctx, cancel := context.WithCancel(context.Background())
	return &DanmuListener{
		client:    client,
		handlers:  make([]CommentHandler, 0),
		isRunning: false,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// RegisterHandler 注册弹幕处理函数
func (l *DanmuListener) RegisterHandler(handler CommentHandler) {
	l.handlers = append(l.handlers, handler)
}

// handleComment 处理弹幕消息
func (l *DanmuListener) handleComment(_ *acfundanmu.AcFunLive, comment *acfundanmu.Comment) {
	for _, handler := range l.handlers {
		handler(comment)
	}
}

// Start 开始监听弹幕
func (l *DanmuListener) Start() error {
	if l.isRunning {
		return fmt.Errorf("弹幕监听器已经在运行")
	}

	// 注册弹幕处理函数
	l.client.OnComment(l.handleComment)

	// 开始获取弹幕，使用事件模式
	errCh := l.client.StartDanmu(l.ctx, true)

	// 记录LiveID
	l.liveID = l.client.GetLiveID()
	log.Printf("开始监听直播间 %s 的弹幕", l.liveID)
	l.isRunning = true

	// 处理错误或结束信号
	go func() {
		select {
		case err := <-errCh:
			if err != nil {
				log.Printf("弹幕监听发生错误: %v", err)
			} else {
				log.Printf("弹幕监听正常结束")
			}
			l.isRunning = false
		case <-l.ctx.Done():
			log.Printf("弹幕监听被手动停止")
			l.isRunning = false
		}
	}()

	return nil
}

// Stop 停止监听弹幕
func (l *DanmuListener) Stop() {
	if l.isRunning {
		l.cancel()
		l.isRunning = false
	}
}

// IsRunning 判断监听器是否正在运行
func (l *DanmuListener) IsRunning() bool {
	return l.isRunning
}

// StartInteractive 开始交互式监听，可以通过Ctrl+C停止
func (l *DanmuListener) StartInteractive() error {
	err := l.Start()
	if err != nil {
		return err
	}

	// 捕获Ctrl+C信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigCh
	log.Println("接收到停止信号，正在关闭弹幕监听...")
	l.Stop()
	
	return nil
}

// DefaultCommentPrinter 默认的弹幕打印函数
func DefaultCommentPrinter(comment *acfundanmu.Comment) {
	timestamp := time.Unix(comment.SendTime/1000, 0).Format("15:04:05")
	fmt.Printf("[%s] %s: %s\n", timestamp, comment.Nickname, comment.Content)
} 