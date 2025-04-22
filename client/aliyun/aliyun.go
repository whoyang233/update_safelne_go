package aliyun

import (
	"fmt"
	cas20200407 "github.com/alibabacloud-go/cas-20200407/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	credential "github.com/aliyun/credentials-go/credentials"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"whoyang.cn/update_cert/utils"
)

type AccessConfig struct {
	KeyId  string
	Secret string
}

type OssConfig struct {
	Endpoint   string
	BucketName string
	Domain     string
}

type CertConfig struct {
	CrtPath string
	KeyPath string
}

var accessKeyId string

var accessKeySecret string

// 创建OSS客户端
var endpoint string
var bucketName string
var domain string
var certCrtPath string
var certKeyPath string

var ossClient *oss.Client
var casClient *cas20200407.Client

func getCasClient() (casClient *cas20200407.Client) {
	credentialConfig := &credential.Config{
		Type: tea.String("access_key"),
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID。
		AccessKeyId: tea.String(accessKeyId),
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
		AccessKeySecret: tea.String(accessKeySecret),
	}

	credential, _err := credential.NewCredential(credentialConfig)
	if _err != nil {
		fmt.Println("获取 CAS 客户端发生异常：", _err)
		os.Exit(-1)
	}

	config := &openapi.Config{
		Credential: credential,
	}

	// Endpoint 请参考 https://api.aliyun.com/product/cas
	config.Endpoint = tea.String("cas.aliyuncs.com")

	casClient = &cas20200407.Client{}
	casClient, _err = cas20200407.NewClient(config)
	if _err != nil {
		fmt.Println("获取 CAS 客户端发生异常：", _err)
		os.Exit(-1)
	}
	return casClient
}

func getOssClient(endpoint string) (ossClient *oss.Client) {
	ossClient, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		fmt.Println("获取 OSS 客户端发生异常：", err)
		os.Exit(-1)
	}
	return ossClient
}

func listBucketCname(bucketName string) (certInfo any) {

	//获取 OOS 对应的映射域名及 SSL证书的相关信息
	bucketCname, err := ossClient.ListBucketCname(bucketName)
	if err != nil {
		fmt.Println("查看 Bucket 列表发生异常: ", bucketCname)
		os.Exit(-1)
	}

	// 打印存储空间信息
	for _, c := range bucketCname.Cname {
		if c.Domain == domain {
			if c.Certificate.CertId != "" {
				return c.Certificate
			} else {
				return oss.Certificate{}
			}
		}
	}

	return nil
}

func putBucketCert() bool {
	certId := uploadCert()

	//验证证书
	certIdStr, certExpired := getCertInfo(certId)

	if certExpired {
		fmt.Printf("%s 域名对应的证书过期，结束操作", domain)
		return false
	}

	bucketCnameConfig := oss.PutBucketCname{
		Cname: domain,
		CertificateConfiguration: &oss.CertificateConfiguration{
			CertId: certIdStr,
		},
	}
	_err := ossClient.PutBucketCnameWithCertificate(bucketName, bucketCnameConfig)
	if _err != nil {
		fmt.Println("Bucket绑定证书发生异常: ", _err)
		return false
	}
	return true
}

func deleteBucketCert() bool {
	bucketCnameConfig := oss.PutBucketCname{
		Cname: domain,
		CertificateConfiguration: &oss.CertificateConfiguration{
			DeleteCertificate: true,
		},
	}
	_err := ossClient.PutBucketCnameWithCertificate(bucketName, bucketCnameConfig)
	if _err != nil {
		fmt.Println("Bucket解除证书绑定发生异常: ", _err)
		return false
	}

	return true
}

func uploadCert() int64 {
	cert := utils.ReadFileOneLine(certCrtPath)
	key := utils.ReadFileOneLine(certKeyPath)

	certName := domain + "_" + time.Now().Format("200601021504")

	uploadUserCertificateRequest := &cas20200407.UploadUserCertificateRequest{
		Name: tea.String(certName),
		Cert: tea.String(cert),
		Key:  tea.String(key),
	}

	runtime := &util.RuntimeOptions{}
	uploadUserCertificeteResponse, err := casClient.UploadUserCertificateWithOptions(uploadUserCertificateRequest, runtime)
	if err != nil {
		var sdkError = &tea.SDKError{}
		sdkError = err.(*tea.SDKError)
		utils.ErrorLog("上传证书发生异常: ", tea.StringValue(sdkError.Message))
		os.Exit(-1)
	}
	body := uploadUserCertificeteResponse.Body
	return tea.Int64Value(body.CertId)
}

