package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"bn/app/action"
	"bn/app/global"
	"bn/app/scheduler"

	"github.com/orzogc/acfundanmu"
)

// DanmuMonitor 弹幕监控器
type DanmuMonitor struct {
	client       *acfundanmu.AcFunLive
	filterList   *scheduler.FilterList
	refreshInterval time.Duration
	listeners    map[int64]*LiveListener
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	isRunning    bool
	autoLike     bool           // 是否自动点赞
	likeInterval int            // 点赞间隔（秒）
}

// LiveListener 单个直播间的监听器
type LiveListener struct {
	client      *acfundanmu.AcFunLive
	userID      int64
	nickname    string
	title       string
	liveID      string
	ctx         context.Context
	cancel      context.CancelFunc
	lastComment time.Time
}

// NewDanmuMonitor 创建新的弹幕监控器
func NewDanmuMonitor(client *acfundanmu.AcFunLive, filterList *scheduler.FilterList, refreshInterval time.Duration, autoLike bool, likeInterval int) *DanmuMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 初始化全局存储
	globalStore := global.GetStore()
	globalStore.SetMainClient(client)
	
	return &DanmuMonitor{
		client:         client,
		filterList:     filterList,
		refreshInterval: refreshInterval,
		listeners:      make(map[int64]*LiveListener),
		ctx:            ctx,
		cancel:         cancel,
		isRunning:      false,
		autoLike:       autoLike,
		likeInterval:   likeInterval,
	}
}

// Start 开始监听所有直播间弹幕
func (m *DanmuMonitor) Start() error {
	if m.isRunning {
		return fmt.Errorf("弹幕监控器已经在运行")
	}
	
	m.isRunning = true
	
	// 启动一个协程定期刷新直播列表
	go m.refreshLoop()
	
	log.Println("开始监控所有直播间弹幕")
	
	return nil
}

// Stop 停止监听
func (m *DanmuMonitor) Stop() {
	if !m.isRunning {
		return
	}
	
	m.cancel()
	m.isRunning = false
	
	// 停止所有监听器
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, listener := range m.listeners {
		listener.cancel()
	}
	m.listeners = make(map[int64]*LiveListener)
	
	log.Println("已停止监控所有直播间弹幕")
}

// refreshLoop 定期刷新直播列表
func (m *DanmuMonitor) refreshLoop() {
	ticker := time.NewTicker(m.refreshInterval)
	defer ticker.Stop()
	
	// 立即获取一次
	m.refreshLiveList()
	
	for {
		select {
		case <-ticker.C:
			m.refreshLiveList()
		case <-m.ctx.Done():
			return
		}
	}
}

// refreshLiveList 刷新直播列表
func (m *DanmuMonitor) refreshLiveList() {
	log.Println("开始刷新直播列表...")
	
	// 获取所有直播列表
	liveList, err := m.client.GetAllLiveList()
	if err != nil {
		log.Printf("获取直播列表失败: %v", err)
		return
	}
	
	log.Printf("获取到 %d 个直播间", len(liveList))
	
	// 记录当前活跃的直播间ID
	activeIDs := make(map[int64]bool)
	globalStore := global.GetStore()
	streamManager := global.GetStreamManager()
	
	// 检查并添加新的直播间
	for _, live := range liveList {
		userID := live.Profile.UserID
		
		// 跳过过滤列表中的主播
		if m.filterList.ShouldFilter(userID) {
			continue
		}
		
		activeIDs[userID] = true
		
		// 如果已经在监听，则跳过
		m.mu.RLock()
		_, exists := m.listeners[userID]
		m.mu.RUnlock()
		
		if exists {
			continue
		}
		
		// 添加新的监听器
		err := m.addListener(live)
		if err != nil {
			log.Printf("添加监听器失败 (UID: %d): %v", userID, err)
		}
	}
	
	// 移除已经不在直播的监听器
	m.mu.Lock()
	for uid, listener := range m.listeners {
		if !activeIDs[uid] {
			log.Printf("直播间已关闭，移除监听器: %s (UID: %d)", listener.nickname, uid)
			listener.cancel()
			delete(m.listeners, uid)
			
			// 从全局存储中移除
			globalStore.RemoveLiveRoom(uid)
			
			// 从直播流管理器中移除
			streamManager.RemoveStream(uid)
		}
	}
	m.mu.Unlock()
	
	log.Printf("直播列表刷新完成，当前监听 %d 个直播间", len(activeIDs))
}

