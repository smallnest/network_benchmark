package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/smallnest/network_benchmark/stat"
	"github.com/smallnest/network_benchmark/udp"
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
	count := lportMax - lportMin

	rate := *rate / count

	var aggrStat = stat.NewAggrStat()

	var clients []*udp.Client
	for i := 0; i < count; i++ {
		lport := lportMin + i
		sport := sportMin + i

		client := udp.NewClient(*localAddr, lport, *serverAddr, sport, *pkgSize, rate, aggrStat)
		clients = append(clients, client)

		go client.Run()
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
