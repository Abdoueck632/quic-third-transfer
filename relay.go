package main

import (
	"fmt"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
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
	newBytes16 := make([]byte, 16)
	newBytes4 := make([]byte, 4)
	newBytes16_2 := make([]byte, 16)
	newBytes4_2 := make([]byte, 4)
	var newBytes [][]byte

	//stream2 :=make([]byte,64)

	// var quic1 quic.Stream
	stream.Read(filename)
	stream.Read(addrClient)
	stream.Read(newBytes16)
	newBytes = append(newBytes, newBytes16)
	stream.Read(newBytes16_2)
	newBytes = append(newBytes, newBytes16_2)
	stream.Read(newBytes4)
	newBytes = append(newBytes, newBytes4)
	stream.Read(newBytes4_2)
	newBytes = append(newBytes, newBytes4_2)
	sess.SetDerivateKey(newBytes[0], newBytes[1], newBytes[2], newBytes[3])
	fmt.Printf("___________________________________________%v", newBytes)
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
	//use the first session with the client and this server
	/*
		sess1, err := quic.DialAddr(addrclient1, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
		utils.HandleError(err)

	*/
	// call SetIPAddress to modify the remote address in this session
	sess.SetIPAddress(addrclient1)
	//sess.CreationRelayPath(addrclient1)
	fmt.Println("session created: ", sess.RemoteAddr())
	/*stream1, err := sess.OpenStream()
	utils.HandleError(err)
	defer stream1.Close()

	*/

	fmt.Println("stream created...")
	fmt.Println("Client connected")
	sendFile(stream, name)
	time.Sleep(2 * time.Second)
	fmt.Println(sess.GetConnectionID())

	fmt.Println("-------------------------------------------")
	fmt.Println(sess.GetPaths())

}
func sendFile(stream quic.Stream, fileToSend string) {
	fmt.Println("A client has connected!")

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
