package main

import (
	"crypto/tls"
	"fmt"
	quic "github.com/Abdoueck632/mp-quic"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
)

func main() {
	var p []byte = make([]byte, config.BUFFERSIZE)

	listener, err := quic.ListenAddr("0.0.0.0:14242", utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)
	fmt.Println("Server started! Waiting for streams from client...")

	sess, err := listener.Accept()
	utils.HandleError(err)

	stream, err := sess.AcceptStream()
	utils.HandleError(err)

	fmt.Println("session created: ", sess.RemoteAddr())

	len, err := stream.Read(p)
	utils.HandleError(err)
	fmt.Println("Received %i bytes", len)

	createConnectionToRelay1("10.0.2.3:4243")
	return
}

func createConnectionToRelay1(relayaddr string) (quic.Stream, quic.Session) {
	var dataString = make([]byte, 1000)

	sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())

	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)

	streamServer.Write(dataString)
	fmt.Println("stream created...")
	fmt.Println("Client connected")

	return streamServer, sessServer
}
