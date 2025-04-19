package utils

import (
	"crypto/tls"
	"io"
	"net/http"
	netUrl "net/url"
	"strings"
)

// 获取 client 客户端
func getClient() *http.Client {
	//跳过不安全的验证
	transport := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := http.Client{Transport: &transport}
	return &client
}

// 获取请求体
func getRequest(url string, method string, header map[string][]string, body any) (request http.Request) {
	rawURL, _ := netUrl.Parse(url)
	request = http.Request{
		// POST请求方法
		Method: method,
		URL:    rawURL,
		// http协议版本
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	if header != nil {
		request.Header = header
	}

	//拼接请求体
	if body != nil {
		request.Body = io.NopCloser(strings.NewReader(body.(string)))
	}

	return request
}

func get(url string) (any, int) {
	client := getClient()
	request := getRequest(url, "GET", nil, nil)

	response, err := client.Do(&request)
	if err != nil {
		TraceLog("", err)
		return "", 500
	}

	bodyByte, err := io.ReadAll(response.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(response.Body)

	responseBody := string(bodyByte)

	if response.StatusCode != http.StatusOK {
		ErrorLog("Get请求失败，返回码为：", response.StatusCode, "，异常返回体为：", responseBody)
		return "", response.StatusCode
	}
	return responseBody, response.StatusCode
}

func post(url string, body any) (any, int) {
	client := getClient()
	var headers = map[string][]string{
		"Content-Type": {"application/json"},
	}

	request := getRequest(url, "POST", headers, body)

	response, err := client.Do(&request)
	if err != nil {
		TraceLog("", err)
		return "", 500
	}

	bodyByte, err := io.ReadAll(response.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(response.Body)

	responseBody := string(bodyByte)

	if response.StatusCode != http.StatusOK {
		ErrorLog("Post请求失败，返回码为：", response.StatusCode, "，异常返回体为：", responseBody)
		return "", response.StatusCode
	}
	return responseBody, response.StatusCode
}
