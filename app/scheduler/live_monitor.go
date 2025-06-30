package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/orzogc/acfundanmu"
)

// LiveMonitor 直播监控器结构体
type LiveMonitor struct {
	ac        *acfundanmu.AcFunLive
	interval  time.Duration
	filterList *FilterList
}

// NewLiveMonitor 创建新的直播监控器
func NewLiveMonitor(ac *acfundanmu.AcFunLive, interval time.Duration, filterListPath string) *LiveMonitor {
	return &LiveMonitor{
		ac:        ac,
		interval:  interval,
		filterList: NewFilterList(filterListPath),
	}
}

// Start 开始监控直播列表
func (m *LiveMonitor) Start() {
	log.Println("开始监控AcFun直播列表")
	
	// 首次获取直播列表
	m.fetchAndProcessLiveList()
	
	// 设置定时器，定期获取直播列表
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	
	for range ticker.C {
		m.fetchAndProcessLiveList()
	}
}

// fetchAndProcessLiveList 获取并处理直播列表
func (m *LiveMonitor) fetchAndProcessLiveList() {
	// 获取所有直播列表
	liveList, err := m.ac.GetAllLiveList()
	if err != nil {
		log.Printf("获取直播列表失败: %v", err)
		return
	}
	
	// 过滤掉需要排除的主播
	var filteredList []acfundanmu.UserLiveInfo
	for _, live := range liveList {
		if !m.filterList.ShouldFilter(live.Profile.UserID) {
			filteredList = append(filteredList, live)
		}
	}
	
	log.Printf("成功获取到%d个直播间，过滤后剩余%d个", len(liveList), len(filteredList))
	
	// 处理每个直播间信息
	for _, live := range filteredList {
		fmt.Printf("主播: %s(ID:%d), 标题: %s, 人气: %d, 直播ID: %s\n", 
			live.Profile.Name, live.Profile.UserID, live.Title, live.OnlineCount, live.LiveID)
	}
}

// ReloadFilterList 重新加载过滤列表
func (m *LiveMonitor) ReloadFilterList(filePath string) error {
	return m.filterList.LoadFromFile(filePath)
}

// StartFollowingLiveMonitor 兼容旧接口的函数
func StartFollowingLiveMonitor(ac *acfundanmu.AcFunLive, interval time.Duration) {
	monitor := NewLiveMonitor(ac, interval, "app/config/filter_list.txt")
	monitor.Start()
} 