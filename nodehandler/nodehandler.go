package nodehandler

import (
	"encoding/json"
	"fmt"
	"net"
)

/* Informations about node */
type Node struct {
	NodeID   	int    	`json:"nodeId"`
	Port     	string 	`json:"port"`
	IsMaster 	bool   	`json:"isMaster"`
	Item			Item		`json:"item"`
}

type Request struct {
	Item	Item 		`json:"item"`
	Type	string	`json:"type"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Item struct {
	ID     int    `json:"id"`
	Tenant string `json:"tenant"`
}

func (node *Node) Run() error {

	fmt.Println(node)

	err := node.ListenOnPort()

	return err
}

func (node *Node) ListenOnPort() error {
	/* Listen for incoming messages */
	ln, _ := net.Listen("tcp", ":" + node.Port)
	/* accept connection on port */
	connIn, err := ln.Accept()
	if err == nil {
		var req Request
		json.NewDecoder(connIn).Decode(&req)
		fmt.Printf("Request: %v", req)
		resp := Response{"OK", "Ciao"}
		json.NewEncoder(connIn).Encode(&resp)
		connIn.Close()
	}
	return err
}