// addListener 添加新的直播间监听器
func (m *DanmuMonitor) addListener(live acfundanmu.UserLiveInfo) error {
	userID := live.Profile.UserID
	nickname := live.Profile.Nickname
	title := live.Title
	
	log.Printf("添加新直播间监听器: %s (UID: %d) - %s", nickname, userID, title)
	
	// 创建新的AcFunLive实例，共享主客户端的Token信息
	tokenInfo := m.client.GetTokenInfo()
	acLive, err := acfundanmu.NewAcFunLive(
		acfundanmu.SetLiverUID(userID),
		acfundanmu.SetTokenInfo(tokenInfo),
	)
	if err != nil {
		return fmt.Errorf("创建AcFunLive实例失败: %w", err)
	}
	
	// 创建上下文
	listenerCtx, cancel := context.WithCancel(m.ctx)
	
	// 创建监听器
	listener := &LiveListener{
		client:   acLive,
		userID:   userID,
		nickname: nickname,
		title:    title,
		liveID:   live.LiveID,
		ctx:      listenerCtx,
		cancel:   cancel,
		lastComment: time.Now(),
	}
	
	// 注册弹幕处理函数
	acLive.OnComment(listener.handleComment)
	acLive.OnDanmuStop(func(ac *acfundanmu.AcFunLive, err error) {
		if err != nil {
			log.Printf("直播间 %s (UID: %d) 弹幕监听停止: %v", listener.nickname, listener.userID, err)
		}
		// 下次刷新时会移除该监听器
	})
	
	// 获取直播流信息并保存
	streamInfo := acLive.GetStreamInfo()
	if streamInfo != nil {
		// 保存到直播流管理器
		streamManager := global.GetStreamManager()
		streamManager.SetStreamInfo(userID, *streamInfo)
		
		// 找出最低画质的流 (不再打印，但仍然计算以便将来使用)
		var lowestBitrate int = -1
		
		for _, stream := range streamInfo.StreamList {
			if lowestBitrate == -1 || stream.Bitrate < lowestBitrate {
				lowestBitrate = stream.Bitrate
			}
		}
	}
	
	// 开始监听
	errCh := acLive.StartDanmu(listenerCtx, true)
	
	// 处理错误
	go func() {
		err := <-errCh
		if err != nil {
			log.Printf("直播间 %s (UID: %d) 弹幕监听错误: %v", listener.nickname, listener.userID, err)
		}
	}()
	
	// 添加到监听器映射
	m.mu.Lock()
	m.listeners[userID] = listener
	m.mu.Unlock()
	
	// 添加到全局存储
	globalStore := global.GetStore()
	globalStore.AddLiveRoom(&global.LiveRoomInfo{
		UserID:   userID,
		LiveID:   live.LiveID,
		Nickname: nickname,
		Title:    title,
		Client:   acLive,
	})
	
	// 如果启用了自动点赞，则启动自动点赞协程
	if m.autoLike {
		go func() {
			log.Printf("启动对直播间 %s (UID: %d) 的自动点赞，间隔 %d 秒", nickname, userID, m.likeInterval)
			// 确保使用新创建的上下文
			likeTicker := time.NewTicker(time.Duration(m.likeInterval) * time.Second)
			defer likeTicker.Stop()
			
			// 立即发送一次点赞
			if err := action.SendLike(acLive, live.LiveID, 1, 800); err != nil {
				log.Printf("直播间 %s (UID: %d) 点赞失败: %v", nickname, userID, err)
			}
			
			for {
				select {
				case <-likeTicker.C:
					if err := action.SendLike(acLive, live.LiveID, 1, 800); err != nil {
						log.Printf("直播间 %s (UID: %d) 点赞失败: %v", nickname, userID, err)
					}
				case <-listenerCtx.Done():
					log.Printf("直播间 %s (UID: %d) 停止自动点赞", nickname, userID)
					return
				}
			}
		}()
	}
	
	return nil
}

// 处理弹幕
func (l *LiveListener) handleComment(_ *acfundanmu.AcFunLive, comment *acfundanmu.Comment) {
	l.lastComment = time.Now()
	
	timestamp := time.Unix(comment.SendTime/1000, 0).Format("15:04:05")
	fmt.Printf("[%s] [%s] %s: %s\n", 
		timestamp, 
		l.nickname, 
		comment.Nickname, 
		comment.Content,
	)
	
	// 这里可以添加弹幕处理逻辑
	// 如保存到数据库、触发特定行为等
}

// GetActiveListenerCount 获取当前活跃的监听器数量
func (m *DanmuMonitor) GetActiveListenerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.listeners)
}

// GetListenerStatus 获取监听器状态
func (m *DanmuMonitor) GetListenerStatus() map[int64]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	status := make(map[int64]string)
	for uid, listener := range m.listeners {
		status[uid] = fmt.Sprintf("%s - %s (最后弹幕: %s)", 
			listener.nickname, 
			listener.title,
			listener.lastComment.Format("15:04:05"),
		)
	}
	
	return status
} 