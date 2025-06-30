package global

import (
	"sync"

	"github.com/orzogc/acfundanmu"
)

// StreamManager 管理直播间的直播流资源
type StreamManager struct {
	streams     map[int64]acfundanmu.StreamInfo // 用主播UID作为key存储直播流信息
	mutex       sync.RWMutex                    // 读写锁保护并发访问
}

var (
	manager     *StreamManager
	managerOnce sync.Once
)

// GetStreamManager 获取全局StreamManager实例
func GetStreamManager() *StreamManager {
	managerOnce.Do(func() {
		manager = &StreamManager{
			streams: make(map[int64]acfundanmu.StreamInfo),
		}
	})
	return manager
}

// SetStreamInfo 保存主播的直播流信息
func (sm *StreamManager) SetStreamInfo(uid int64, streamInfo acfundanmu.StreamInfo) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.streams[uid] = streamInfo
}

// GetStreamInfo 获取主播的直播流信息
func (sm *StreamManager) GetStreamInfo(uid int64) (acfundanmu.StreamInfo, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	info, exists := sm.streams[uid]
	return info, exists
}

// GetLowestQualityStream 获取主播的最低画质直播流
func (sm *StreamManager) GetLowestQualityStream(uid int64) (acfundanmu.StreamURL, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	info, exists := sm.streams[uid]
	if !exists || len(info.StreamList) == 0 {
		return acfundanmu.StreamURL{}, false
	}
	
	// 找到码率最低的流
	lowestStream := info.StreamList[0]
	for _, stream := range info.StreamList {
		if stream.Bitrate < lowestStream.Bitrate {
			lowestStream = stream
		}
	}
	
	return lowestStream, true
}

// RemoveStream 移除主播的直播流信息
func (sm *StreamManager) RemoveStream(uid int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if _, exists := sm.streams[uid]; exists {
		delete(sm.streams, uid)
	}
}

// GetAllStreamUIDs 获取所有保存的直播流UID列表
func (sm *StreamManager) GetAllStreamUIDs() []int64 {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	uids := make([]int64, 0, len(sm.streams))
	for uid := range sm.streams {
		uids = append(uids, uid)
	}
	return uids
} 