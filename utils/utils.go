// utils.go
package utils

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/coutcin-xw/go-logs"
	"github.com/coutcin-xw/goutils/nettools"

	"github.com/chainreactors/fingers"
	"github.com/spaolacci/murmur3"
)

var log = logs.Log

// 计算数据流的 MD5 哈希值
func CalculateMD5FromBytes(bytes []byte) (string, error) {
	// 创建一个 MD5 hash 计算器
	hash := md5.New()

	hash.Write(bytes)

	// 获取计算出的 MD5 值并转换为 16 进制字符串
	md5Hash := fmt.Sprintf("%x", hash.Sum(nil))
	return md5Hash, nil
}

// 计算数据流的 MMH3 哈希值
func CalculateMMH3FromBytes(bytes []byte) (uint32, error) {
	// 创建一个 MMH3 hash 计算器
	hasher := murmur3.New32()

	// 将数据流传给 MMH3 计算器
	hasher.Write(bytes)

	// 获取计算出的 MMH3 值
	hash := hasher.Sum32()
	return hash, nil
}

// 计算文件的 MD5 哈希值
func CalculateMD5FromFile(filePath string) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建一个 MD5 hash 计算器
	hash := md5.New()

	// 将文件内容传给 MD5 计算器
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// 获取计算出的 MD5 值并转换为 16 进制字符串
	md5Hash := fmt.Sprintf("%x", hash.Sum(nil))
	return md5Hash, nil
}

// 计算文件的 MMH3 哈希值
func CalculateMMH3FromFile(filePath string) (uint32, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// 创建一个 MMH3 hash 计算器
	hasher := murmur3.New32()

	// 将文件内容传给 MMH3 计算器
	if _, err := io.Copy(hasher, file); err != nil {
		return 0, err
	}

	// 转换为十六进制字符串
	hash := hasher.Sum32()
	return hash, nil
}

// CalculateFavicon 计算 favicon 的 MD5 和 MMH3 哈希
func CalculateFavicon(url string) (string, string) {
	// 创建一个自定义的 HTTP 客户端
	client := &http.Client{
		Transport: &http.Transport{
			// 配置跳过证书验证
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 忽略证书验证
			},
		},
	}
	// 发送 HTTP 请求
	resp, err := client.Get(url)
	if err != nil {
		log.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 将响应体读取到缓冲区
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, resp.Body)
	if err != nil {
		log.Errorf("读取响应体失败: %v", err)
	}

	// 计算 MD5 哈希
	md5Hash, err := CalculateMD5FromBytes(buffer.Bytes())
	if err != nil {
		log.Errorf("计算 MD5 失败: %v", err)
	}
	log.Debugf("MD5 哈希值: %s", md5Hash)

	mmh3Hash, err := CalculateMMH3FromBytes(buffer.Bytes())
	if err != nil {
		log.Errorf("计算 MMH3 失败: %v", err)
	}

	log.Debugf("MMH3 哈希值: %x", mmh3Hash)
	mh, err := CalculateMMH3FromBytes(StandBase64(buffer.Bytes()))
	log.Debugf("fofa哈希值: %d", int32(mh))
	if err != nil {
		log.Errorf("转换十六进制字符串失败: %v", err)
	}

	return md5Hash, fmt.Sprintf("%x", mmh3Hash)
}

// writeHTMLToFile 将 HTML 内容写入到指定的文件
func WriteHTMLToFile(filePath, content string) error {
	// 打开或创建文件，如果文件存在则覆盖
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将内容写入文件
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	log.Debugf("页面内容已成功写入到文件: %s", filePath)
	return nil
}

func WriteURIToFile(filePath, content string) error {
	// 打开文件，如果文件不存在则创建，如果文件存在则追加
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将内容写入文件
	_, err = file.WriteString(content + "\n")
	if err != nil {
		return err
	}

	log.Debugf("写入URL: %s", filePath)
	return nil
}

// StandBase64 计算 base64 的值
func StandBase64(braw []byte) []byte {
	bckd := base64.StdEncoding.EncodeToString(braw)
	var buffer bytes.Buffer
	for i := 0; i < len(bckd); i++ {
		ch := bckd[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')
	return buffer.Bytes()
}

// SplitChar76 按照 76 字符切分
func SplitChar76(braw []byte) []byte {
	// 去掉 data:image/vnd.microsoft.icon;base64
	if strings.HasPrefix(string(braw), "data:image/vnd.microsoft.icon;base64,") {
		braw = braw[37:]
	}

	var buffer bytes.Buffer
	for i := 0; i < len(braw); i++ {
		ch := braw[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')

	return buffer.Bytes()
}

func FingersEngine(url string, body []byte) []string {
	engine, err := fingers.NewEngine()
	if err != nil {
		panic(err)
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("错误: %v", err)
		return []string{}
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	content, _ := nettools.ReadResponse(resp, false)
	frames, err := engine.DetectContent(content)
	if err != nil {
		log.Errorf("错误: %v", err)
		return []string{}
	}
	return FingersFormat(frames.String())
}
func FingersEngine2(urls string) string {
	engine, err := fingers.NewEngine()
	if err != nil {
		panic(err)
	}
	resp, err := http.Get(urls)
	if err != nil {
		log.Errorf("错误: %v", err)
		return ""
	}

	content, _ := nettools.ReadResponse(resp, false)

	frames, err := engine.DetectContent(content)
	if err != nil {
		log.Errorf("错误: %v", err)
		return ""
	}
	return frames.String()
}

func FingersFaviconEngine(resp *http.Response) []string {
	engine, err := fingers.NewEngine()
	if err != nil {
		panic(err)
	}

	body, _ := nettools.ReadResponseBody(resp)
	frames := engine.DetectFavicon(body)
	if frames == nil {
		log.Errorf("错误: %s", "frames 为空")
		return []string{}
	}
	return FingersFormat(frames.String())
}
func FingersEngine3(content []byte) []string {
	engine, err := fingers.NewEngine()
	if err != nil {
		panic(err)
	}
	frames, err := engine.DetectContent(content)
	if err != nil && frames == nil {
		log.Errorf("错误: %s%v", content, err)
		return nil
	}

	return FingersFormat(frames.String())
}

// FingersFormat 提取并格式化指纹信息
func FingersFormat(content string) []string {
	// 结果切片，用来保存指纹信息
	var result []string

	// 按 '||' 分割字符串
	items := strings.Split(content, "||")

	// 遍历每个项
	for _, item := range items {
		// 按 ':' 分割每个项
		parts := strings.Split(item, ":")
		if len(parts) > 1 {
			// 提取指纹部分（即 ':' 前的内容）
			fingerprint := parts[0]
			// 将指纹加入结果列表
			result = append(result, fingerprint)
		}
	}

	return result
}

func PrintfingersByBytes(bytes []byte) {
	engine, err := fingers.NewEngine()
	if err != nil {
		panic(err)
	}
	frames, err := engine.DetectContent(bytes)
	if err != nil {
		log.Errorf("错误: %v", err)
		return
	}
	fmt.Println("识别到指纹：" + frames.String())
}

func IsInMatchs(s string, patterns []string) bool {
	for _, pattern := range patterns {
		if match, _ := regexp.MatchString(pattern, s); match {
			return true
		}
	}
	return false
}

func IsInContains(s string, patterns []string) bool {
	for _, pattern := range patterns {
		if match := strings.Contains(s, pattern); match {
			return true
		}
	}
	return false
}
