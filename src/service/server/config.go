package server

import "fmt"

// 结构体Server
type config_server struct {
	url    string // 配置服务器地址
	repo   string // 配置的仓库名
	branch string // 配置分支名称
	token  string // token
}

func New(url string, repo string, branch string, token string) config_server {
	return config_server{url, repo, branch, token}
}

func (server config_server) Tostring() string {
	return fmt.Sprintf("Begin Download From Server url:%s ,repo:%s ,branch:%s ,token:%s", server.url, server.repo, server.branch, "*****")
}
