package main

import (
    "crypto/tls"
    "fmt"
    "os"
    "strconv"
    "time"
    "encoding/json"  

    utils "./utils"
    config "./config"
    quic "github.com/lucas-clemente/quic-go"
)

const threshold = 5 * 1024  // 1KB

func main() {

    quicConfig := &quic.Config{
        CreatePaths: true,
    }

    fileToSend := os.Args[1]
    addr := os.Args[2] + ":4242"
    addrServer := "10.0.2.2:4242"
    
    fmt.Println("Server Address: ", addr)
    fmt.Println("Sending File: ", fileToSend)

    file, err := os.Open(fileToSend)
    utils.HandleError(err)

    fileInfo, err := file.Stat()
    utils.HandleError(err)
    fmt.Println("Size file : ",fileInfo.Size())
    if fileInfo.Size() <= threshold {
        quicConfig.CreatePaths = false
        fmt.Println("File is small, using single path only.")
    } else {
        fmt.Println("file is large, using multipath now.")
    }
    fileName := utils.FillString(fileInfo.Name(), 64)
    file.Close()

    fmt.Println("Trying to connect to: ", addr ,"and 10.0.2.2")
    sessServer, err := quic.DialAddr(addrServer, &tls.Config{InsecureSkipVerify: true}, quicConfig)
    sess, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, quicConfig)
    utils.HandleError(err)
    stream, err := sess.OpenStream()
    utils.HandleError(err) 
    

    b1, err := json.Marshal(&stream)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(string(b1))


    fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())
    fmt.Println("session created: ", sess.RemoteAddr())
    //fmt.Print("----------------la taille de filename est :",len(stream))
    streamServer, err := sessServer.OpenStream()
    utils.HandleError(err)
    addr1 := utils.FillString(addr, 14)
    b, err := json.Marshal(streamServer)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("-----",string(b))
    
    fmt.Println("stream created...")
    fmt.Println("Client connected")

   
    sendFile(stream, fileToSend)
   
    streamServer.Write([]byte(addr1))
    streamServer.Write([]byte(fileName))
 
    //streamServer.Write(bin_buf,binary.BigEndian,stream)
   
    streamServer.Close()
    streamServer.Close()


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
        fmt.Printf("\033[2K\rSent: %d / %d", sentBytes, fileInfo.Size())
    }
    elapsed := time.Since(start)
    fmt.Println("\nTransfer took: ", elapsed)

   // stream.Close()
    //stream.Close()
   // time.Sleep(2 * time.Second)
    fmt.Println("\n\nFile has been sent, closing stream!")
    return
}
