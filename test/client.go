package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	// "time"
)

type Message struct {
	ID   string
	Data string
}

func send(conn net.Conn) {
	// lets create the message we want to send accross
	msg := Message{ID: "Yo", Data: "Hello"}
	bin_buf := new(bytes.Buffer)

	// create a encoder object
	gobobj := gob.NewEncoder(bin_buf)
	// encode buffer and marshal it into a gob object
	gobobj.Encode(msg)

	conn.Write(bin_buf.Bytes())
}

func recv(conn net.Conn) {
	// create a temp buffer
	tmp := make([]byte, 500)
	conn.Read(tmp)

	// convert bytes into Buffer (which implements io.Reader/io.Writer)
	tmpbuff := bytes.NewBuffer(tmp)
	tmpstruct := new(Message)

	// creates a decoder object
	gobobjdec := gob.NewDecoder(tmpbuff)
	// decodes buffer and unmarshals it into a Message struct
	gobobjdec.Decode(tmpstruct)

	fmt.Println(tmpstruct)
}

func main() {
	conn, _ := net.Dial("tcp", ":8081")

	// Uncomment to test timeout
	// time.Sleep(5 * time.Second)
	// return

	send(conn)
	recv(conn)
}
