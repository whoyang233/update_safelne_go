package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)
import netUrl "net/url"

var version = "v0.1"

// 是否启用 debug
var debugSwitch = false

// 全局服务 URL
var baseServerUrl string

// 登录后的 JWT
var baseJwt any

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
func getRequest(path string, method string, body any) (request http.Request, url *netUrl.URL) {
	url, _ = netUrl.Parse(baseServerUrl + path)
	request = http.Request{
		// POST请求方法
		Method: method,
		URL:    url,
		// http协议版本
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		// 请求头
		Header: map[string][]string{
			"Content-Type": {"application/json"}, // 使用json格式载荷体
		},
	}

	//拼接登录 jwt
	if baseJwt != nil {
		request.Header.Set("authorization", baseJwt.(string))
	}

	//拼接请求体
	if body != nil {
		request.Body = io.NopCloser(strings.NewReader(body.(string)))
	}

	return request, url
}

// debug 日志
func debugLog(v ...any) {
	if debugSwitch {
		log.Println(v)
	}
}

// 异常日志
func errorLog(v ...any) {
	log.Fatal(v)
}

// get 请求方法
func get(path string) (body string, statusCode int) {
	client := getClient()
	request, url := getRequest(path, "GET", nil)

	debugLog("url: ", url, " method: get")
	debugLog("url: ", url, " head: ", request.Header)
	resp, err := client.Do(&request)
	if err != nil {
		debugLog(err)
		return "", 500
	}
	bodyByte, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	body = string(bodyByte)
	debugLog("url: ", url, " responseBody: ", body)

	if resp.StatusCode != http.StatusOK {
		debugLog("Get请求失败，返回码为：", resp.StatusCode, "，异常返回体为：", body)
		return "", resp.StatusCode
	}

	if err != nil {
		debugLog(err)
		return "", 500
	}

	return body, resp.StatusCode
}

// post 请求方法
func post(path string, body any) string {
	client := getClient()

	request, url := getRequest(path, "POST", body)

	debugLog("url: ", url, " method: post", " request: ", body)
	debugLog("url: ", url, " head: ", request.Header)
	resp, err := client.Do(&request)
	if err != nil {
		debugLog(err)
		return ""
	}
	responseBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		debugLog("POST请求失败，返回码为：", resp.StatusCode, "，异常返回体为：", string(responseBody))
		return ""
	}
	if err != nil {
		debugLog(err)
	}
	debugLog("url: ", url, " responseBody: ", string(responseBody))
	return string(responseBody)
}

// 解析请求的返回体，过滤 err 信息，只返回 data 信息
func getResponseData(responseJson string) (data map[string]interface{}, dateErr any) {
	var responseInfo interface{}

	jsonErr := json.Unmarshal([]byte(responseJson), &responseInfo)
	if jsonErr != nil {
		debugLog(jsonErr)
		return nil, jsonErr
	}

	responseMap := responseInfo.(map[string]interface{})
	repDataMsg := responseMap["msg"]
	fmt.Println("repDataMsg: ", repDataMsg)
	if repDataMsg != nil {
		if repDataMsg != "" {
			debugLog(repDataMsg)
			return nil, repDataMsg.(string)
		}
	}

	repDataErr := responseMap["err"]
	if repDataErr != nil {
		debugLog(repDataErr)
		return nil, repDataErr.(string)
	}

	repData := responseMap["data"]
	if repData == nil {
		return nil, nil
	}
	dataMap := repData.(map[string]interface{})
	return dataMap, nil
}

// 读取文件
func readFile(filePath string) any {
	file, err := os.ReadFile(filePath)
	if err != nil {
		errorLog("读取文件异常：", err)
		return nil
	}
	return string(file)
}

// 1、获取 csrf_token
func getCsrfToken() string {
	csrfTokenJson, statusCode := get("/api/open/auth/csrf")
	if statusCode != 200 {
		errorLog("获取获取csrf_token接口调用异常")
		return ""
	}
	dataMap, _ := getResponseData(csrfTokenJson)
	return dataMap["csrf_token"].(string)
}

// 2、登录
func login(csrfToken string, loginUser string, loginPassword string) {

	loginRequest := struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		CsrfToken string `json:"csrf_token"`
	}{loginUser, loginPassword, csrfToken}

	loginRequestJson, jsonErr := json.Marshal(&loginRequest)
	if jsonErr != nil {
		errorLog("登录接口JSON解析异常：", jsonErr)
		return
	}

	responseJson := post("/api/open/auth/login", string(loginRequestJson))
	data, dateErr := getResponseData(responseJson)
	if dateErr != nil {
		errorLog("登录接口调用失败：", dateErr)
	}
	jwt := data["jwt"].(string)
	baseJwt = "Bearer " + jwt
}

// 证书对象
type CertInfo struct {
	Id          int    `json:"id"`
	Domain      string `json:"domain"`
	Issuer      string `json:"issuer"`
	ValidBefore string `json:"valid_before"`
}

