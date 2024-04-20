package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestPathTransformFunc(t *testing.T){
	key := "momsbestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedOriginal := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if(pathKey.PathName !=  expectedPathName){
		t.Errorf("have %s want %s",pathKey.PathName,expectedPathName)
	}
	if(pathKey.FileName !=  expectedOriginal){
		t.Errorf("have %s want %s",pathKey.FileName,expectedOriginal)
	}
}

func TestStore(t *testing.T){
	opts := StoreOpts {
		PathTransfromfunc : CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "momspecials"
	data := []byte("some jpg bytes")
	if err := s.writestream(key,bytes.NewReader(data));err!=nil{
		t.Error(err)
	}
	r,err :=s.Read(key);if(err!=nil){
		t.Error(err)
	}
	b,_ := ioutil.ReadAll(r)
	if string(b) != string(data){
		t.Errorf("wants %s got &%s",data,b)
	}

	s.Delete(key)
}

func TestStoreDelete(t* testing.T){
	opts := StoreOpts {
		PathTransfromfunc : CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "momspecials"
	data := []byte("some jpg bytes")
	if err := s.writestream(key,bytes.NewReader(data));err!=nil{
		t.Error(err)
	}
	if err := s.Delete(key);err!=nil{
		t.Error(err)
	}
}