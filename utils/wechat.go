package utils

import (
	"encoding/json"
	"io/ioutil"
	"log/slog"
	"net/http"
	"stock/logger"
	"strings"
	"time"
)

func init() {
	// 确保日志系统已初始化
	logger.Init()
}

type TokenInfo struct {
	Access_token string `json:"access_token"`
	Errcode      int    `json:"errcode"`
}

type Con struct {
	Content string `json:"content"`
}

type Message struct {
	Touser  string `json:"touser"`
	Toparty string `json:"toparty"`
	Msgtype string `json:"msgtype"`
	Agentid string `json:"agentid"`
	Text    Con    `json:"text"`
	Safe    string `json:"safe"`
}

var (
	m_appID     = "wxde9f70b5702a6cb4"
	m_appsecret = "71e61317e46772a16d19e5bda8010af9"
	m_touser    = "ofL-H06NfYiu5WQ_n-lKratYPiVs"

	m_corpid     = "ww1907ef862b680387"
	m_corpsecret = "5PGSetsGMdxql_qamMHRBVh1uKssWyA2y41S424kJmM"
)

func GetToken() (tokenRes string) {
	// 创建 client 和 resp 对象
	var client http.Client
	var resp *http.Response

	// 设置了10秒钟的超时
	client = http.Client{Timeout: 10 * time.Second}
	var err error

	// 这里使用了 Get 方法，并判断异常
	resp, err = client.Get("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=ww1907ef862b680387&corpsecret=5PGSetsGMdxql_qamMHRBVh1uKssWyA2y41S424kJmM")
	if err != nil {
		return tokenRes
	}
	// 释放对象
	defer resp.Body.Close()

	// 把获取到的页面作为返回值返回
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return tokenRes
	}

	myInfo := TokenInfo{}
	json.Unmarshal(body, &myInfo)

	// 释放对象
	defer client.CloseIdleConnections()

	slog.Info("received response", "response", string(body), "token", myInfo.Access_token)
	return myInfo.Access_token
}

func SendWeChatMessage(message string) {
	m1 := Message{
		Touser:  "@all",
		Toparty: "1",
		Msgtype: "text",
		Agentid: "1000002",
		Text:    Con{message},
		Safe:    "0",
	}
	send, _ := json.Marshal(m1)
	// 创建 client 和 resp 对象
	var client http.Client
	var resp *http.Response
	// 设置了10秒钟的超时
	client = http.Client{Timeout: 10 * time.Second}
	url := "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=" + GetToken()
	// 这里使用了 Get 方法，并判断异常
	resp, err := client.Post(url,
		"application/json",
		strings.NewReader(string(send)))
	if err != nil {
		return
	}
	// 释放对象
	defer resp.Body.Close()
	// 把获取到的页面作为返回值返回
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return
	}
	slog.Info("received response", "response", string(body))
	// 释放对象
	defer client.CloseIdleConnections()
	return
}
