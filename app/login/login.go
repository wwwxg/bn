package login

import (
	"github.com/orzogc/acfundanmu"
)

// Login 使用账号密码登录AcFun
func Login(account, password string) (acfundanmu.Cookies, error) {
	return acfundanmu.Login(account, password)
}
