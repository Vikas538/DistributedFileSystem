package p2p

import (
	
	"fmt"
	"net"

)

type TCPTransportOps struct{
	ListenAddr string
	HandshakeFunc	HandshakeFunc
	Decoder			Decoder 
	OnPeer func(Peer) error
}

type TCPTransport struct {
	listener 		net.Listener
	TCPTransportOps
	rpcch chan RPC
}

type TCPPeer struct {
	conn net.Conn
	//when dial outboud true
	outboud bool
	
}

func NewTCPPeer(conn *net.Conn,outboud bool) *TCPPeer{

	return &TCPPeer{
		conn:*conn,
		outboud:outboud,
	}
}

func NewTcpTransport(ops TCPTransportOps)*TCPTransport{

		return &TCPTransport{
			TCPTransportOps: ops,
			rpcch: make(chan RPC),
		}

}

func (p *TCPPeer) Close() error{
	return p.conn.Close()
}

// consume implement the transport interface, which will return read only channel for reading incoming msg received from peer in a a network
func (t* TCPTransport) Consume() <-chan RPC{
	return t.rpcch
}

func (t* TCPTransport) ListenAndAccept() error{
		var err error 
		t.listener,err=net.Listen("tcp",t.ListenAddr)
		if err!=nil{
				fmt.Println("Connection Failed to establish")
				return err
		}

		go t.startAcceptLoop()
		return err

}

func (t* TCPTransport) startAcceptLoop() {
	for{
			conn,err:=t.listener.Accept()
			print("TCP connection : %s\n",conn.RemoteAddr())
			if err != nil {
				println("Tcp error :%s\n",err)
				conn.Close()
			}

			go t.handleConnection(conn)
	}
}
type Temp struct{}

func (t* TCPTransport) handleConnection(conn net.Conn){
	var err error
	defer func(){
			fmt.Printf("Closing Connection: %s",err)
		}()
		peer := NewTCPPeer(&conn,true)
		if err = t.HandshakeFunc(peer);err!=nil{
				conn.Close()
		}
		if t.OnPeer != nil {
			 err = t.OnPeer(peer);if err!=nil{
				return
			 }
		}
		rpc := RPC{}
		for{
				if err:= t.Decoder.Decode(conn,&rpc);err !=nil{
					fmt.Printf("TCP error :%s\n",err)
					continue
				}
				rpc.From = conn.RemoteAddr()
				t.rpcch <- rpc
		}

}

