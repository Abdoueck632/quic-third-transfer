package main

import (

	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"log"
	"bufio"
	"fmt"

	utils "./utils"
	config "./config"
	quic "github.com/lucas-clemente/quic-go"
)

const addr = "0.0.0.0:" + config.PORT
var FILENAME=""
var NUMBERFILE=-1
func main() {

	savePath := os.Args[1]
	fmt.Println("Saving file to: ", savePath)
	serverAddr:="10.0.0.2"
	quicConfig := &quic.Config{
		CreatePaths: true,
	}

	fmt.Println("Attaching to: ", addr)
	listener, err := quic.ListenAddr(addr, utils.GenerateTLSConfig(), quicConfig)
	utils.HandleError(err)
    for{
		fmt.Println("Server started! Waiting for streams from client...")

		sess, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
 
		// If you want, you can increment a counter here and inject to handleClientRequest below as client identifier
		go receiveFile(sess,savePath,serverAddr)

		
	}
}
func receiveFile(sess quic.Session, savePath string,serverAddr string){
		//defer sess.Close()
	    fmt.Println("session created: ", sess.RemoteAddr())
		server:=fmt.Sprintf("%v",sess.RemoteAddr())
		server=strings.Split(server, ":")[0]
		stream, err := sess.AcceptStream()
		fmt.Println("--------- the ReadPosInFrame",stream.ReadPosInFrame())
		utils.HandleError(err)

		fmt.Println("stream created: ", stream.StreamID())

		defer stream.Close()
		fmt.Println("Connected to server, start receiving the file name and file size")
		bufferFileName := make([]byte, 64)
		bufferFileSize := make([]byte, 10)

		stream.Read(bufferFileSize)
		fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

		fmt.Println("file size received: ", fileSize)

		stream.Read(bufferFileName)
		fileName := strings.Trim(string(bufferFileName), ":")

		fmt.Println("file name received: ", fileName)
		
		if  server==serverAddr{

			NUMBERFILE=0
			FILENAME=fileName
			stream.Close()
			stream.Close()
			
			
		} else{
		

		
			newFile, err := os.Create(savePath + "/" + fileName)
			utils.HandleError(err)

			defer newFile.Close()
			var receivedBytes int64
			start := time.Now()

			for {
				if (fileSize - receivedBytes) < config.BUFFERSIZE {
				// fmt.Println("\nlast chunk of file.")

					recv, err := io.CopyN(newFile, stream, (fileSize - receivedBytes))
					utils.HandleError(err)

					stream.Read(make([]byte, (receivedBytes + config.BUFFERSIZE) - fileSize))
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
			if NUMBERFILE==2 {
				FILENAME=savePath + FILENAME
				Join(FILENAME,2)
				fmt.Println("ŒŒŒŒŒŒŒŒŒŒ Bravo SECK :)",FILENAME)
				NUMBERFILE=0
			}
			
				
			//time.Sleep(2 * time.Second)
			stream.Close()
			stream.Close()
			fmt.Println("\n\nReceived file completely!")
		}
}
func Join(startFileName string, numberParts int) {
	a := len(startFileName)
	b := a // pat defaut -4
	iFileName := startFileName[:b]
	fmt.Println("--- FileName ",iFileName)
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