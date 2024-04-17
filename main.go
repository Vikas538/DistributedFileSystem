package main

import (
	"fmt"

	"github.com/Vikas538/DistibutedFileSystem/p2p"
)

func OnPeer(peer p2p.Peer)error{
	peer.Close()
	fmt.Printf("doing some logic with the peer outside the ")
	return nil
}

func main(){
	tcpOpts := p2p.TCPTransportOps{
		ListenAddr: ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
		OnPeer: OnPeer,
	}
	tr:= p2p.NewTcpTransport(tcpOpts)
	go func ()  {
		for {
			rpc := <-tr.Consume()
			fmt.Printf("%+v\n",rpc)
		}
	}()
	if err:= tr.ListenAndAccept();err!=nil{
		fmt.Println(err)
	}
	select{}
	// fmt.Println("Read Gucci")
}



