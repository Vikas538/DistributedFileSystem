package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface{
    Decode(io.Reader,*RPC) error
}

type GOBDecoder struct{}

func (g *GOBDecoder) Decode(r io.Reader, msg *RPC) error {
    return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader,msg *RPC) error{
    peekBuf := make([]byte,1)
    if _,err :=r.Read(peekBuf);err !=nil{
        return nil
    }
    // In case f stea, we are not decoing what is being sent over the network and we are jst setting stream true so we can handle in the or logic 
    stream := peekBuf[0] == IncomingStream
    if stream {
        msg.Stream = true
        return nil
    }
    buf := make([]byte,1028)
    n,err := r.Read(buf)
    if err != nil{
        return err
    }
    msg.Payload = buf[:n]
    return nil
}