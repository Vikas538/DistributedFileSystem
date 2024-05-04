package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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
	Size int64
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

func (s *FileServer)stream(msg *Message)error{
	peers := []io.Writer{}
	for _,peer := range s.peers{
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)		
}

func (s *FileServer)broadcast(msg *Message)error{
	msgbuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgbuf).Encode(msg);err!=nil{
		return err
	}

	for _,peer :=range s.peers{
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(msgbuf.Bytes());err!=nil{
			return err
		}
	}
	return nil
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

type MessageGetFile struct{
	Key string
}

func (s *FileServer) Get(key string) (io.Reader,error){
	if s.store.Has(key){
	fmt.Printf("[%s] service file (%s) from local disk\n",s.Transport.Addr(),key)
		return s.store.Read(key)
	}
	fmt.Printf("[%s] dont have file (%s)locally fetching from network\n",s.Transport.Addr(),key)
	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err :=s.broadcast(&msg);err!=nil{
		return nil,err
	}
	time.Sleep(time.Millisecond * 3)
	for _,peer := range s.peers{
		 n,err :=s.store.Write(key,io.LimitReader(peer,22));if err !=nil{
			return nil,err
		}
		// fileBuffer := new(bytes.Buffer)
		// n,err := io.Copy(fileBuffer,peer)
		// if err!=nil{
		// 	return nil,err
		// }
		fmt.Printf("[%s] received (%d )bytes over the network from : (%s) \n",s.Transport.Addr(),n,peer.RemoteAddr())
		peer.CloseStream()
		
	}
	return s.store.Read(key)
}

func (s *FileServer)Store(key string, r io.Reader)error{
		var (		
			fileBuffer = new(bytes.Buffer)
			tee = io.TeeReader(r,fileBuffer)
	)


		size,err := s.store.Write(key,tee)
		if err!=nil{
			return err
		}
		msg :=  Message{
			Payload: MessageStoreFile{
				Key: key,
				Size: size,
			},
		}
		if err:=s.broadcast(&msg);err!=nil{
			return nil
		}
		time.Sleep(time.Millisecond * 500)


		for _, peer := range s.peers{
			peer.Send([]byte{p2p.IncomingStream})
			n,err := io.Copy(peer,fileBuffer)
			if err !=nil{
				return err
			}
			fmt.Println("received and written bytes to disk :: ",n )
		}

		return nil
	
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
		case MessageGetFile:
			return s.handleMessageGetFile(from,v)
	}
	return nil 
}

func (s *FileServer)handleMessageGetFile(from string,msg MessageGetFile)error{
	fmt.Printf("[%s] need to get a file (%s) from the disk and set it over the wire\n",s.Transport.Addr(),msg.Key)
	if !s.store.Has(msg.Key){
		return fmt.Errorf("file with key (%s) not fousnd in the remote peer\n",msg.Key)
	}
	r,err :=s.store.Read(msg.Key);if err!=nil{
		return err
	}
	peer,ok :=s.peers[from]
	if!ok{
		return fmt.Errorf("peer not in peers list %s\n",from)
	}
	peer.Send([]byte{p2p.IncomingStream})
	n,err :=io.Copy(peer,r)
	if err !=nil{
		return err
	}
	fmt.Printf("[%s] written %d bytes over the network to %s\n",s.Transport.Addr(),n,from)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string,msg MessageStoreFile)error{
	fmt.Printf("received store file msg : %+v\n",msg)
	peer,ok := s.peers[from]
	if !ok{
		return fmt.Errorf("peer (%s) could not be found in the peer list",peer)
	}
	 n,err := s.store.Write(msg.Key,io.LimitReader(peer,msg.Size)) ;
	if err !=nil {
		return err
	}
	fmt.Printf("[%s] Written %d bytes to disk \n",s.Transport.Addr(),n)
	// peer.(*p2p.TCPPeer).Wg.Done()
	peer.CloseStream()
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
	gob.Register(MessageGetFile{})
}