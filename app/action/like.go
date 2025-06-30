package action

import (
	"fmt"
	"log"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/orzogc/acfundanmu"
)

const (
	likeURL = "https://api.kuaishouzt.com/rest/zt/live/web/audience/action/like?subBiz=mainApp&kpn=ACFUN_APP&kpf=PC_WEB&userId=%d&did=%s&%s=%s"
)

// SendLike 发送点赞请求到AcFun直播间
// count: 点赞次数，通常为1
// durationMs: 点赞动画持续时间，单位毫秒，通常为800
func SendLike(client *acfundanmu.AcFunLive, liveID string, count int, durationMs int) error {
	tokenInfo := client.GetTokenInfo()
	
	// 构建请求URL
	apiURL := fmt.Sprintf(likeURL, 
		tokenInfo.UserID, 
		tokenInfo.DeviceID, 
		"acfun.midground.api_st", 
		tokenInfo.ServiceToken,
	)
	
	// 构建请求参数
	form := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(form)
	
	form.Set("visitorId", fmt.Sprintf("%d", tokenInfo.UserID))
	form.Set("liveId", liveID)
	form.Set("count", fmt.Sprintf("%d", count))
	form.Set("durationMs", fmt.Sprintf("%d", durationMs))
	
	// 发送请求
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	
	req.SetRequestURI(apiURL)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.SetReferer("https://live.acfun.cn/")
	req.Header.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")
	req.Header.Set("Origin", "https://live.acfun.cn")
	req.SetBody(form.QueryString())
	
	// 执行请求
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	
	err := fasthttp.Do(req, resp)
	if err != nil {
		return fmt.Errorf("发送点赞请求失败: %v", err)
	}
	
	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("点赞请求返回非200状态码: %d", resp.StatusCode())
	}
	
	log.Printf("成功向直播间 %s 发送点赞", liveID)
	return nil
}

// StartAutoLike 开始自动点赞
// interval: 点赞间隔，单位秒
func StartAutoLike(client *acfundanmu.AcFunLive, liveID string, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	
	log.Printf("开始对直播间 %s 每 %d 秒自动点赞", liveID, interval)
	
	// 立即发送一次点赞
	if err := SendLike(client, liveID, 1, 800); err != nil {
		log.Printf("自动点赞失败: %v", err)
	}
	
	for range ticker.C {
		if err := SendLike(client, liveID, 1, 800); err != nil {
			log.Printf("自动点赞失败: %v", err)
		}
	}
} 