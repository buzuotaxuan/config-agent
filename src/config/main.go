package main

import (
	"service/server"
	"log"
	"flag"
)

// 程序入口
func main() {
	// 启动参数配置
	url := flag.String("url", "http://autozi.git.lipg.cn", "config server url")
	repo := flag.String("repo", "sps/service-ci", "config server repo; ")
	branch := flag.String("branch", "master", "config server repo; ")
	path := flag.String("path", "/tmp/config/config.tar", "config server download path")
	token := flag.String("token", "rC7xuKSTtACqQKfwb88m", "config server token")
	flag.Parse()

	log.Println("开始获取项目配置信息.repo:" + *repo)
	repo_server := server.New(*url, *repo, *branch, *token)
	log.Println(repo_server.Tostring())
	repo_server.Download(*path)
	log.Print("Config Is Downloaded")
}
