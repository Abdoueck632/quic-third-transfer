package main

import (
	"fmt"
	"crypto/tls"
	"os"
	"strconv"
	"time"
        "strings"
	//"bytes"
	//god "encoding/gob"
	utils "./utils"
	config "./config"
	quic "github.com/lucas-clemente/quic-go"
)

const addr = "0.0.0.0:" + config.PORT
const threshold = 5 * 1024  // 1KB
 
func main() {

	savePath := os.Args[1]
	fmt.Println("Saving file to: ", savePath)

	quicConfig := &quic.Config{
		CreatePaths: true,
	}

	fmt.Println("Attaching to: ", addr)
	listener, err := quic.ListenAddr(addr, utils.GenerateTLSConfig(), quicConfig)
	utils.HandleError(err)
  
    for{

   
        fmt.Println("Server started! Waiting for streams from client...")

        sess, err := listener.Accept()
        utils.HandleError(err)

        fmt.Println("session created: ", sess.RemoteAddr())

        stream, err := sess.AcceptStream()
        utils.HandleError(err)

        fmt.Println("stream created: ", stream.StreamID())

        defer stream.Close()
        fmt.Println("Connected to server, start receiving the file name and file size")
        filename := make([]byte,64)
        addrClient := make([]byte, 14)
            //stream2 :=make([]byte,64)
            stream.Read(addrClient)
            
            addrclient1 := strings.Trim(string(addrClient), ":")
            
        // var quic1 quic.Stream
        stream.Read(filename)
        //stream.Read(stream2)
        // quic1:= quic.Stream(stream2)
        
        //fmt.Println("_______________hello ", quic1)
            stream.Close()
            stream.Close()
        
            filename1 := strings.Trim(string(filename), ":")
            fmt.Println("---- filename :",filename1," and adresse client ",addrclient1)
            fmt.Print("----------------la taille de filename est :",len(filename1))
            name :=savePath+filename1
            file, err := os.Open(name)
            utils.HandleError(err)
            
        fileInfo, err := file.Stat()

            utils.HandleError(err)	
        if fileInfo.Size() <= threshold {
                quicConfig.CreatePaths = false
                fmt.Println("File is small, using single path only.")
        } else {
                fmt.Println("file is large, using multipath now.")
        }
            file.Close()
            
            fmt.Println("Trying to connect to: ", addrClient)
            sess1, err := quic.DialAddr(addrclient1, &tls.Config{InsecureSkipVerify: true}, quicConfig)
            utils.HandleError(err)

            fmt.Println("session created: ", sess.RemoteAddr())

            stream1, err := sess1.OpenStream()
            utils.HandleError(err)

            fmt.Println("stream created...")
            fmt.Println("Client connected")
            sendFile(stream1,name)
            time.Sleep(2 * time.Second)
    }
	
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

    stream.Close()
    stream.Close()
    time.Sleep(2 * time.Second)
    fmt.Println("\n\nFile has been sent, closing stream!")
    
    return
}
