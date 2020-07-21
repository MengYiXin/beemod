package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type WxClient struct {
	AppID       string
	AppSecret   string
	AccessToken string
	OpenID      string
	Name        string
	RedirectURI string
}

type wxGetTokenResponse struct {
	AccessToken string `json:"access_token"`
	Openid      string `json:"openid"`
	ErrorCode   int    `json:"errcode"`
	ErrorMsg    string `json:"errmsg"`
}

type wxGetWxInfoResponse struct {
	Openid     string `json:"openid"`
	Nickname   string `json:"nickname"`
	HeadImgurl string `json:"headimgurl"`
	Unionid    string `json:"unionid"`
	ErrorCode  int    `json:"errcode"`
	ErrorMsg   string `json:"errmsg"`
}

func NewWxOauth2Service(app_id, app_secret, redirectURI string) SocialService {
	return &WxClient{
		AppID:       app_id,
		AppSecret:   app_secret,
		RedirectURI: redirectURI,
		Name:        "wx",
	}
}

//get login page
func (c *WxClient) LoginPage(state string) string {
	return "https://open.weixin.qq.com/connect/qrconnect?appid=" + c.AppID + "&redirect_uri=" + c.RedirectURI + "&response_type=code&scope=snsapi_login&state=" + state
}

func (c *WxClient) GetAccessToken(code string) (*BasicTokenInfo, error) {
	var openIdUrl = "https://api.weixin.qq.com/sns/oauth2/access_token"
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, openIdUrl, nil); err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("code", code)
	q.Add("appid", c.AppID)
	q.Add("secret", c.AppSecret)
	q.Add("grant_type", "authorization_code")
	req.URL.RawQuery = q.Encode()

	var client = http.Client{
		Timeout: 10 * time.Second,
	}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var ret wxGetTokenResponse
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	if ret.ErrorCode != 0 {
		return nil, errors.New(fmt.Sprintf("GetWxTokenByCode error = %d", ret.ErrorCode))
	}
	basicToken := BasicTokenInfo{}
	basicToken.AccessToken = ret.AccessToken
	c.OpenID = ret.Openid
	return &basicToken, nil
}

func (c *WxClient) GetUserInfo(accessToken string) (*BasicUserInfo, error) {

	var openIdUrl = "https://api.weixin.qq.com/sns/userinfo"
	var req *http.Request
	var err error
	if req, err = http.NewRequest(http.MethodGet, openIdUrl, nil); err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("access_token", accessToken)
	q.Add("openid", c.OpenID)
	req.URL.RawQuery = q.Encode()

	var client = http.Client{
		Timeout: 10 * time.Second,
	}
	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var ret wxGetWxInfoResponse
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}

	if ret.ErrorCode != 0 {
		return nil, errors.New(fmt.Sprintf("GetWxInfoByToken error = %d", ret.ErrorCode))
	}

	basicUser := BasicUserInfo{
		NickName: ret.Nickname,
		HeadIcon: ret.HeadImgurl,
	}

	return &basicUser, nil
}

func (c *WxClient) GetType() string {
	return c.Name
}
