package core

import (
	"os/exec"
)

func Apply(path string) error {
	return exec.Command("nft", "-f", path).Run()
}
