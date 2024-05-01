package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
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
	net.Conn
	//when dial outboud true
	outboud bool
	Wg *sync.WaitGroup
}

// func (p * TCPPeer) RemoteAddr() net.Addr{
// 	return p.conn.RemoteAddr()
// }


func (p *TCPPeer)Send(b []byte)error{
	_,err := p.Conn.Write(b)
	return err
}

func (t *TCPTransport) Dial(addr string)error{
	conn,err := net.Dial("tcp",addr)
	if err !=nil{
		return err
	}
	go t.handleConnection(conn,true)
	return nil
}

func NewTCPPeer(conn net.Conn,outboud bool) *TCPPeer{

	return &TCPPeer{
		Conn: conn,
		outboud:outboud,
		Wg: &sync.WaitGroup{},
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

// func (p *TCPPeer) Close() error{
// 	return p.conn.Close()
// }

// consume implement the transport interface, which will return read only channel for reading incoming msg received from peer in a a network
func (t* TCPTransport) Consume() <-chan RPC{
	return t.rpcch
}

func (t* TCPTransport) ListenAndAccept() error{
		var err error 
		fmt.Printf("listen addres len = %dand add %s ",len(t.ListenAddr),t.ListenAddr)
		if len(t.ListenAddr) ==0{ return nil}
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
			}
			fmt.Printf("new incoiming connection %v\n",conn)
			go t.handleConnection(conn,false)
	}
}
type Temp struct{}

func (t* TCPTransport) handleConnection(conn net.Conn,outbound bool){
	var err error
	defer func(){
			fmt.Printf("Closing Connection: %s",err)
		}()
		peer := NewTCPPeer(conn,outbound)
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
				rpc.From = conn.RemoteAddr().String()
				peer.Wg.Add(1)
				fmt.Println("Waiting till stream is done")
				t.rpcch <- rpc
				peer.Wg.Wait()
				fmt.Println("stream done continuing normal loop")
		}

}



