package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"

	"github.com/PuerkitoBio/goquery"
)

type configData struct {
	LoginEmail    string
	LoginPassword string
	LoginURL      string
}

var cd configData

// ログイン前のサイト内にtokenが入っているので、htmlを取得。
// サイト内の構造を取り出して<input>内からtokenを取得
func getToken(client *http.Client) string {
	req, _ := http.NewRequest("GET", cd.LoginURL, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var token string
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		key, _ := s.Attr("name")
		val, _ := s.Attr("value")
		if key == "token" {
			token = val
		}
	})
	return token
}
func printRequest(r *http.Request) {
	d, _ := httputil.DumpRequest(r, true)
	fmt.Printf("===Dump Request[START]\n%s\n===Dump Request[END]\n", d)
}
func printResponse(r *http.Response) {
	d, _ := httputil.DumpResponse(r, true)
	fmt.Printf("===Dump Response[START]\n%s\n===Dump Response[END]\n", d)
}
func searchShop(c *http.Client) {
	url := "https://my.hamazushi.com/shop/search/"
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		kenURL, _ := s.Attr("href")
		kenName := s.Text()
		if len(kenURL) > 21 && kenURL[:21] == "/shop/search/extract/" {
			fmt.Printf("data = %s,%s\n", kenURL, kenName)
			searchShop2(c, kenURL)
		}
	})
}
func searchShop2(c *http.Client, url string) {
	req, _ := http.NewRequest("GET", "https://my.hamazushi.com"+url, nil)
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	printResponse(resp)
}
func getWaitNumber(c *http.Client) {
	url := "https://my.hamazushi.com/shop/?shopId=10121"
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

}
func main() {
	// コンフィグファイルを読み込む
	f, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(f, &cd)
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	// ログイン前にtokenを取得する。
	token := getToken(client)
	if token == "" {
		log.Fatal("token is empty")
	}

	// 2度目のリクエスト、ここでログインする。
	query := fmt.Sprintf("login-email=%s&login-password=%s&token=%s", cd.LoginEmail, cd.LoginPassword, token)
	str := []byte(query)
	req, _ := http.NewRequest("POST", cd.LoginURL, bytes.NewBuffer(str))

	// このヘッダは必要。
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// 店舗情報を取得
	searchShop(client)
	// 3度目のリクエスト、ログイン済みなので、取れる情報が変わる。
	//	getWaitNumber(client)
}
