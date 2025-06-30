package action

import (
	"fmt"

	"bn/app/global"

	"github.com/orzogc/acfundanmu"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

const (
	// 发送弹幕的API地址
	commentURL = "https://api.kuaishouzt.com/rest/zt/live/web/audience/action/comment?subBiz=mainApp&kpn=ACFUN_APP&kpf=PC_WEB&userId=%d&did=%s&%s=%s"
)

// SendComment 发送弹幕到指定直播间
// 这是底层API函数，需要提供具体的client和liveID
// client: AcFunLive客户端实例
// liveID: 直播ID，不是主播UID
// content: 弹幕内容
// 返回错误信息，nil表示发送成功
func SendComment(client *acfundanmu.AcFunLive, liveID string, content string) error {
	if client == nil {
		return fmt.Errorf("客户端为空")
	}

	if liveID == "" {
		return fmt.Errorf("直播ID为空")
	}

	if content == "" {
		return fmt.Errorf("弹幕内容为空")
	}

	// 获取必要的令牌信息
	tokenInfo := client.GetTokenInfo()
	if tokenInfo == nil {
		return fmt.Errorf("获取令牌信息失败")
	}

	userID := tokenInfo.UserID
	deviceID := tokenInfo.DeviceID
	serviceToken := tokenInfo.ServiceToken

	// 构建请求URL
	reqURL := fmt.Sprintf(commentURL, userID, deviceID, "acfun.midground.api_st", serviceToken)

	// 构建请求参数
	form := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(form)

	form.Set("visitorId", fmt.Sprintf("%d", userID))
	form.Set("liveId", liveID)
	form.Set("content", content)

	// 发送请求
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqURL)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://live.acfun.cn")
	req.Header.Set("Referer", "https://live.acfun.cn/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")
	req.SetBody(form.QueryString())

	if err := fasthttp.Do(req, resp); err != nil {
		return fmt.Errorf("发送弹幕请求失败: %w", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("发送弹幕请求状态码错误: %d", resp.StatusCode())
	}

	// 解析响应
	body := resp.Body()
	var p fastjson.Parser
	v, err := p.ParseBytes(body)
	if err != nil {
		return fmt.Errorf("解析弹幕响应失败: %w", err)
	}

	result := v.GetInt("result")
	if result != 1 {
		return fmt.Errorf("发送弹幕失败，响应: %s", string(body))
	}

	return nil
}

// SendCommentToRoom 向指定UID的直播间发送弹幕
// 这是推荐使用的高级函数，只需提供主播UID和弹幕内容
// uid: 主播的UID
// content: 弹幕内容
// 返回错误信息，nil表示发送成功
// 
// 使用示例:
// err := action.SendCommentToRoom(23682490, "你好，这是一条测试弹幕")
// if err != nil {
//     log.Printf("发送弹幕失败: %v", err)
// }
func SendCommentToRoom(uid int64, content string) error {
	// 获取全局存储
	store := global.GetStore()
	
	// 获取直播间信息
	room := store.GetLiveRoom(uid)
	if room == nil {
		return fmt.Errorf("未找到UID %d的直播间信息", uid)
	}
	
	// 发送弹幕
	return SendComment(room.Client, room.LiveID, content)
} 