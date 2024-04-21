package main

import (
	"fmt"

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
	store *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer{
	storeOpts := StoreOpts{
		Root : opts.StorageRoot,
		PathTransfromfunc: opts.PathTransfromfunc,
	}
	 return &FileServer{
		FileServerOpts: opts,
		store: NewStore(storeOpts),
		quitch : make(chan struct{}),
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
	s.loop()
	return nil
}

func (s *FileServer) bootstrapNetwork() error{
	for _,addr := range s.BootstrapNodes{
		if err :=	s.Transport.Dial(addr) ; err !=nil{
			return err
		}
	}
	return nil
}

func (s * FileServer) Stop(){
	close(s.quitch)
}

func (s *FileServer) loop() {
		for{
			select{
			case msg := <- s.Transport.Consume():
				fmt.Println(msg)
			case <-s.quitch:
				return 
			}
		}
}