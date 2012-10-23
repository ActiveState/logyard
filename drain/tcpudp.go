package drain

import (
	"launchpad.net/tomb"
	"log"
	"logyard"
	"net"
	"time"
)

// IPConnDrain is a drain based on net.IPConn
type IPConnDrain struct {
	log *log.Logger
	tomb.Tomb
}

func NewIPConnDrain(log *log.Logger) Drain {
	rd := &IPConnDrain{}
	rd.log = log
	return rd
}

func (d *IPConnDrain) Start(config *DrainConfig) {
	defer d.Done()

	if !(config.Scheme == "udp" || config.Scheme == "tcp") {
		d.Killf("Invalid scheme: %s", config.Scheme)
		return
	}

	d.log.Printf("Connecting to %s addr %s ...", config.Scheme, config.Host)
	conn, err := net.DialTimeout(config.Scheme, config.Host, 10*time.Second)
	if err != nil {
		d.Kill(err)
		return
	}
	defer conn.Close()
	d.log.Printf("Connected to %s addr %s\n", config.Scheme, config.Host)

	c, err := logyard.NewClientGlobal()
	if err != nil {
		d.Kill(err)
		return
	}
	defer c.Close()

	ss, err := c.Recv(config.Filters)
	if err != nil {
		d.Kill(err)
		return
	}
	defer ss.Stop()

	for {
		select {
		case msg := <-ss.Ch:
			data, err := config.FormatJSON(msg.Value)
			if err != nil {
				d.Kill(err)
				return
			}
			_, err = conn.Write(data)
			if err != nil {
				d.Kill(err)
				return
			}
		case <-d.Dying():
			return
		}
	}
	d.log.Println("Exiting")
}

func (d *IPConnDrain) Stop() error {
	d.log.Println("Stopping...")
	d.Kill(nil)
	return d.Wait()
}
