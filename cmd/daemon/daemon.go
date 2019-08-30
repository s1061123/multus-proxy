//
package main

import (
	"math/rand"
	"os"
	"time"
	"github.com/s1061123/multus-proxy/pkg/proxy"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	command := proxy.NewDaemonCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
