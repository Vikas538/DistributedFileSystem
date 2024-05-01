package main

import (
	"bytes"
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
	s :=  NewFileServer(FileServerOpts)

	tcpTransport.OnPeer =s.OnPeer

	return s
}

func main(){
	s1 := makeServer(":3000","")
	s2 := makeServer(":4000",":3000")
	go func() {
		log.Fatal(s1.start())
	}()
	time.Sleep(1*time.Second)
	go s2.start()
	time.Sleep(1*time.Second)
	data := bytes.NewReader([]byte("my big data file here!"))
	s2.StoreData("myprivatekey",data)
	select {}
}



