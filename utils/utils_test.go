package utils

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/chainreactors/fingers"
	"github.com/chainreactors/utils/httputils"
)

func TestFavicon(t *testing.T) {
	engine, err := fingers.NewEngine()
	if err != nil {
		panic(err)
	}
	resp, err := http.Get("http://baidu.com/favicon.ico")
	if err != nil {
		return
	}
	content := httputils.ReadRaw(resp)
	body, _, _ := httputils.SplitHttpRaw(content)
	frame := engine.DetectFavicon(body)
	fmt.Println(frame)
}
