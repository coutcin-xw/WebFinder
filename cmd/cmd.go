package cmd

import (
	"WebFinder/chromedpclient"
	"fmt"
	"time"

	"os"

	"github.com/coutcin-xw/go-logs"
	"github.com/spf13/cobra"
)

var (
	debugMode  bool
	targetURL  string
	timeout    int
	chromePath string
	configFile string
)

// 通过ldflags传入构建时间
var buildTime string
var description = fmt.Sprintf(`
__          __     _     _____  _            _
\ \        / /    | |   |  ___|(_)          | |
 \ \  /\  / / ___ | |__ | |__   _  _ __   __| | ___  _ __
  \ \/  \/ / / _ \| '_ \|  __| | || '_ \ / _' |/ _ \| '__|
   \  /\  / |  __/| |_) | |    | || | | | (_| |  __/| |
    \/  \/   \___||_.__/|_|    |_||_| |_|\__,_|\___||_|

    Build Time: %s 
    WebFinder 是一款指纹加URL提取工具。`, buildTime)
var rootCmd = &cobra.Command{
	Use:   "WebFinder",
	Short: "WebFinder 是一款指纹加URL提取工具。",
	Long:  description,
	Run: func(cmd *cobra.Command, args []string) {
		// 这个命令本身没有执行什么操作，默认情况下会显示帮助信息
		if !cmd.HasSubCommands() && cmd.Flags().NFlag() == 0 && len(args) == 0 {
			cmd.Help()
			logs.Log.Error("您未输入任何参数")
			return
		}
		logs.Log.Console("正在运行 WebFinder 工具...\n")
		runWebFinder()
	},
}

// Execute 运行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 这里可以添加一些全局参数配置
	// 注册全局参数
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "指定配置文件 (默认路径为 $HOME/.webfinder.yaml)")
	rootCmd.PersistentFlags().StringVarP(&targetURL, "url", "u", "", "目标地址")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 30, "请求超时时间 (秒)")
	rootCmd.PersistentFlags().StringVarP(&chromePath, "chrome-path", "P", "", "指定 Chrome 浏览器的路径")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "是否开启 Debug 模式")
	// 初始化操作，比如加载配置文件
	cobra.OnInitialize(func() {
		// initializeConfig()
	})
}

// 初始化配置文件逻辑
func initializeConfig() {
	if configFile != "" {
		if _, err := os.Stat(configFile); err == nil {
			fmt.Printf("加载配置文件: %s\n", configFile)
			// 加载配置逻辑，比如解析 YAML 或 JSON 文件
		} else {
			fmt.Printf("未找到配置文件: %s，使用默认配置\n", configFile)
		}
	}

	// 如果开启了 debug 模式，输出调试信息
	if debugMode {
		logs.Log.Debug("Debug 模式已启用")
		logs.Log.Debugf("目标地址: %s", targetURL)
		logs.Log.Debugf("超时时间: %d 秒", timeout)
		logs.Log.Debugf("Chrome 路径: %s", chromePath)
	}
}

// 主功能逻辑
func runWebFinder() {
	if debugMode {
		logs.Log.SetLevel(logs.Debug)
	}

	// 检查 URL 参数是否提供
	if targetURL == "" {
		fmt.Println("错误: 必须提供目标地址 --url")
		os.Exit(1)
	}
	// conf.ChromePath = *chrome_path
	// chromedpclient.InitChromedp()
	// // 调用封装好的 chromedp 任务
	pageTitle, _, err := chromedpclient.RunChromedpTask(targetURL, time.Duration(timeout)*time.Second)
	if err != nil {
		logs.Log.Errorf("%v", err) // 错误检查
		return
	}
	logs.Log.Infof("页面标题：%s", pageTitle)
}
