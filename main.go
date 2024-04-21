package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Vikas538/DistibutedFileSystem/p2p"
)

func OnPeer(peer p2p.Peer)error{
	peer.Close()
	fmt.Printf("doing some logic with the peer outside the ")
	return nil
}

func main(){
	tcptransportOPts := p2p.TCPTransportOps{
		ListenAddr: ":3000",
		Decoder: p2p.DefaultDecoder{},
		HandshakeFunc: p2p.NOPHandshakeFunc,
	}
	tcpTransport := p2p.NewTcpTransport(tcptransportOPts)
	FileServerOpts:=FileServerOpts{
		StorageRoot: "3000_network",
		PathTransfromfunc: CASPathTransformFunc,
		Transport: tcpTransport,
	}
	s := NewFileServer(FileServerOpts)
	go func (){
			time.Sleep(time.Second * 3)
			s.Stop()
	}()
	if err := s.start();err!=nil{
		log.Fatal(err)
	}


	select{}
}



