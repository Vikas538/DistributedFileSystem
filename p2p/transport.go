package p2p

// peer is the remote node 
type Peer interface{
	Close() error
}

//interface wich handles the communication it can be TCP UDP or websockets
type Transport interface{
	listenAndAccept() error
	Consume() <- chan RPC
}