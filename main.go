package main

import (
	"fmt"
	"log"
	"github.com/Vikas538/DistibutedFileSystem/p2p"
)

func OnPeer(peer p2p.Peer)error{
	peer.Close()
	fmt.Printf("doing some logic with the peer outside the ")
	return nil
}

func makeServer(listenAddr string,nodes ...string) *FileServer{
	tcpTransaportOps := p2p.TCPTransportOps{
		ListenAddr: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTcpTransport(tcpTransaportOps)
	FileServerOpts := FileServerOpts{
		StorageRoot: listenAddr + "_network",
		PathTransfromfunc: CASPathTransformFunc,
		Transport: tcpTransport,
		BootstrapNodes: nodes,
	} 
	return NewFileServer(FileServerOpts)
}

func main(){
	s1 := makeServer(":3000","")
	s2 := makeServer(":4000",":3000")
	go func() {
		log.Fatal(s1.start())
	}()
	s2.start()
}



