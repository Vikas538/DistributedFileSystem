package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"


	"github.com/Vikas538/DistibutedFileSystem/p2p"
)

type FileServerOpts struct {
	ListenAddr string
	StorageRoot string
	PathTransfromfunc PathTransfromfunc
	Transport p2p.Transport
	BootstrapNodes []string
}

type FileServer struct{
	FileServerOpts
	peerLock sync.Mutex
	peers map[string]p2p.Peer
	store *Store
	quitch chan struct{}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key string 
}

// type DataMessage struct{
// 	Key string
// 	Data []byte
// }

func NewFileServer(opts FileServerOpts) *FileServer{
	storeOpts := StoreOpts{
		Root : opts.StorageRoot,
		PathTransfromfunc: opts.PathTransfromfunc,
	}
	 return &FileServer{
		FileServerOpts: opts,
		store: NewStore(storeOpts),
		quitch : make(chan struct{}),
		peers: make(map[string]p2p.Peer),
	 }
}

func (s * FileServer) start()error{
	defer func(){
		fmt.Println(("file server stoping due to user quit action "))
		s.Transport.Close()
	}()
	if err :=s.Transport.ListenAndAccept() ; err!=nil{
		return err
	}

		s.bootstrapNetwork()

	s.loop()
	return nil
}

func (s *FileServer)broadcast(msg *Message)error{
	peers := []io.Writer{}
	for _,peer := range s.peers{
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)		
}

func (s *FileServer) handleMessageStoreFile(from string,msg MessageStoreFile)error{
	fmt.Printf("received store file msg : %+v\n",msg)
	peer,ok := s.peers[from]
	if !ok{
		return fmt.Errorf("peer (%s) counld not be found in the peer list",peer)
	}
	peer.(*p2p.TCPPeer).Wg.Done()
	return s.store.Write(msg.Key,peer)

}

func (s *FileServer) loop() {
	for{
		select{
		case rpc := <- s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg);err != nil{
				log.Fatal(err)
			}
			if err:= s.handleMessage(rpc.From,&msg);err!=nil{
				log.Fatal(err)
			}

		
		case <-s.quitch:
			return 
		}
	}
}

func (s *FileServer)StoreData(key string, r io.Reader)error{
		// 1 store this file to disk 
		// 2 broadcast this file to all known peers in the network
		buf := new(bytes.Buffer)
		msg :=  Message{
			Payload: MessageStoreFile{
				Key: key,
			},
		}
		if err := gob.NewEncoder(buf).Encode(msg);err!=nil{
			return err
		}

		for _,peer :=range s.peers{
			if err := peer.Send(buf.Bytes());err!=nil{
				return err
			}
		}
		// time.Sleep(time.Second * 3)

		// payload := []byte("THIS IS A LARGE  FILE")
		// for _, peer := range s.peers{
		// 	if err := peer.Send(payload);err !=nil{
		// 		return err
		// 	}
		// }

		return nil
		
		// buf := new(bytes.Buffer)
		// tee := io.TeeReader(r,buf)
		// if err := s.store.Write(key,tee);err!=nil{
		// 	return err
		// }buf.Bytes()

		// _,err :=io.Copy(buf,r)
		// if err!=nil{
		// 	return err
		// }
		// p:=&DataMessage{
		// 	Key: key,
		// 	Data: buf.Bytes(),
		// }
		// fmt.Println(buf.Bytes())
		// return s.broadcast(&Message{
		// 	From: "todo",
		// 	Payload: p,

		// })
}

func (s *FileServer) bootstrapNetwork() error{
	go func(){
	for _,addr := range s.BootstrapNodes{
		if len(addr)==0 {
			continue
		}
			if err :=	s.Transport.Dial(addr) ; err !=nil{
				log.Println("dial err ",err)
				continue
			}
		}
		}()
	return nil
}

func (s *FileServer) handleMessage(from string,msg *Message) error{
	switch v:= msg.Payload.(type){
		case MessageStoreFile:
			return s.handleMessageStoreFile(from,v)	
	}
	return nil 
}

func (s * FileServer) Stop(){
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer)error{
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] =p
	log.Printf("connected with peer remote %s",p.RemoteAddr())
	return nil 
}

func init(){
	gob.Register(MessageStoreFile{})
}