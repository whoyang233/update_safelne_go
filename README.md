# 长亭雷池WAF证书更新
## 说明
    通过调用接口的方式，模拟用户操作更新指定站点的证书信息

## 用法
    目前只编译了 linux 的版本
#### 在指定目录创建.env文件
```shell
#服务器IP地址
BASE_SERVER_URL=https://127.0.0.1:9443
#长亭雷池WAF用户名
API_TOKEN=
#证书的 ID，可以通过页面开发者工具查看
CERT_ID=1
#新的证书路径
CERT_CRT_PATH=/live/cert/certificate.crt
#新的证书私钥路径
CERT_KEY_PATH=/live/cert/private.pem
```
#### 直接执行
```shell
./update_safelne
```
## 调用的接口
    3、获取证书列表
    4、更新证书
    5、获取证书详情