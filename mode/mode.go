package mode

import "time"

// Config 是整个配置文件的结构体
type Config struct {
	Timeout    time.Duration       `yaml:"timeout"`
	ChromePath string              `yaml:"chrome_path"`
	LogLevel   int                 `yaml:"log_level"`
	JsFind     []string            `yaml:"jsFind"`
	UrlFind    []string            `yaml:"urlFind"`
	InfoFind   map[string][]string `yaml:"infoFiler"`
	Risks      []string            `yaml:"risks"`
	JsFiler    []string            `yaml:"jsFiler"`
	UrlFiler   []string            `yaml:"urlFiler"`
	JsFuzzPath []string            `yaml:"jsFuzzPath"`
}

type Link struct {
	Url      string   //url
	Status   string   //状态码
	Size     string   //返回包大小
	Title    string   //标题
	Redirect string   //重定向
	Source   string   //来源
	Fingers  []string //指纹
}
