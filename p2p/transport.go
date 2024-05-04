package p2p

import "net"

// peer is the remote node
type Peer interface{
	net.Conn
	Send([]byte) error
	CloseStream()
}

//interface wich handles the communication it can be TCP UDP or websockets
type Transport interface{
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <- chan RPC
	Close() error
}
