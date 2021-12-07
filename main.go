package main

import (
	"flag"
	"fmt"

	"github.com/kzw200015/nfg/config"
	"github.com/kzw200015/nfg/core"
	"github.com/kzw200015/nfg/log"
)

func main() {
	isWatch := flag.Bool("w", false, "Run as a watch mode")
	isGenerate := flag.Bool("g", false, "Generate to stdout")
	src := flag.String("s", "", "The source of config")
	temp := flag.String("t", "/tmp/nat.nft", "The rule file path")
	flag.Parse()

	if *src == "" {
		log.Logger.Errorln("Please input the config file source")
		flag.Usage()
		return
	}

	if *isWatch {
		w := core.NewWatcher(*src, *temp)
		w.Watch()
	} else if *isGenerate {
		c, err := config.NewConfig(*src)
		if err != nil {
			log.Logger.Panicln(err)
		}
		fmt.Print(core.Generate(c.Rules))
	} else {
		flag.Usage()
	}
}
