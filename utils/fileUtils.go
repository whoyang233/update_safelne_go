package utils

import (
	"fmt"
	"os"
	"strings"
)

// 读取文件函数，参数为文件路径，返回值为any类型
func ReadFile(filePath string) string {
	// 使用os包的ReadFile函数读取文件内容，返回值为[]byte类型
	file, err := os.ReadFile(filePath)
	// 如果读取文件出现错误，打印错误信息并返回nil
	if err != nil {
		fmt.Print("读取文件异常：", err)
		return ""
	}
	// 将读取到的文件内容转换为string类型并返回
	return string(file)
}

// 读取文件并返回文件内容
func ReadFileOneLine(filePath string) string {
	// 读取文件内容
	file, err := os.ReadFile(filePath)
	// 如果读取文件出现异常，打印异常信息并返回nil
	if err != nil {
		fmt.Print("读取文件异常：", err)
		return ""
	}
	// 将文件内容转换为字符串
	fileBody := string(file)
	// 将文件内容中的回车符和换行符替换为空字符串
	fileBody = strings.ReplaceAll(fileBody, "\r", "")
	fileBody = strings.ReplaceAll(fileBody, "\n", "")
	// 返回文件内容
	return fileBody
}
