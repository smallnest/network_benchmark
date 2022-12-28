package tcp

import (
	"io"
	"log"
	"net"
)

type Server struct {
	addr             string
	portMin, portMax int
}

func NewServer(addr string, portMin, portMax int) *Server {
	return &Server{
		addr:    addr,
		portMin: portMin,
		portMax: portMax,
	}
}

func (s *Server) Run() {
	// 每个端口一个goroutine
	for i := s.portMin; i < s.portMax; i++ {
		go s.runByPort(i)
	}
}

func (s *Server) runByPort(port int) {
	conn, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(s.addr),
		Port: port,
	})

	if err != nil {
		panic(err)
	}

	for {
		conn, err := conn.Accept()
		if err != nil {
			log.Fatalf("failed to accept: %v", err)
		}
		go s.handleConn(conn.(*net.TCPConn))
	}

}

func (s *Server) handleConn(conn *net.TCPConn) {
	conn.SetReadBuffer(20 * 1024 * 1024)
	conn.SetWriteBuffer(20 * 1024 * 1024)

	// data := make([]byte, 2048)
	for {
		// n, err := conn.Read(data)
		// if err != nil {
		// 	log.Fatalf("failed to read: %v", err)
		// }
		// if n <= 0 {
		// 	continue
		// }

		// conn.Write(data[:n])

		_, err := io.Copy(conn, conn)
		if err != nil {
			log.Fatalf("failed to read: %v", err)
		}
	}
}
