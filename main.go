package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
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

type Rule struct {
	LocalPort  string
	RemotePort string
	RemoteAddr string
	LocalAddr  string
}

func saveNFTFile(rules []Rule) {
	log.Println("generate nftables rules")
	err := ioutil.WriteFile(tempNFTPath, []byte(generateNFT(rules)), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func executeNFT() {
	log.Println("execute nftables")
	cmd := exec.Command("nft", "-f", tempNFTPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

func do(config string) {
	rules, err := parseConfig(config)
	if err != nil {
		log.Println("config parse error:" + err.Error())
		return
	}
	saveNFTFile(rules)
	executeNFT()
}

func main() {
	config := flag.String("c", defaultConfigFilePath, "config file")
	help := flag.Bool("h", false, "show help")

	flag.Parse()

	if *help {
		fmt.Println(helpMessage)
		return
	}

	do(*config)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if (event.Op == fsnotify.Write) || (event.Op == fsnotify.Chmod) {
					do(*config)
					log.Println("waiting")
				}
				//VIM修改文件会将原文件移除，导致无法监听，需要重现添加监听
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
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(*config)
	if err != nil {
		log.Fatal(err)
	}

	//添加定时任务
	c := cron.New()
	_, err = c.AddFunc("r@every 1m", func() {
		do(*config)
	})
	if err != nil {
		log.Fatal(err)
	}

	c.Start()

	log.Println("waiting")
	<-done
}
