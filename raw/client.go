package raw

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/kataras/golog"
	"github.com/smallnest/network_benchmark/stat"
	"go.uber.org/ratelimit"
	"golang.org/x/net/bpf"
	"golang.org/x/net/ipv4"
)

type Client struct {
	localAddr   string
	localPorts  []int
	serverAddr  string
	serverPorts []int
	pktSize     int
	rate        int

	conn  *net.TCPConn
	stats *stat.Stats
}

func NewClient(localAddr string, localPorts []int, serverAddr string, serverPorts []int, pktSize int, rate int, aggrStat *stat.AggrStat) *Client {
	return &Client{
		localAddr:   localAddr,
		localPorts:  localPorts,
		serverAddr:  serverAddr,
		serverPorts: serverPorts,
		pktSize:     pktSize,
		rate:        rate,
		stats:       stat.NewStats(fmt.Sprintf("%s:%v->%s:%v", localAddr, localPorts, serverAddr, serverPorts), aggrStat),
	}
}

func (c *Client) Run() {
	// srcAddr := &net.TCPAddr{IP: net.ParseIP(c.localAddr), Port: c.localPort}
	// dstAddr := &net.TCPAddr{IP: net.ParseIP(c.serverAddr), Port: c.serverPort}

	// conn, err := net.DialTCP("tcp", srcAddr, dstAddr)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	go c.read()

	localIP := net.ParseIP(c.localAddr).To4()
	serverIP := net.ParseIP(c.serverAddr).To4()

	var seq uint64

	for i := 0; i < len(c.localPorts); i++ {
		localPort := c.localPorts[i]
		serverPort := c.serverPorts[i]

		// fmt.Println("localPort:", localPort, "serverPort:", serverPort)

		// 每个端口单独一个goroutine发送数据
		go func() {
			conn, err := net.ListenPacket("ip4:udp", c.localAddr)
			if err != nil {
				golog.Fatalf("failed to ListenPacket: %v", err)
			}
			defer conn.Close()

			// 只负责发送，不处理接受
			ipv4Conn := ipv4.NewPacketConn(conn)
			filter := createEmptyFilter()
			if assembled, err := bpf.Assemble(filter); err == nil {
				ipv4Conn.SetBPF(assembled)
			}

			payload := bytes.Repeat([]byte("a"), c.pktSize)

			rateLimter := ratelimit.New(c.rate, ratelimit.Per(time.Second))

			for {
				rateLimter.Take()
				bizSeq := atomic.AddUint64(&seq, 1)

				ts := time.Now().UnixNano()
				binary.LittleEndian.PutUint64(payload[:8], bizSeq)
				binary.LittleEndian.PutUint64(payload[8:16], uint64(ts))
				data, err := encodeUDPPacket(localIP, serverIP, uint16(localPort), uint16(serverPort), 64, payload)

				if err != nil {
					golog.Fatalf("failed to encodeUDPPacket: %v", err)
					continue
				}

				c.stats.AddSent(bizSeq, ts)
				if _, err := conn.WriteTo(data, &net.IPAddr{IP: serverIP}); err != nil {
					golog.Errorf("failed to write packet: %v", err)
					continue
				}
			}
		}()
	}

}

func encodeUDPPacket(localIP, remoteIP net.IP, localPort, remotePort uint16, tos uint8, payload []byte) ([]byte, error) {
	ip := &layers.IPv4{
		Version:  4,
		TTL:      128,
		SrcIP:    localIP,
		DstIP:    remoteIP,
		TOS:      tos,
		Protocol: layers.IPProtocolUDP,
	}

	udp := &layers.UDP{
		SrcPort: layers.UDPPort(localPort),
		DstPort: layers.UDPPort(remotePort),
	}
	udp.SetNetworkLayerForChecksum(ip)

	// Serialize.  Note:  we only serialize the TCP layer, because the
	// socket we get with net.ListenPacket wraps our data in IPv4 packets
	// already.  We do still need the IP layer to compute checksums
	// correctly, though.
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	err := gopacket.SerializeLayers(buf, opts, udp, gopacket.Payload(payload))

	return buf.Bytes(), err
}

func (c *Client) read() {

	conn, err := net.ListenPacket("ip4:udp", c.localAddr)
	if err != nil {
		golog.Fatalf("failed to ListenPacket: %v", err)
	}

	// // 设置buffer
	cc := conn.(*net.IPConn)
	cc.SetReadBuffer(20 * 1024 * 1024)
	// cc.SetWriteBuffer(20 * 1024 * 1024)

	ipv4Conn := ipv4.NewPacketConn(conn)
	filter := createFilter(c.localPorts[0], c.localPorts[len(c.localPorts)-1], 64)
	if assembled, err := bpf.Assemble(filter); err == nil {
		ipv4Conn.SetBPF(assembled)
	}

	go c.readResponse(conn)

}

func (c *Client) readResponse(conn net.PacketConn) {
	defer conn.Close()

	b := make([]byte, 2048)
	errCount := 10
	for {
		n, remoteAddr, err := conn.ReadFrom(b)
		if err != nil {
			golog.Errorf("failed to read : %v", err)
			errCount++
			if errCount == 20 {
				os.Exit(-1)
			}
			continue
		}

		errCount = 0

		c.handlePacket(remoteAddr, b[:n])
	}
}

func (c *Client) handlePacket(remoteAddr net.Addr, pkt []byte) {
	// Decode a packet
	packet := gopacket.NewPacket(pkt, layers.LayerTypeUDP, gopacket.NoCopy)
	// Get the TCP layer from this packet
	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		app := packet.ApplicationLayer()
		if app == nil {
			return
		}
		payload := app.Payload()
		if len(payload) != c.pktSize {
			return
		}

		seq := binary.LittleEndian.Uint64(payload[:8])
		ts := binary.LittleEndian.Uint64(payload[8:16])

		// fmt.Println("seq:", seq, "ts:", ts)
		c.stats.AddRecv(seq, int64(ts))
	}
}

func (c *Client) Close() {

}
