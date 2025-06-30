package login

import (
	"github.com/orzogc/acfundanmu"
	"github.com/valyala/fasthttp"
)

func Login(account, password string) ([]*fasthttp.Cookie, error) {
	cookies, err := acfundanmu.Login(account, password)
	if err != nil {
		return nil, err
	}
	return cookies, nil
}
