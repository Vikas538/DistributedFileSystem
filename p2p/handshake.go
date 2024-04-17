package p2p

import "errors"

var ErrInvalidHandShake = errors.New("invalid handshake")

type HandshakeFunc func(Peer) error

func NOPHandshakeFunc(peer Peer) error { return nil}