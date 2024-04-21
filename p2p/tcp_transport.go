package p2p

import (
	"errors"
	"fmt"
	"log"
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

func (t *TCPTransport) Dial(addr string)error{
	conn,err := net.Dial("tcp",addr)
	if err !=nil{
		return err
	}
	go t.handleConnection(conn,true)
	return nil
}

func NewTCPPeer(conn *net.Conn,outboud bool) *TCPPeer{

	return &TCPPeer{
		conn:*conn,
		outboud:outboud,
	}
}

func (t *TCPTransport) Close()error{
	fmt.Println("closing the the tcp transport connection")
	return t.listener.Close()
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
		log.Printf("tcp transport listining at port : %s\n",t.ListenAddr)
		return err

}

func (t* TCPTransport) startAcceptLoop() {
	for{
			conn,err:=t.listener.Accept()
			if errors.Is(err,net.ErrClosed){
				return 
			}
			print("TCP connection : %s\n",conn.RemoteAddr())
			if err != nil {
				println("Tcp read error :%s\n",err)
				conn.Close()
			}

			go t.handleConnection(conn,false)
	}
}
type Temp struct{}

func (t* TCPTransport) handleConnection(conn net.Conn,outbound bool){
	var err error
	defer func(){
			fmt.Printf("Closing Connection: %s",err)
		}()
		peer := NewTCPPeer(&conn,outbound)
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
				if err = t.Decoder.Decode(conn,&rpc);
				
				err !=nil {
					fmt.Printf("TCP error :%s\n",err)
					return
				}
				rpc.From = conn.RemoteAddr()
				t.rpcch <- rpc
		}

}

