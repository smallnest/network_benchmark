package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/smallnest/network_benchmark/tcp"
)

var (
	addr    = flag.String("s", "127.0.0.1", "ip address")
	portMin = flag.Int("pmin", 20000, "min port number")
	portMax = flag.Int("pmax", 20016, "max port number, exclude")
)

func main() {
	flag.Parse()

	s := tcp.NewServer(*addr, *portMin, *portMax)
	s.Run()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
