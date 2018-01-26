package main

import (
	"github.com/tigerlaibao/chatroom/netx"
	"flag"
)

var (
	port int
)

func init(){
	flag.IntVar(&port , "port" , 8080 , "the port")
	flag.Parse()
}

func main() {
	netx.StartServer(port)
}
