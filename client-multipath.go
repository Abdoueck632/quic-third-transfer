package main

import (
    "crypto/tls"
    "os"
    "io"
    "strconv"
    "time"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math" 
    "log"

    utils "./utils"
    config "./config"
    quic "github.com/lucas-clemente/quic-go"
)

const threshold = 5 * 1024  // 1KB
var CLIENTADDR="10.0.3.2:4242"
var quicConfig = &quic.Config{
    CreatePaths: true,
}
func main() {

    
    fileToSend := os.Args[1]
    
    addrServer := [2]string{"10.0.2.2:4242","10.0.2.3:4242"}
    fmt.Println(len(os.Args))

    
    addr := os.Args[2] + ":4242"

    fmt.Println("Server Address: ", addr)
    fmt.Println("Sending File: ", fileToSend)
    
    fmt.Println("Trying to connect to: ", addr)
    
    //fmt.Print("----------------la taille de filename est :",len(stream))

    SendAll(addr, fileToSend,false)
    
    go sendRelayData(addrServer[0],fileToSend+".pt1",addr)
    go sendRelayData(addrServer[1],fileToSend+".pt2",addr)
    
}

func sendRelayData(relayaddr string,filename string,addrServer string){
    
    sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, quicConfig)
    utils.HandleError(err)

    fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())
    
    streamServer, err := sessServer.OpenStream()
    utils.HandleError(err)
    addr1 := utils.FillString(addrServer, 14)
    fileName := utils.FillString(filename, 64) // par defaut fileInfo.Name()import socket

    fmt.Println("session created: ", sessServer.RemoteAddr())

    fmt.Println("stream created...")
    fmt.Println("Client connected")
    streamServer.Write([]byte(addr1))
    streamServer.Write([]byte(fileName))
    streamServer.Close()
    streamServer.Close()

}

func SendAll(addr string, fileToSend string, isRelay bool) {
    
   

    //fmt.Println("Size file : ",fileInfo.Size())

    

    sess, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, quicConfig)
    utils.HandleError(err)
    stream, err := sess.OpenStream()
    utils.HandleError(err) 
    fmt.Println("A client has connected!")
    defer stream.Close()

    file, err := os.Open(fileToSend)
    utils.HandleError(err)

    fileInfo, err := file.Stat()
    utils.HandleError(err)

    if fileInfo.Size() <= threshold {
        quicConfig.CreatePaths = false
        fmt.Println("File is small, using single path only.")
    } else {
        fmt.Println("file is large, using multipath now.")
    }

    fileSize := utils.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
    fileName := utils.FillString(fileInfo.Name(), 64)

    fmt.Println("Sending filename and filesize!")
    stream.Write([]byte(fileSize))
    stream.Write([]byte(fileName))
    if isRelay==true {
        SendData(stream,fileToSend,fileInfo.Size())
    }
    
    stream.Close()
    stream.Close()
    
}
func SendData(stream quic.Stream,fileToSend string,filesize int64){

    file, err := os.Open(fileToSend)
    utils.HandleError(err)

    sendBuffer := make([]byte, config.BUFFERSIZE)
    fmt.Println("Start sending file!   with buffersize = ", config.BUFFERSIZE," \n")

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
        fmt.Printf("\033[2K\rSent: %d / %d", sentBytes, filesize)
    }
    elapsed := time.Since(start)
    fmt.Println("\nTransfer took: ", elapsed)

   // stream.Close()
    //stream.Close()
   // time.Sleep(2 * time.Second)
    fmt.Println("\n\nFile has been sent, closing stream!")
}
func Hasher(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		log.Fatal(err)
	}
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
}
func Split(filename string, splitsize int) {
	bufferSize := 1024 // 1 KB for optimal splitting
	fileStats, _ := os.Stat(filename)
	pieces := int(math.Ceil(float64(fileStats.Size()) / float64(splitsize*1048576)))
	nTimes := int(math.Ceil(float64(splitsize*1048576) / float64(bufferSize)))
	file, err := os.Open(filename)
	hashFileName := filename + "-split-hash.txt"
	hashFile, err := os.OpenFile(hashFileName, os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	i := 1
	for i <= pieces {
		partFileName := filename + ".pt" + strconv.Itoa(i)
		pfile, _ := os.OpenFile(partFileName, os.O_CREATE|os.O_WRONLY, 0644)
		fmt.Println("Creating file:", partFileName)
		buffer := make([]byte, bufferSize)
		j := 1
		for j <= nTimes {
			_, inFileErr := file.Read(buffer)
			if inFileErr == io.EOF {
				break
			}
			_, err2 := pfile.Write(buffer)
			if err2 != nil {
				log.Fatal(err2)
			}
			j++
		}
		partFileHash := Hasher(partFileName)
		s := partFileName + ": " + partFileHash + "\n"
		hashFile.WriteString(s)
		pfile.Close()
		i++
	}
	s := "Original file hash: " + Hasher(filename) + "\n"
	hashFile.WriteString(s)
	file.Close()
	hashFile.Close()
	fmt.Printf("Splitted successfully! Find the individual file hashes in %s", hashFileName)
}
