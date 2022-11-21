package main

import (
	"abdou.seck/quic-third-transfer/config"
	"abdou.seck/quic-third-transfer/utils"
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	quic "github.com/Abdoueck632/mp-quic"
)

var FILENAME = ""
var NUMBERFILE = -1

// run programme file filename and the savepath in argument
func main() {

	fileToReceive := os.Args[1]
	savePath := os.Args[2]
	fmt.Println("Saving file to: ", savePath)
	serverAddr := "10.0.0.2:4242"

	fmt.Println("Attaching to: ", config.Addr)
	SendfileName(serverAddr, fileToReceive)
	listener, err := quic.ListenAddr(config.Addr, utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("Server started! Waiting for streams from client...")

	sess, err := listener.Accept()
	receiveFile(sess, savePath)

}
func receiveFile(sess quic.Session, savePath string) {
	//defer sess.Close()

	stream, err := sess.AcceptStream()
	defer stream.Close()

	fmt.Println("stream created: ", stream.StreamID())
	fmt.Println("session created: ", sess.RemoteAddr())

	fmt.Println("Connected to server, start receiving the file name and file size")
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)
	stream.Read(bufferFileSize)
	stream.Read(bufferFileName)

	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	fmt.Println("file size received: ", fileSize)

	fileName := strings.Trim(string(bufferFileName), ":")

	fmt.Println("file name received: ", fileName)

	newFile, err := os.Create(savePath + "/" + fileName)
	utils.HandleError(err)

	defer newFile.Close()
	var receivedBytes int64
	start := time.Now()
	for {
		if (fileSize - receivedBytes) < config.BUFFERSIZE {

			recv, err := io.CopyN(newFile, stream, (fileSize - receivedBytes))
			utils.HandleError(err)

			stream.Read(make([]byte, (receivedBytes+config.BUFFERSIZE)-fileSize))
			receivedBytes += recv
			fmt.Printf("\033[2K\rReceived: %d / %d", receivedBytes, fileSize)

			break
		}
		_, err := io.CopyN(newFile, stream, config.BUFFERSIZE)
		utils.HandleError(err)

		receivedBytes += config.BUFFERSIZE

		fmt.Printf("\033[2K\rReceived: %d / %d", receivedBytes, fileSize)
	}
	elapsed := time.Since(start)
	fmt.Println("\nTransfer took: ", elapsed)
	NUMBERFILE++
	if NUMBERFILE == 100 {
		FILENAME = savePath + FILENAME
		Join(FILENAME, 2)
		fmt.Println("ŒŒŒŒŒŒŒŒŒŒ Bravo SECK :)", FILENAME)
		NUMBERFILE = 0
	}
	time.Sleep(2 * time.Second)

	fmt.Println("\n\nReceived file completely!")
	fmt.Println("session created: ", sess.RemoteAddr())

}

func SendfileName(addr string, fileToSend string) {

	sess, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	stream, err1 := sess.OpenStream()
	defer stream.Close()
	utils.HandleError(err1)
	fmt.Println("A server has connected!")

	fileName := utils.FillString(fileToSend, 64)
	stream.Write([]byte(fileName))
	fmt.Println("Sending filename to the server! with filename ", fileName)

}

func Join(startFileName string, numberParts int) {
	a := len(startFileName)
	b := a // pat defaut -4
	iFileName := startFileName[:b]
	fmt.Println("--- FileName ", iFileName)
	_, err := os.Create(iFileName)
	jointFile, err := os.OpenFile(iFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatal(err)
	}
	i := 1
	for i <= numberParts {
		partFileName := iFileName + ".pt" + strconv.Itoa(i)
		fmt.Println("Processing file:", partFileName)
		pfile, _ := os.Open(partFileName)
		pfileinfo, err := pfile.Stat()
		if err != nil {
			log.Fatal(err)
		}
		pfilesize := pfileinfo.Size()
		pfileBytes := make([]byte, pfilesize)
		readSrc := bufio.NewReader(pfile)
		_, err = readSrc.Read(pfileBytes)
		if err != nil {
			log.Fatal(err)
		}
		_, err = jointFile.Write(pfileBytes)
		if err != nil {
			log.Fatal(err)
		}
		pfile.Close()
		jointFile.Sync()
		pfileBytes = nil
		i++
	}
	jointFile.Close()
	fmt.Printf("Combined successfully!")
}