func getCertInfo(certId int64) (certIdStr string, certExpired bool) {
	//18173192
	//18151516-cn-hangzhou
	// 创建获取用户证书详情的请求
	getUserCertificateDetailRequest := &cas20200407.GetUserCertificateDetailRequest{
		CertId:     tea.Int64(certId),
		CertFilter: tea.Bool(true),
	}
	// 创建运行时选项
	runtime := &util.RuntimeOptions{}

	// 调用获取用户证书详情的接口
	getUserCertificateDetailResponse, err := casClient.GetUserCertificateDetailWithOptions(getUserCertificateDetailRequest, runtime)
	// 如果调用接口出错，则打印错误信息并退出程序
	if err != nil {
		var sdkError = &tea.SDKError{}
		sdkError = err.(*tea.SDKError)
		utils.ErrorLog("获取证书详情发生异常: ", tea.StringValue(sdkError.Message))
		os.Exit(-1)
	}

	// 获取用户证书详情的响应体
	userCertificateDetailBody := getUserCertificateDetailResponse.Body

	// 返回证书标识符和证书是否过期
	return tea.StringValue(userCertificateDetailBody.CertIdentifier), tea.BoolValue(userCertificateDetailBody.Expired)
}

func deleteCert(certId int64) bool {
	deleteUserCertificateRequest := &cas20200407.DeleteUserCertificateRequest{
		CertId: tea.Int64(certId),
	}
	runtime := &util.RuntimeOptions{}
	_, err := casClient.DeleteUserCertificateWithOptions(deleteUserCertificateRequest, runtime)
	if err != nil {
		var sdkError = &tea.SDKError{}
		sdkError = err.(*tea.SDKError)
		utils.ErrorLog("删除证书发生异常: ", tea.StringValue(sdkError.Message))
		return false
	}
	return true
}

func Run(access AccessConfig, config OssConfig, cert CertConfig, version string) {

	fmt.Println("====================================")
	fmt.Println("阿里云 OSS 证书同步工具 ", version)
	fmt.Println("====================================")
	fmt.Println("")

	accessKeyId = access.KeyId
	accessKeySecret = access.Secret

	endpoint = config.Endpoint
	bucketName = config.BucketName
	domain = config.Domain

	certCrtPath = cert.CrtPath
	certKeyPath = cert.KeyPath

	if accessKeyId == "" {
		fmt.Println("RAM用户AccessKeyID不能为空")
		return
	}
	if accessKeySecret == "" {
		fmt.Println("RAM用户AccessKey密钥不能为空")
		return
	}

	if endpoint == "" {
		fmt.Println("OSS配置项服务地址不能为空！详情参考：https://api.aliyun.com/product/Oss")
		return
	}
	if bucketName == "" {
		fmt.Println("OSS配置项Bucket名称不能为空")
		return
	}
	if domain == "" {
		fmt.Println("OSS配置项绑定域名不能为空！需要先手动再系统添加域名，详情查看：对象存储/Bucket 列表/XX/域名管理")
		return
	}

	if certCrtPath == "" {
		fmt.Println("域名对应的SSL证书公钥地址不能为空")
		return
	}
	if certKeyPath == "" {
		fmt.Println("域名对应的SSL证书私钥地址不能为空")
		return
	}

	ossClient = getOssClient(endpoint)
	casClient = getCasClient()

	// 列举存储空间中的域名
	bucketCertInfo := listBucketCname(bucketName)

	if bucketCertInfo == nil {
		fmt.Println("当前对象存储的Bucket未绑定域名，结束")
		return
	}

	// 如果没有绑定域名，则结束
	if !reflect.DeepEqual(bucketCertInfo, oss.Certificate{}) {
		fmt.Println("")
		fmt.Println("当前对象存储的Bucket已绑定域名，已添加证书，进行后续操作")
		fmt.Println("")
		fmt.Println("")
		certInfo := bucketCertInfo.(oss.Certificate)
		certIdStr := certInfo.CertId
		certId, _ := strconv.ParseInt(certIdStr[0:strings.Index(certIdStr, "-")], 10, 64)

		validEndDate, _ := time.Parse("Jan 02 15:04:05 2006 MST", certInfo.ValidEndDate)

		timeSub := validEndDate.Sub(time.Now())
		if timeSub.Hours() > 72 {
			fmt.Printf("%s 域名对应的证书还在有效期，暂不更新\n", domain)
		} else {
			fmt.Printf("%s 域名对应的证书将要过期，上传本地证书替换\n", domain)

			//删除存储空间与证书的绑定关系
			deleteBucketCert()

			//删除旧证书
			deleteCert(certId)

			//读取本地证书，上传并绑定证书
			if putBucketCert() {
				fmt.Printf("%s 域名对应的证书更新成功，结束操作\n", domain)
			}
		}
		return
	}
	fmt.Println("")
	fmt.Printf("%s 域名未添加证书，上传本地证书并绑定\n", domain)
	fmt.Println("")
	fmt.Println("")
	//读取本地证书，上传并绑定证书
	if putBucketCert() {
		fmt.Printf("%s 域名绑定证书成功，结束操作\n", domain)
	}
}
