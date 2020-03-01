package gologinhust

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestEncrypt(t *testing.T) {
	nonce := "LT-267469-asdL7NqjEizVxBS0Z9afD9l7axu7TF-cas"
	username := "fuck"
	password := "you"
	mine := encrypt(username+password+nonce, "1", "2", "3")
	wanted := "1C1F28703B90E489E3A62F786DBB9345CDA12FBD42FB8805D45A635FE15E661380CD7478FED8BC725EDAFE28BFDEFE8CFB18028EB731A9907C685198E33C4BA6B37ADDDBD85C4AA0BA094406EA558FC921DC99C33262E8D458A42B403F8BA1AF369041DED879FBF3"
	if mine != wanted {
		t.Fatal("not equal my:" + mine)
	}
}

func TestGetLoginClient(t *testing.T) {
	client, _ := GetLoginClient("Your username", "Your password", "http://hubs.hust.edu.cn/hustpass.action")
	resp, _ := client.Get("http://hubs.hust.edu.cn/hustpass.action")
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	LogoutClient(client)
	fmt.Println("-----------------------------------------")
	resp, _ = client.Get("http://hubs.hust.edu.cn/hustpass.action")
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

}
