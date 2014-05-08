// lineserver emulates a line-based UDP server
package lineserver

import (
	"bufio"
	"launchpad.net/tomb"
	"net"
)

// LineServer is a line-based UDP server à la `nc -u -l`. Ch channel
// will receive incoming lines from all clients.
type LineServer struct {
	Ch        chan string
	conn      *net.UDPConn
	tomb.Tomb // provides: Done, Kill, Dying
}

func NewLineServer(addr string) (*LineServer, error) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}
	var ls LineServer
	ls.Ch = make(chan string)
	ls.conn = conn
	return &ls, nil
}

func (srv *LineServer) GetAddr() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", srv.conn.LocalAddr().String())
}

// Start starts the server. Call as a goroutine.
func (srv *LineServer) Start() {
	defer srv.Done()

	scanner := bufio.NewScanner(srv.conn)
	// Scanned tokens are limited in max size (64 * 1024); see
	// pkg/bufio/scan.go:MaxScanTokenSize in Go source tree.
	for scanner.Scan() {
		select {
		case srv.Ch <- scanner.Text():
		case <-srv.Dying():
			return
		}
	}
	if err := scanner.Err(); err != nil {
		srv.Kill(err)
	}
}
