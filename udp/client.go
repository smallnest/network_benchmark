package udp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/smallnest/network_benchmark/stat"
	"go.uber.org/ratelimit"
)

type Client struct {
	localAddr  string
	localPort  int
	serverAddr string
	serverPort int
	pktSize    int
	rate       int

	conn  *net.UDPConn
	stats *stat.Stats
}

func NewClient(localAddr string, localPort int, serverAddr string, serverPort int, pktSize int, rate int, aggrStat *stat.AggrStat) *Client {
	return &Client{
		localAddr:  localAddr,
		localPort:  localPort,
		serverAddr: serverAddr,
		serverPort: serverPort,
		pktSize:    pktSize,
		rate:       rate,
		stats:      stat.NewStats(fmt.Sprintf("%s:%d->%s:%d", localAddr, localPort, serverAddr, serverPort), aggrStat),
	}
}

func (c *Client) Run() {
	defer func() {
		time.Sleep(5 * time.Second)
		c.stats.Close()
	}()

	srcAddr := &net.UDPAddr{IP: net.ParseIP(c.localAddr), Port: c.localPort}
	dstAddr := &net.UDPAddr{IP: net.ParseIP(c.serverAddr), Port: c.serverPort}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		log.Fatal(err)
	}
	c.conn = conn
	go c.read()

	data := bytes.Repeat([]byte("a"), c.pktSize)

	rateLimter := ratelimit.New(c.rate, ratelimit.Per(time.Second))

	for seq := uint64(0); ; seq++ {
		rateLimter.Take()
		ts := time.Now().UnixNano()
		binary.LittleEndian.PutUint64(data[:8], seq)
		binary.LittleEndian.PutUint64(data[8:16], uint64(ts))
		c.stats.AddSent(seq, ts)
		_, err = conn.Write(data)
		if err != nil {
			return
		}
	}
}

func (c *Client) read() {
	data := make([]byte, c.pktSize)

	for {
		n, err := c.conn.Read(data)
		if err != nil {
			return
		}
		if n < c.pktSize {
			continue
		}

		seq := binary.LittleEndian.Uint64(data[:8])
		ts := binary.LittleEndian.Uint64(data[8:16])

		c.stats.AddRecv(seq, int64(ts))
	}
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
