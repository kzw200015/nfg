package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/robfig/cron/v3"

	"github.com/fsnotify/fsnotify"
)

const (
	defaultConfigFilePath = "nat.conf"
	tempNFTPath           = "/tmp/nft-diy.nft"
	helpMessage           = `配置文件格式：本地端口,远程端口,远程地址[,本地地址]
一行一条规则
例：
8081,22,example.com
8082,12345,example.com,192.168.1.6`
)

func do(config string) {
	rules, err := parseConfig(config)
	if err != nil {
		log.Println("配置文件解析失败：" + err.Error())
		return
	}
	saveNFTFile(rules)
	executeNFT()
}

func main() {
	config := flag.String("c", defaultConfigFilePath, "配置文件")
	help := flag.Bool("h", false, "显示帮助")

	flag.Parse()

	if *help {
		fmt.Println(helpMessage)
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	err = watcher.Add(*config)
	if err != nil {
		log.Fatal(err)
	}

	//添加定时任务
	c := cron.New()
	_, err = c.AddFunc("@every 1m", func() {
		log.Println("定时更新转发规则")
		do(*config)
	})
	if err != nil {
		log.Fatal(err)
	}

	c.Start()

	//启动时执行一遍规则
	do(*config)
	log.Println("等待配置文件变更...")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if (event.Op == fsnotify.Write) || (event.Op == fsnotify.Chmod) {
				log.Println("检测到配置文件变更")
				do(*config)
				log.Println("等待配置文件变更...")
			}
			//VIM修改文件会将原文件移除，导致无法监听，需要重新添加监听
			if event.Op == fsnotify.Remove {
				err = watcher.Add(*config)
				if err != nil {
					log.Fatal(err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println(err)
		}
	}
}
