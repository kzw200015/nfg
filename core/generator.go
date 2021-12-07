package core

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kzw200015/nfg/config"
)

const (
	rulePrefix = `add table ip nat
delete table ip nat
add table ip nat
add chain nat PREROUTING { type nat hook prerouting priority -100 ; }
add chain nat POSTROUTING { type nat hook postrouting priority 100 ; }

`
	ruleFormat = `add rule ip nat PREROUTING %v dport %v counter dnat to %v:%v
add rule ip nat POSTROUTING ip daddr %v %v dport %v counter snat to %v
`
)

func Generate(rules []config.Rule) string {
	var builder strings.Builder
	builder.WriteString(rulePrefix)

	for _, rule := range rules {
		switch rule.Protocol {
		case "tcp":
			builder.WriteString(genTcpRule(rule))
		case "udp":
			builder.WriteString(genUdpRule(rule))
		case "both":
			builder.WriteString(genTcpRule(rule))
			builder.WriteString(genUdpRule(rule))
		}
	}

	return builder.String()
}

func genTcpRule(rule config.Rule) string {
	return fmt.Sprintf(ruleFormat, "tcp", rule.SrcPort, rule.DstAddr, rule.DstPort, rule.DstAddr, "tcp", rule.DstPort, rule.SrcAddr)
}

func genUdpRule(rule config.Rule) string {
	return fmt.Sprintf(ruleFormat, "udp", rule.SrcPort, rule.DstAddr, rule.DstPort, rule.DstAddr, "udp", rule.DstPort, rule.SrcAddr)
}

func SaveToFile(rules []config.Rule, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	io.WriteString(f, Generate(rules))

	return nil
}
