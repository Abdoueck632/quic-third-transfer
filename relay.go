package main

import (
	"abdou.seck/quic-third-transfer/config"
	"abdou.seck/quic-third-transfer/utils"
	"crypto/tls"
	"fmt"
	"os"

	quic "github.com/Abdoueck632/mp-quic"
	"strconv"
	"strings"
	"time"
)

func main() {

	savePath := os.Args[1]
	fmt.Println("Saving file to: ", savePath)

	fmt.Println("Attaching to: ", config.Addr)
	listener, err := quic.ListenAddr(config.Addr, utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("Server started! Waiting for streams from client...")

	sess, err := listener.Accept()
	utils.HandleError(err)

	fmt.Println("session created: ", sess.RemoteAddr())

	stream, err := sess.AcceptStream()
	utils.HandleError(err)

	fmt.Println("stream created: ", stream.StreamID())

	defer stream.Close()
	fmt.Println("Connected to server, start receiving the file name and file size")
	filename := make([]byte, 64)
	addrClient := make([]byte, 20)
	//stream2 :=make([]byte,64)

	// var quic1 quic.Stream
	stream.Read(filename)
	stream.Read(addrClient)
	//stream.Read(stream2)
	// quic1:= quic.Stream(stream2)

	//fmt.Println("_______________hello ", quic1)

	filename1 := strings.Trim(string(filename), ":")
	addrclient1 := strings.Trim(string(addrClient), ":")

	name := savePath + filename1
	file, err := os.Open(name)
	utils.HandleError(err)

	fileInfo, err := file.Stat()

	utils.HandleError(err)
	if fileInfo.Size() <= config.Threshold {
		config.QuicConfig.CreatePaths = false
		fmt.Println("File is small, using single path only.")
	} else {
		fmt.Println("file is large, using multipath now.")
	}
	file.Close()

	fmt.Println("Trying to connect to: ", addrclient1, "Filename ", filename1)
	sess1, err := quic.DialAddr(addrclient1, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created: ", sess.RemoteAddr())

	stream1, err := sess1.OpenStream()
	utils.HandleError(err)

	fmt.Println("stream created...")
	fmt.Println("Client connected")
	sendFile(stream1, name)
	time.Sleep(2 * time.Second)
	fmt.Println(sess.GetConnectionID())
	fmt.Println(sess1.GetConnectionID())

}
func sendFile(stream quic.Stream, fileToSend string) {
	fmt.Println("A client has connected!")
	defer stream.Close()

	file, err := os.Open(fileToSend)
	utils.HandleError(err)

	fileInfo, err := file.Stat()
	utils.HandleError(err)

	fileSize := utils.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := utils.FillString(fileInfo.Name(), 64)

	fmt.Println("Sending filename and filesize!")
	stream.Write([]byte(fileSize))
	stream.Write([]byte(fileName))

	sendBuffer := make([]byte, config.BUFFERSIZE)
	fmt.Println("Start sending file!\n")

	var sentBytes int64
	start := time.Now()

	for {
		sentSize, err := file.Read(sendBuffer)
		if err != nil {
			break
		}

		stream.Write(sendBuffer)
		if err != nil {
			break
		}

		sentBytes += int64(sentSize)
		fmt.Printf("\033[2K\rSent: %d / %d", sentBytes, fileInfo.Size())
	}
	elapsed := time.Since(start)
	fmt.Println("\nTransfer took: ", elapsed)

	time.Sleep(2 * time.Second)
	fmt.Println("\n\nFile has been sent, closing stream!")
	return
}
