package main

import (
	// 导入 chromedpclient 模块
	"WebFinder/cmd"
	"os"

	"github.com/coutcin-xw/go-logs"
)

func main() {
	// conf := config.GetConfig()
	logs.Log.SetColor(true)
	logs.Log.SetLevel(logs.Info)
	// 运行根命令
	if err := cmd.Execute(); err != nil {
		logs.Log.Error(err)
		os.Exit(1)
	}

}
