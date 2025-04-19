# 证书更新工具
## 说明

    1、更新长亭雷池WAF证书
    2、更新阿里云 OSS 域名绑定证书


## 支持服务器
    目前只编译了 linux 的版本，其他版本可以下载编译

#### 在指定目录创建.env文件
```shell
#服务器IP地址
BASE_SERVER_URL=https://127.0.0.1:9443
#长亭雷池WAF用户名
API_TOKEN=
#证书的 ID，可以通过页面开发者工具查看
CERT_ID=1


#阿里云 RAM 用户 AccessKeyId
ALIYUN_ACCESS_KEY_ID=
#阿里云 RAM 用户 AccessKeySecret
ALIYUN_ACCESS_SECRET=

#阿里云 OSS 配置项服务地址 
#详情参考：https://api.aliyun.com/product/Oss
ALIYUN_OSS_Endpoint=http://oss-cn-zhangjiakou.aliyuncs.com
#阿里云 OSS Bucket名称
ALIYUN_OSS_BUCKET_NAME=
#阿里云 OSS 配置项绑定域名 
ALIYUN_OSS_DOMAIN=


#新的证书路径
CERT_CRT_PATH=/live/cert/certificate.crt
#新的证书私钥路径
CERT_KEY_PATH=/live/cert/private.pem
```
#### 直接执行
```shell
./update_safelne

====================================
                证书同步工具
====================================

all：更新同步阿里云OSS证书和长亭雷池证书
aliyun：更新同步阿里云 OSS 证书
safeline：更新长亭雷池证书
不填写默认执行长亭雷池更新的更新操作

```
