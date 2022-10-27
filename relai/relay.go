package main

import (
	"fmt"
	//"crypto/tls"
	"os"
	"strconv"
	"time"
        "strings"
	//"bytes"
	//god "encoding/gob"
    "encoding/json"
    
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
	listener,_, err := quic.ListenAddr(addr, utils.GenerateTLSConfig(), quicConfig)
	utils.HandleError(err)

    fmt.Println("Server started! Waiting for streams from client...")

    sess, err := listener.Accept()
    utils.HandleError(err)

    fmt.Println("session created: ", sess.RemoteAddr())

    stream, err := sess.AcceptStream()
    utils.HandleError(err)

    fmt.Println("stream created: ", stream.StreamID())
    b, _ := json.Marshal(sess)
    //a:=fmt.Sprintf("\n %+v\n", sess)
    fmt.Println("stream created...",string(b))
    defer stream.Close()
    fmt.Println("Connected to server, start receiving the file name and file size")
    filename := make([]byte,64)
    idconn := make([]byte, 500)
    ipadre := make([]byte, 20)
    
    stream.Read(filename)
    
    
    
    
    // var quic1 quic.Stream
    stream.Read(idconn)
    stream.Read(ipadre)


    // a,_,_,_:=listener.GetAttribut()
    //c:=a.GetAttribut()
    //fmt.Printf("\n new struct : %+v :\n",a.FromGOB64(sfcg))


    
    //stream.Read(stream2)
    // quic1:= quic.Stream(stream2)
    
    //fmt.Println("_______________hello ", quic1)
    stream.Close()
    stream.Close()

    filename1 := strings.Trim(string(filename), ":")
    ipaddr := strings.Trim(string(ipadre), ":")
    idconn2 := strings.Trim(string(idconn), ":")
    //stringMystruct := strings.Trim(string(idconn), ":")
    
    var e1Converted quic.MyStruct
    err = json.Unmarshal([]byte(idconn2), &e1Converted)
    if err != nil {
        fmt.Printf("Error occured during unmarshaling. Error: %s", err.Error())
    }
    fmt.Printf("MyStruct Struct: %s\n", e1Converted.IdConnection)
    //mystruct:=sess.FromGOB64(stringMystruct)
    fmt.Printf("----  my struct: %+v\n %s\n", e1Converted)
    

    fmt.Println("---- filename :",filename1," and adresse client ",ipaddr)
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
    
    
    //sess1, err := quic.DialAddr(addrclient1, &tls.Config{InsecureSkipVerify: true}, quicConfig)
    //utils.HandleError(err)
    sess.SetIPAddress(ipaddr)
    
    
    
    sess.SetIdConn(e1Converted.IdConnection)
    //fmt.Println(sess.GetIdConn())
    fmt.Println("session created: ", sess.RemoteAddr())

    //stream1, err := sess.OpenStream()
    utils.HandleError(err)
    
    fmt.Println("Client connected")
    stream.Close()
    stream.Close()
    /*path:=sess.GetPaths()
    err1:= sess.SendPing(path[0])
    fmt.Printf("struct path: %+v \n",path[0])
    fmt.Println("Ping to client : ",err1)*/

    sendFile(sess,name)
        

	
}
func sendFile(sess quic.Session, fileToSend string) {

    stream, err := sess.AcceptStream()
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

    stream.Close()
    stream.Close()
    time.Sleep(2 * time.Second)
    fmt.Println("\n\nFile has been sent, closing stream!")
    
    return
}
