package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type PathTransfromfunc func(string) PathKey

const defaultRootFolderName = "ggnetwork"

type PathKey struct{
	PathName string
	FileName string
}


func NewStore(opts StoreOpts) * Store{
	if opts.PathTransfromfunc == nil{
		opts.PathTransfromfunc=DefaultPathTrasformFunc
	}
	if len(opts.Root)==0{
		opts.Root = defaultRootFolderName
	}
	return &Store{
			StoreOpts: opts,
	}
}

func (s *Store) Read(key string) (io.Reader,error){
	f,err := s.readStream(key)
	if err != nil{
		return nil,err
	}
	buf := new(bytes.Buffer)
	_,err = io.Copy(buf,f)
	return buf,nil
}

func (s *Store) Has(key string) bool{
	PathKey:= s.PathTransfromfunc(key)
	filePathWithRoot := fmt.Sprintf("%s/%s",s.Root,PathKey.FullPath())
	_,err := os.Stat(filePathWithRoot) ; 
	if errors.Is(err,os.ErrNotExist){
		return false
	}
	return true 
}

func (p *PathKey) FirstPathName() string{
	return strings.Split(p.FullPath(), "/")[0]
}

func (s *Store)Delete(key string)error{
	pathKey := s.PathTransfromfunc(key)
	defer func(){
		log.Printf("deleted [%s] from disk",pathKey.FileName)
	}()
	firstPathWithRoot :=  s.Root+"/"+pathKey.FirstPathName()
	return os.RemoveAll(firstPathWithRoot)
}


func (s *Store) readStream(key string) ( io.ReadCloser,error){
		pathKey := s.PathTransfromfunc(key)
		pathKeyWithRoot :=fmt.Sprintf("%s/%s",s.Root,pathKey.FullPath())
		return os.Open(pathKeyWithRoot) 
}

func CASPathTransformFunc(key string) PathKey{
	hash := sha1.Sum([]byte(key))
	hashstr := hex.EncodeToString(hash[:])

	blocksize := 5
	sliceLen := len(hashstr)/blocksize
	paths := make([]string,sliceLen)

	for i:=0 ;i<sliceLen;i++{
		from,to := i * blocksize,(i*blocksize)+blocksize
		paths[i]=hashstr[from:to]
	}
	return PathKey{
		PathName: strings.Join(paths,"/"),
		FileName: hashstr,
	}
}

var DefaultPathTrasformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

type StoreOpts struct{
	Root string
	PathTransfromfunc PathTransfromfunc
}

type Store struct{
	StoreOpts
}


func (s *Store) writestream(key string,r io.Reader)error{
	pathKey := s.PathTransfromfunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s",s.Root,pathKey.PathName)
	if err:= os.MkdirAll(pathNameWithRoot,os.ModePerm);err != nil{
		return err
	}

	
	fullPathWithRoot := fmt.Sprintf("%s/%s",s.Root,pathKey.FullPath())
	f,err:=os.Create(fullPathWithRoot);if err !=nil{
			return err
	}
	n,err := io.Copy(f,r)
	log.Printf("writter %d bytes to disk %s",n,fullPathWithRoot)
	return nil
} 


func (p *PathKey) FullPath()string{
		return fmt.Sprintf("%s/%s",p.PathName,p.FileName)
}

