package chromedpclient

import (
	"WebFinder/config"
	"WebFinder/mode"
	"WebFinder/result"
	"WebFinder/utils"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/coutcin-xw/go-logs"
	"github.com/coutcin-xw/goutils/nettools"
)

var opts []func(*chromedp.ExecAllocator)
var TaskCtx context.Context
var log = logs.Log
var requestMap sync.Map

// 创建一个 WaitGroup 来等待所有 goroutine 完成
var wg sync.WaitGroup

// RunChromedpTask 运行 chromedp 任务，获取页面渲染后的 HTML 内容并写入文件
func RunChromedpTask(url string, timeout time.Duration) (string, string, error) {
	// 自定义 Chrome 启动选项
	opts = append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,                              // 禁用 GPU 加速
		chromedp.NoDefaultBrowserCheck,                   // 禁用默认的浏览器检查
		chromedp.Flag("headless", true),                  // 启动非无头模式 (显示浏览器窗口)
		chromedp.Flag("ignore-certificate-errors", true), // 忽略证书错误
		chromedp.Flag("window-size", "50,400"),           // 设置浏览器窗口大小
	)
	var cancel context.CancelFunc
	// 创建任务上下文并配置日志输出
	// 创建浏览器分配器 (Allocator)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 使用分配器创建任务上下文，并设置日志输出
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Debugf))
	defer cancel()

	// 设置任务超时时间
	TaskCtx, cancel = context.WithTimeout(taskCtx, 60*time.Second)
	defer cancel()
	listenForNetworkEvent(TaskCtx)
	// 执行任务
	var pageTitle string
	var pageHTML string
	if err := chromedp.Run(TaskCtx,
		network.Enable(),
		chromedp.Navigate(url),  // 导航到指定的 URL
		chromedp.Sleep(timeout), // 等待页面渲染完成
		chromedp.OuterHTML("html", &pageHTML),
		chromedp.Title(&pageTitle), // 获取页面标题
	); err != nil {
		return "", "", err
	}
	return pageTitle, pageHTML, nil
}

// 监听网络事件
func listenForNetworkEvent(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			requestMap.Store(ev.RequestID, ev.Request)
		case *network.EventResponseReceived:
			log.Debugf("Response for RequestID: %s", ev.RequestID)

			// 如果上下文已经被取消，返回
			if ctx.Err() != nil {
				log.Errorf("Context error: %v", ctx.Err())
				return
			}
			// 增加 WaitGroup 计数
			wg.Add(1)
			// 异步处理响应体
			go func(ev *network.EventResponseReceived) {
				defer wg.Done() // 确保完成时调用 Done()

				// 使用 chromedp.Run 确保获取响应体
				if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					// 获取响应内容
					data, err := network.GetResponseBody(ev.RequestID).Do(ctx)
					if err != nil {
						log.Errorf("Failed to get response body for RequestID: %s, error: %v", ev.RequestID, err)
						if req, ok := requestMap.Load(ev.RequestID); ok {
							tmpt, err := convertToHTTPRequest(ctx, req.(*network.Request), ev.RequestID)
							if err != nil {
								log.Errorf("Failed to convert to http.Request: %v", err)
							} else {
								requset, _ := nettools.RemoveQueryParams(tmpt.URL.String())
								log.Errorf("HTTP Request: %s", requset)
							}
						}
						return nil
					}

					// 提取请求信息并转换为 http.Request
					var httpReq *http.Request
					if req, ok := requestMap.Load(ev.RequestID); ok {
						httpReq, err = convertToHTTPRequest(ctx, req.(*network.Request), ev.RequestID)
						if err != nil {
							log.Errorf("Failed to convert to http.Request: %v", err)
						} else {
							requset, _ := nettools.ReadRequest(httpReq, true)
							log.Debugf("HTTP Request: \n%s", requset)
						}
					}

					// 转换响应体为 http.Response
					log.Debugf("RequestURL:%s:PRo: %s", ev.Response.URL, ev.Response.Protocol)
					httpResp, err := convertToHTTPResponse(ev.Response, data, httpReq)
					if err != nil {
						log.Errorf("Failed to convert to http.Response: %v", err)
					} else {
						respx, _ := nettools.ReadResponse(httpResp, true)
						log.Debugf("HTTP Response: \n%s", respx)
						tmp_url, _ := nettools.RemoveQueryParams(ev.Response.URL)
						if utils.IsInMatchs(tmp_url, config.Config.JsFind) {
							result.JsLinks = append(result.JsLinks, mode.Link{Url: tmp_url, Fingers: utils.FingersEngine3(respx)})
							log.Infof("%s指纹：%s", tmp_url, utils.FingersEngine3(respx))
						} else {
							log.Debugf("其他uri：%s", tmp_url)
						}
					}

					return nil
				})); err != nil {
					log.Errorf("Error during chromedp.Run: %v", err)
				}
			}(ev)

		}
	})
}

// 将 chromedp 的网络请求转换为 http.Request
func convertToHTTPRequest(ctx context.Context, req *network.Request, requestID network.RequestID) (*http.Request, error) {
	httpReq, err := http.NewRequest(req.Method, req.URL, nil)
	if err != nil {
		return nil, err
	}

	// 设置头信息
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value.(string))
	}

	postData, err := network.GetRequestPostData(requestID).Do(ctx) // 使用事件中的 RequestID
	if err != nil {
		httpReq.Body = nil
		// log.Warnf("Failed to get http.Response.body: %v", err)
		return httpReq, nil
	} else {
		httpReq.Body = io.NopCloser(bytes.NewReader([]byte(postData)))
	}

	return httpReq, nil
}

func convertToHTTPResponse(resp *network.Response, body []byte, req *http.Request) (*http.Response, error) {

	// 将 CDP 的响应头转为标准 HTTP Header
	headers := make(http.Header)
	for k, v := range resp.Headers {
		headers.Set(k, fmt.Sprintf("%v", v))
	}

	httpResp := &http.Response{
		Status:        resp.StatusText,
		StatusCode:    int(resp.Status),
		Proto:         strings.ToUpper(toProto(resp.Protocol)),
		Header:        make(http.Header),
		ContentLength: int64(len(body)),
		Body:          io.NopCloser(bytes.NewReader(body)),
		Request:       req,
	}

	// 设置响应头
	for key, value := range resp.Headers {
		httpResp.Header.Set(key, value.(string))
	}
	return httpResp, nil
}

func toProto(protocol string) string {
	// 检查输入是否为空
	if strings.TrimSpace(protocol) == "" {
		return ""
	}

	// 支持的协议别名映射
	aliases := map[string]string{
		"h2": "http/2.0",
		"h3": "http/3.0",
	}

	// 替换别名为标准形式
	if standard, exists := aliases[protocol]; exists {
		protocol = standard
	}
	return protocol
}
