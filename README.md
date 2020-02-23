# GoLoginHust

## 简介

一个基于Go的HustPass登录库，用于登录华科的统一认证登录系统，可用于校内绝大多数平台。

代码参考了[@naivekun](https://github.com/naivekun)的使用python编写的[登录库](https://github.com/naivekun/libhustpass)。

## 使用方法

这个库使用了Golang的新的包管理方式，请确保你的Golang版本高于1.12，并在代码中使用import导入本库之后，使用go mod tidy更新依赖，**而不是使用go get**。

代码示例

```
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/black-binary/gologinhust"
)

func main() {
	client, _ := gologinhust.GetLoginClient("你的用户名", "你的密码", "http://hubs.hust.edu.cn/hustpass.action")
	postParam := url.Values{}
	postParam.Add("start", "2020-01-01")
	postParam.Add("end", "2020-03-01")
	req, _ := http.NewRequest("POST", "http://hubs.hust.edu.cn/aam/score/CourseInquiry_ido.action", strings.NewReader(postParam.Encode()))
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf(string(body))
}

```

GetLoginClient函数将尝试登录该url，函数将返回一个http.Client，这个client已经包含登录该平台所需的cookie。示例代码调用了课程表接口，获取了1月1日至1月3日的课程表。注意在访问一些特定接口的时候，可能需要先抓包然后设置对应http头部字段。

## 实现细节

HustPass更新后由原来的RSA改为使用变形的DES三重加密，三个密钥分别是"1", "2", "3"，对pc1变换矩阵进行了修改，其他基本一致。参见[naivekun实现的登录库](https://github.com/naivekun/libhustpass) (wtf, 这样不是安全性更低了吗??)。代码中的sbdes库是在go的cipher密码学支持库中的des中复制出来然后修改的。

GetLoginClient将使用GET访问目标URL，并跟随重定向到达HustPass登录界面。然后使用输入的用户名和密码登录。注意Go的HTTP库重定向跟随支持有一些问题，校内[一些平台](http://one.hust.edu.cn)使用http-equiv="refresh"的方法进行重定向，可能导致重定向跟随失败从而无法正确登录。解决方法是使用完整的[主页路径](http://one.hust.edu.cn/dcp/index.jsp)进行登录。



## 许可

GNU GPLv3