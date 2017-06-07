package server

import (
	"net/http"
	"os"
	"io"
	"fmt"
	"errors"
	"crypto/tls"
)

func (server config_server) Download(path string) {
	// 创建client，用以进行请求
	client :=http.DefaultClient
	// 创建下载请求
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s/repository/archive.tar?ref=%s", server.url, server.repo, server.branch), nil)
	// 注入basic认证
	req.Header.Add("PRIVATE-TOKEN", server.token)
	// 信任任何证书
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client.Transport=tr
	// 执行访问
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if response.StatusCode != 200 {
		panic(errors.New(fmt.Sprintf("download config error, error:%s", response.Status)))
	}
	// 创建目录，并删除文件
	os.MkdirAll(path, os.ModePerm)
	os.Remove(path)
	// 创建文件
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	io.Copy(f, response.Body)
}
