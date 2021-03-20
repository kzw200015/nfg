package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	nftPrefix = `#!/usr/sbin/nft -f

add table ip nat
delete table ip nat
add table ip nat
add chain nat PREROUTING { type nat hook prerouting priority -100 ; }
add chain nat POSTROUTING { type nat hook postrouting priority 100 ; }
`
	nftFormat = `add rule ip nat PREROUTING %v dport %v counter dnat to %v:%v
add rule ip nat POSTROUTING ip daddr %v %v dport %v counter snat to %v`
)

func parseConfig(filePath string) ([]Rule, error) {
	var rules []Rule

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	defaultLocalAddr, err := getLocalAddr()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var localAddr string

		ruleArr := strings.Split(scanner.Text(), ",")
		//解析失败跳过当前循环
		if len(ruleArr) < 3 {
			continue
		}
		//判断是否自定义了localAddr
		if len(ruleArr) == 4 {
			localAddr = ruleArr[3]
		} else {
			localAddr = defaultLocalAddr
		}

		remoteAddr, err := resolveDomain(ruleArr[2])
		if err != nil {
			return nil, err
		}

		rules = append(rules, Rule{
			LocalPort:  ruleArr[0],
			RemotePort: ruleArr[1],
			RemoteAddr: remoteAddr,
			LocalAddr:  localAddr,
		})

	}

	return rules, nil
}

func getLocalAddr() (string, error) {
	conn, err := net.Dial("tcp", "8.8.8.8:53")
	if err != nil {
		return "", err
	}
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}

func resolveDomain(domain string) (string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}

	return ips[0].String(), nil
}

func generateNFT(rules []Rule) string {
	var nft strings.Builder
	nft.WriteString(nftPrefix)
	nft.WriteString("\n")
	for _, rule := range rules {
		nft.WriteString(fmt.Sprintf(nftFormat, "tcp", rule.LocalPort, rule.RemoteAddr, rule.RemotePort, rule.RemoteAddr, "tcp", rule.RemotePort, rule.LocalAddr))
		nft.WriteString("\n")
		nft.WriteString(fmt.Sprintf(nftFormat, "udp", rule.LocalPort, rule.RemoteAddr, rule.RemotePort, rule.RemoteAddr, "udp", rule.RemotePort, rule.LocalAddr))
		nft.WriteString("\n")
	}
	return nft.String()
}
