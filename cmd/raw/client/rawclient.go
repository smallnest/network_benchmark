package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/smallnest/network_benchmark/raw"
)

var (
	localAddr   = flag.String("laddr", "", "local address")
	localPorts  = flag.String("lports", "20000,20016", "local ports")
	serverPorts = flag.String("sports", "20000,20016", "server ports")
	serverAddr  = flag.String("saddr", "", "server address")
	pkgSize     = flag.Int("pkgsize", 64, "package size")
	rate        = flag.Int("rate", 10000, "rate")
)

func main() {
	flag.Parse()

	localPorts := strings.Split(*localPorts, ",")
	serverPorts := strings.Split(*serverPorts, ",")

	lportMin, _ := strconv.Atoi(localPorts[0])
	sportMin, _ := strconv.Atoi(serverPorts[0])
	lportMax, _ := strconv.Atoi(localPorts[1])
	// sportMax, _ := strconv.Atoi(serverPorts[1])
	count := lportMax - lportMin + 1

	rate := *rate / count

	var lPorts []int
	var sPorts []int

	for i := 0; i < count; i++ {
		lPorts = append(lPorts, lportMin+i)
		sPorts = append(sPorts, sportMin+i)
	}

	client := raw.NewClient(*localAddr, lPorts, *serverAddr, sPorts, *pkgSize, rate)
	go client.Run()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