// 3、获取证书列表，并返回 map 对象
func certList() (certInfoMap map[int]CertInfo) {
	certInfoMap = make(map[int]CertInfo)
	responseJson, statusCode := get("/api/open/cert")
	if statusCode != 200 {
		errorLog("获取证书列表接口调用异常")
		return nil
	}
	data, _ := getResponseData(responseJson)
	nodes := data["nodes"].([]interface{})
	for i := range nodes {
		node := nodes[i].(map[string]interface{})
		id := int(node["id"].(float64))
		var domain string
		domains := node["domains"].([]interface{})
		for j := range domains {
			domain += domains[j].(string)
		}
		issuer := node["issuer"].(string)

		validBeforeTime, _ := time.Parse(time.RFC3339, node["valid_before"].(string))
		validBefore := validBeforeTime.Format("2006-01-02 15:04:05")
		certInfo := CertInfo{
			id, domain, issuer, validBefore,
		}
		certInfoMap[certInfo.Id] = certInfo
	}
	return certInfoMap
}

// 4、更新证书
func certUpdate(certId int, baseCertCrtPath string, baseCertKeyPath string) bool {
	domain, _, validBefore, crt := getCertInfo(certId)
	if domain == "" && crt == "" {
		return false
	}

	validBeforeTime, _ := time.Parse("2006-01-02 15:04:05", validBefore)
	timeSub := validBeforeTime.Sub(time.Now())
	if timeSub.Hours() > 72 {
		errorLog("证书还在有效期，暂不更新")
	}

	type manual struct {
		Crt string `json:"crt"`
		Key string `json:"key"`
	}

	certCrt := readFile(baseCertCrtPath)
	if certCrt == nil {
		errorLog("读取证书信息异常")
		return false
	}

	certKey := readFile(baseCertKeyPath)
	if certKey == nil {
		errorLog("读取证书私钥异常")
		return false
	}

	manualInfo := manual{
		certCrt.(string),
		certKey.(string),
	}

	certInfo := struct {
		Id     int    `json:"id"`
		Type   int    `json:"type"`
		Manual manual `json:"manual"`
	}{
		certId,
		2,
		manualInfo,
	}

	certUpdateRequestJson, jsonErr := json.Marshal(&certInfo)
	if jsonErr != nil {
		errorLog("更新证书接口JSON解析异常：", jsonErr)
		return false
	}

	responseJson := post("/api/open/cert", string(certUpdateRequestJson))
	getResponseData(responseJson)
	return true
}

// 5、获取证书详情
func getCert(certId int) (domain string, crt string) {

	getCertJson, statusCode := get(fmt.Sprint("/api/open/cert/", certId))
	if statusCode != 200 {
		errorLog("获取证书详情接口调用异常，可能是证书ID不存在")
		return "", ""
	}
	data, _ := getResponseData(getCertJson)
	fmt.Println(data)
	acme := data["acme"].(map[string]interface{})
	domains := acme["domains"].([]interface{})
	for i := range domains {
		domain += domains[i].(string)
	}
	manual := data["manual"].(map[string]interface{})
	crt = manual["crt"].(string)
	return domain, crt
}

// 获取证书详情，通过列表和详情接口拼接而成
func getCertInfo(certId int) (domain string, issuer string, validBefore string, crt string) {
	domain, crt = getCert(certId)
	certInfoMap := certList()
	certInfo := certInfoMap[certId]
	issuer = certInfo.Issuer
	validBefore = certInfo.ValidBefore
	return domain, issuer, validBefore, crt
}

func main() {
	fmt.Println("====================================")
	fmt.Println("长亭雷池WAF站点证书同步工具 ", version)
	fmt.Println("====================================")
	fmt.Println("")
	//加载.env文件
	err := godotenv.Load(".env")
	if err != nil {
		errorLog("加载.env文件异常", err)
	}

	//加载环境变量的参数
	debugSwitch, _ = strconv.ParseBool(os.Getenv("DEBUG"))

	baseServerUrl = os.Getenv("BASE_SERVER_URL")
	if baseServerUrl == "" {
		baseServerUrl = "https://127.0.0.1:9443"
	}
	loginUser := os.Getenv("LOGIN_USER")
	if loginUser == "" {
		loginUser = "admin"
	}
	loginPassword := os.Getenv("LOGIN_PASSWORD")
	if loginPassword == "" {
		fmt.Println("长亭雷池WAF登录密码不能为空")
		return
	}
	certIdSrt := os.Getenv("CERT_ID")
	if certIdSrt == "" {
		certIdSrt = "1"
	}
	certId, _ := strconv.Atoi(certIdSrt)
	certCrtPath := os.Getenv("CERT_CRT_PATH")
	if certCrtPath == "" {
		fmt.Println("长亭雷池WAF站点新证书地址不能为空")
		return
	}
	certKeyPath := os.Getenv("CERT_KEY_PATH")
	if certKeyPath == "" {
		fmt.Println("长亭雷池WAF站点新证书私钥地址不能为空")
		return
	}

	//调用接口
	csrfToken := getCsrfToken()
	login(csrfToken, loginUser, loginPassword)

	updateResult := certUpdate(certId, certCrtPath, certKeyPath)
	if updateResult == true {
		domain, issuer, validBefore, crt := getCertInfo(certId)
		//展示最终结果
		fmt.Println("长亭雷池WAF站点证书同步成功，同步内容如下：\n",
			"域名：", domain, "\r\n",
			"颁发机构：", issuer, "\r\n",
			"有效期至：", validBefore, "\r\n",
			"证书信息：", crt)
	} else {
		fmt.Print("长亭雷池WAF站点证书同步失败，请核查日志")
	}
}
