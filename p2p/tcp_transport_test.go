package p2p

import (
	"testing"
	"github.com/stretchr/testify/assert"
)



func TestTcpTransport(t * testing.T){
	listenAddr:=":8080"
	tcpOpts := TCPTransportOps{
		ListenAddr: listenAddr,
		HandshakeFunc: NOPHandshakeFunc,
		Decoder: DefaultDecoder{},
	}
	tr:= NewTcpTransport(tcpOpts)
	assert.Equal(t,tr.ListenAddr,listenAddr)
	assert.Nil(t,tr.ListenAndAccept())
}