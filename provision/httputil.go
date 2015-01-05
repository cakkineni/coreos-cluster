package provision

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type HttpUtil struct {
	APIEndPoint string
	HttpClient  http.Client
	Headers     []KeyValue
}

type KeyValue struct {
	Key   string
	Value string
}

type localCookieJar struct {
	jar map[string][]*http.Cookie
}

func (p *localCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	p.jar[u.Host] = cookies
}

func (p *localCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return p.jar[u.Host]
}

func NewHttpUtil() HttpUtil {
	retVal := HttpUtil{}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)
	retVal.HttpClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool},
		},
	}
	jar := &localCookieJar{}
	jar.jar = make(map[string][]*http.Cookie)
	retVal.HttpClient.Jar = jar
	return retVal
}

func (httpUtil HttpUtil) doBasicAuth(url string, username string, password string, params interface{}) string {
	url1 := httpUtil.APIEndPoint + url
	postData, _ := json.Marshal(params)
	reqData := strings.NewReader(string(postData[:]))
	req, err := http.NewRequest("POST", url1, reqData)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)
	resp, err := httpUtil.HttpClient.Do(req)

	if err != nil {
		fmt.Printf("\n\nError : %s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}
	//debug(body, err)
	return fmt.Sprintf("%s", body)
}

func (httpUtil HttpUtil) postJSONData(apiEndPoint string, params interface{}) string {
	url1 := httpUtil.APIEndPoint + apiEndPoint
	postData, _ := json.Marshal(params)

	//fmt.Printf("%+vs", params)
	reqData := strings.NewReader(string(postData[:]))
	req, err := http.NewRequest("POST", url1, reqData)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	for _, kv := range httpUtil.Headers {
		req.Header.Add(kv.Key, kv.Value)
	}

	resp, err := httpUtil.HttpClient.Do(req)

	if err != nil {
		fmt.Printf("\n\nError : %s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}
	//debug(body, err)
	return fmt.Sprintf("%s", body)
}

func debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}
