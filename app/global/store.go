package global

import (
	"sync"

	"github.com/orzogc/acfundanmu"
	"github.com/valyala/fasthttp"
)

// GlobalStore 存储全局共享的API令牌和直播信息
type GlobalStore struct {
	mu               sync.RWMutex
	mainClient       *acfundanmu.AcFunLive     // 主客户端实例
	deviceID         string                    // 设备ID
	userID           int64                     // 用户ID
	serviceToken     string                    // API令牌 (acfun.midground.api_st)
	securityKey      string                    // 安全密钥
	cookies          []*fasthttp.Cookie        // 完整的Cookie列表
	activeLiveRooms  map[int64]*LiveRoomInfo   // 当前活跃的直播间信息
}

// LiveRoomInfo 存储单个直播间的信息
type LiveRoomInfo struct {
	UserID     int64                 // 主播UID
	LiveID     string                // 直播ID
	Nickname   string                // 主播昵称
	Title      string                // 直播标题
	Client     *acfundanmu.AcFunLive // 该直播间的客户端实例
}

var (
	store     *GlobalStore
	storeOnce sync.Once
)

// GetStore 获取全局存储的单例实例
func GetStore() *GlobalStore {
	storeOnce.Do(func() {
		store = &GlobalStore{
			activeLiveRooms: make(map[int64]*LiveRoomInfo),
		}
	})
	return store
}

// SetMainClient 设置主客户端实例
func (s *GlobalStore) SetMainClient(client *acfundanmu.AcFunLive) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.mainClient = client
	
	// 提取令牌信息
	tokenInfo := client.GetTokenInfo()
	s.deviceID = tokenInfo.DeviceID
	s.userID = tokenInfo.UserID
	s.serviceToken = tokenInfo.ServiceToken
	s.securityKey = tokenInfo.SecurityKey
	s.cookies = tokenInfo.Cookies
}

// GetMainClient 获取主客户端实例
func (s *GlobalStore) GetMainClient() *acfundanmu.AcFunLive {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mainClient
}

// GetAPIToken 获取API令牌 (acfun.midground.api_st)
func (s *GlobalStore) GetAPIToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.serviceToken
}

// GetDeviceID 获取设备ID
func (s *GlobalStore) GetDeviceID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deviceID
}

// GetUserID 获取用户ID
func (s *GlobalStore) GetUserID() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userID
}

// GetCookies 获取Cookies
func (s *GlobalStore) GetCookies() []*fasthttp.Cookie {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cookies
}

// AddLiveRoom 添加直播间信息
func (s *GlobalStore) AddLiveRoom(info *LiveRoomInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeLiveRooms[info.UserID] = info
}

// RemoveLiveRoom 移除直播间信息
func (s *GlobalStore) RemoveLiveRoom(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.activeLiveRooms, userID)
}

// GetLiveRoom 获取直播间信息
func (s *GlobalStore) GetLiveRoom(userID int64) *LiveRoomInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeLiveRooms[userID]
}

// GetAllLiveRooms 获取所有直播间信息
func (s *GlobalStore) GetAllLiveRooms() map[int64]*LiveRoomInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// 返回副本以避免并发修改问题
	result := make(map[int64]*LiveRoomInfo, len(s.activeLiveRooms))
	for k, v := range s.activeLiveRooms {
		result[k] = v
	}
	return result
}

// GetLiveRoomCount 获取直播间数量
func (s *GlobalStore) GetLiveRoomCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.activeLiveRooms)
} 