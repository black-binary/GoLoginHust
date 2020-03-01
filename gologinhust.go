package gologinhust

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/black-binary/gologinhust/sbdes"
)

func getNonce(source string) string {
	exp, _ := regexp.Compile(`<input type="hidden" id="lt" name="lt" value="(.*)" />`)
	findResult := exp.FindStringSubmatch(source)
	if len(findResult) != 2 {
		return ""
	}
	nonce := findResult[1]
	return nonce
}

func getAction(source string) string {
	exp, _ := regexp.Compile(`<form id="loginForm" action="(.*)" method="post">`)
	findResult := exp.FindStringSubmatch(source)
	if len(findResult) != 2 {
		return ""
	}
	action := findResult[1]
	return action
}

func padAndAlign(s string) []byte {
	//扩展为宽字节
	b := []byte(s)
	buf := bytes.NewBufferString("")
	for _, v := range b {
		buf.WriteByte(0)
		buf.WriteByte(v)
	}

	//末尾填充0，对齐8
	for buf.Len()%8 != 0 {
		buf.WriteByte(0)
	}
	return buf.Bytes()
}

func encrypt(data string, key1 string, key2 string, key3 string) string {
	alignedData := padAndAlign(data)
	alignedKey1 := padAndAlign(key1)
	alignedKey2 := padAndAlign(key2)
	alignedKey3 := padAndAlign(key3)

	buf := bytes.NewBuffer([]byte{})

	for i := 0; i < len(alignedData); i += 8 {
		tmp := make([]byte, 8)
		copy(tmp, alignedData[i:i+8])
		//Encrypt函数使用了unsafe指针操作，稳妥起见使用复制

		for j := 0; j < len(alignedKey1); j += 8 {
			block, _ := sbdes.NewCipher(alignedKey1[j : j+8])
			tmpOriginal := make([]byte, 8)
			copy(tmpOriginal, tmp)
			block.Encrypt(tmp, tmpOriginal)
		}

		for j := 0; j < len(alignedKey2); j += 8 {
			block, _ := sbdes.NewCipher(alignedKey2[j : j+8])
			tmpOriginal := make([]byte, 8)
			copy(tmpOriginal, tmp)
			block.Encrypt(tmp, tmpOriginal)
		}

		for j := 0; j < len(alignedKey3); j += 8 {
			block, _ := sbdes.NewCipher(alignedKey3[j : j+8])
			tmpOriginal := make([]byte, 8)
			copy(tmpOriginal, tmp)
			block.Encrypt(tmp, tmpOriginal)
		}

		buf.Write(tmp)
	}

	result := ""

	for _, v := range buf.Bytes() {
		result += fmt.Sprintf("%02X", v)
	}

	return result
}

//GetLoginClient 使用用户名和密码登录，获得targetURL对应平台的含有合法cookies的HTTP client
func GetLoginClient(username string, password string, targetURL string) (*http.Client, error) {
	//设置调试代理
	/*
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse("http://127.0.0.1:8080")
		}
	*/
	transport := &http.Transport{
		//Proxy:           proxy,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	jar, _ := cookiejar.New(nil)
	client := http.Client{
		Transport: transport,
		Jar:       jar,
	}
	req, err := http.NewRequest("GET", targetURL, nil)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if !strings.Contains(resp.Request.URL.String(), "pass.hust.edu.cn") {
		return nil, errors.New("Didn't redirected to hust pass")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	nonce := getNonce(string(body))

	shit := encrypt(username+password+nonce, "1", "2", "3")

	postParm := url.Values{}
	postParm.Add("rsa", shit)
	postParm.Add("ul", fmt.Sprintf("%d", len(username)))
	postParm.Add("pl", fmt.Sprintf("%d", len(password)))
	postParm.Add("lt", nonce)
	postParm.Add("execution", "e1s1")
	postParm.Add("_eventId", "submit")

	req, _ = http.NewRequest("POST", "https://pass.hust.edu.cn/cas/login?service="+url.PathEscape(targetURL), strings.NewReader(postParm.Encode()))
	//action := getAction(string(body))
	//req, _ = http.NewRequest("POST", "https://pass.hust.edu.cn"+action, strings.NewReader(postParm.Encode()))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	ticket := ""

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		t, _ := req.URL.Query()["ticket"]
		if t != nil && len(t) > 0 {
			ticket = t[0]
			fmt.Printf("ticket found %s!\n", ticket)
		}
		return nil
	}
	client.Do(req)

	if ticket != "" {
		return &client, nil
	} else {
		return nil, errors.New("Failed to get ticket")
	}

}

//LogoutClient 注销一个已经登录的Client
func LogoutClient(client *http.Client) {
	req, _ := http.NewRequest("GET", "http://pass.hust.edu.cn/portal/logout.jsp", strings.NewReader("?service=http://one.hust.edu.cn"))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36")
	client.Do(req)
}
