package zeroutine

import (
	zmq "github.com/alecthomas/gozmq"
)

type Zeroutine struct {
	PubAddr         string // Publisher Endpoint Address 
	SubAddr         string // Subscriber Endpoint Address
	BufferSize      int    // Memory buffer size
	SubscribeFilter string
}

func (z Zeroutine) RunBroker() error {
	broker, err := NewBroker(z)
	if err == nil {
		err = broker.Run()
	}
	return err
}

func (z Zeroutine) NewPubSocket() (zmq.Socket, error) {
	ctx, err := GetGlobalContext()
	if err != nil {
		return nil, err
	}

	sock, err := ctx.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}

	// prevent 0mq from infinitely buffering messages
	for _, hwm := range []zmq.IntSocketOption{zmq.SNDHWM, zmq.RCVHWM} {
		err = sock.SetSockOptInt(hwm, z.BufferSize)
		if err != nil {
			sock.Close()
			return nil, err
		}
	}

	return sock, nil
}
