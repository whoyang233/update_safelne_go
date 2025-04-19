package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"whoyang.cn/update_cert/client/aliyun"
	"whoyang.cn/update_cert/client/safeline"
)

var aliyunVersion = "v0.1"
var safelineVersion = "v0.2"

func main() {

	//加载.env文件
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("加载.env文件异常：", err)
		return
	}

	args := os.Args[1:]
	var updateType = ""
	if len(args) != 0 {
		updateType = args[0]
	}

	if updateType == "help" {
		fmt.Println("====================================")
		fmt.Println("\t\t证书同步工具\t\t")
		fmt.Println("====================================")
		fmt.Println("")
		fmt.Println("")
		fmt.Println("all：更新同步阿里云OSS证书和长亭雷池证书")
		fmt.Println("aliyun：更新同步阿里云 OSS 证书")
		fmt.Println("safeline：更新长亭雷池证书")
		fmt.Println("不填写默认执行长亭雷池更新的更新操作")
		return
	} else {
		if updateType == "all" || updateType == "aliyun" {
			aliyun.Run(aliyun.AccessConfig{
				KeyId:  os.Getenv("ALIYUN_ACCESS_KEY_ID"),
				Secret: os.Getenv("ALIYUN_ACCESS_SECRET"),
			}, aliyun.OssConfig{
				Endpoint:   os.Getenv("ALIYUN_OSS_Endpoint"),
				BucketName: os.Getenv("ALIYUN_OSS_BUCKET_NAME"),
				Domain:     os.Getenv("ALIYUN_OSS_DOMAIN"),
			}, aliyun.CertConfig{
				CrtPath: os.Getenv("CERT_CRT_PATH"),
				KeyPath: os.Getenv("CERT_KEY_PATH"),
			}, aliyunVersion)

			fmt.Println("")
			fmt.Println("阿里云 OSS 证书更新操作完成")
			fmt.Println("")
		}

		if updateType == "" || updateType == "all" || updateType == "safeline" {
			safeline.Run(safeline.ServerConfig{
				Url:      os.Getenv("BASE_SERVER_URL"),
				ApiToken: os.Getenv("API_TOKEN"),
				CertId:   os.Getenv("CERT_ID"),
			}, false, safelineVersion)

			fmt.Println("")
			fmt.Println("长亭雷池证书更新操作完成")
			fmt.Println("")
		}

	}

}
