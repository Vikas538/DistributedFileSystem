package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

type PathTransfromfunc func(string) PathKey

type PathKey struct{
	PathName string
	FileName string
}


func NewStore(opts StoreOpts) * Store{
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
	_,err := os.Stat(PathKey.FullPath()) ; if err != fs.ErrNotExist{
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
	return os.RemoveAll(pathKey.FirstPathName())
}


func (s *Store) readStream(key string) ( io.ReadCloser,error){
		pathKey := s.PathTransfromfunc(key)
		return os.Open(pathKey.FullPath()) 
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

var DefaultPathTrasformFunc = func(key string) string {
	return key
}

type StoreOpts struct{
	PathTransfromfunc PathTransfromfunc
}

type Store struct{
	StoreOpts
}


func (s *Store) writestream(key string,r io.Reader)error{
	pathKey := s.PathTransfromfunc(key)
	if err:= os.MkdirAll(pathKey.PathName,os.ModePerm);err != nil{
		return err
	}

	fullPath := pathKey.FullPath()

	f,err:=os.Create(fullPath);if err !=nil{
			return err
	}
	n,err := io.Copy(f,r)
	log.Printf("writter %d bytes to disk %s",n,fullPath)
	return nil
} 


func (p *PathKey) FullPath()string{
		return fmt.Sprintf("%s/%s",p.PathName,p.FileName)
}

