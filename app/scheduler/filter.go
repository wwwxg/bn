package scheduler

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// FilterList 过滤列表结构体
type FilterList struct {
	userIDs map[int64]struct{}
	mu      sync.RWMutex
}

// NewFilterList 创建并初始化过滤列表
func NewFilterList(filePath string) *FilterList {
	fl := &FilterList{
		userIDs: make(map[int64]struct{}),
	}
	fl.LoadFromFile(filePath)
	return fl
}

// LoadFromFile 从文件加载过滤列表
func (fl *FilterList) LoadFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("过滤列表文件 %s 不存在，将使用空过滤列表", filePath)
			return nil
		}
		return err
	}
	defer file.Close()

	fl.mu.Lock()
	defer fl.mu.Unlock()
	
	// 清空现有列表
	fl.userIDs = make(map[int64]struct{})

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// 尝试转换为int64
		userID, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			log.Printf("警告：过滤列表文件第%d行无法解析为用户ID: %s", lineNum, line)
			continue
		}
		
		// 添加到过滤列表
		fl.userIDs[userID] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	log.Printf("成功从 %s 加载了 %d 个过滤用户ID", filePath, len(fl.userIDs))
	return nil
}

// ShouldFilter 判断用户ID是否应该被过滤
func (fl *FilterList) ShouldFilter(userID int64) bool {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	
	_, exists := fl.userIDs[userID]
	return exists
}

// AddUserID 添加用户ID到过滤列表
func (fl *FilterList) AddUserID(userID int64) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	
	fl.userIDs[userID] = struct{}{}
}

// RemoveUserID 从过滤列表移除用户ID
func (fl *FilterList) RemoveUserID(userID int64) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	
	delete(fl.userIDs, userID)
} 