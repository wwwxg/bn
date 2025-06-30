package action

import (
	"fmt"
	"log"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/orzogc/acfundanmu"
)

const (
	giftSendURL = "https://api.kuaishouzt.com/rest/zt/live/web/gift/send?subBiz=mainApp&kpn=ACFUN_APP&kpf=PC_WEB&userId=%d&did=%s&%s=%s"
)

// SendGift 发送礼物请求到AcFun直播间
// giftId: 礼物ID，普通的香蕉是1
// batchSize: 礼物数量
// comboKey: 连击标识，格式通常为"1_1_时间戳"
func SendGift(client *acfundanmu.AcFunLive, liveID string, giftID int, batchSize int, comboKey string) error {
	tokenInfo := client.GetTokenInfo()
	
	// 构建请求URL
	apiURL := fmt.Sprintf(giftSendURL, 
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
	form.Set("giftId", fmt.Sprintf("%d", giftID))
	form.Set("batchSize", fmt.Sprintf("%d", batchSize))
	
	// 如果没有提供comboKey，则生成一个
	if comboKey == "" {
		comboKey = fmt.Sprintf("1_1_%d", time.Now().UnixMilli())
	}
	form.Set("comboKey", comboKey)
	
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
		return fmt.Errorf("发送礼物请求失败: %v", err)
	}
	
	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("礼物请求返回非200状态码: %d", resp.StatusCode())
	}
	
	// 打印响应体（用于调试）
	responseBody := resp.Body()
	log.Printf("成功向直播间 %s 发送礼物，响应: %s", liveID, string(responseBody))
	
	return nil
} 