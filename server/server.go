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
    //"strings"
    "encoding/json"
    
    
    
    utils "./utils"
    config "./config"
    quic "github.com/lucas-clemente/quic-go"
   
)

const threshold = 5 * 1024  // 1KB
const addr = "0.0.0.0:" + config.PORT
var CLIENTADDR="10.0.3.2:4242"
var quicConfig = &quic.Config{
    CreatePaths: false,
}
var addrServer = [2]string{"10.0.2.2:4242","10.0.2.3:4242"}

func main() {

    
   
    
    
    fmt.Println(len(os.Args))

    
    //addrClient := "10.0.3.2:4242"
    //sendRelayData(addrServer[0],"go.zip.pt1",nil)
    _,sess:=WaitClientRequest()
   
    //Split(fileToSend,64)
    fmt.Println("----------------le filename est :",sess)
    
    path:=sess.GetPaths()
    _= sess.SendPing(path[0])

    _,packer,crypto:=sess.GetIdConnAndpacker()
    
	fmt.Printf("\n packer : %+v", packer)
	fmt.Printf("\n crypto : %+v", crypto)
    a,b,c,d,e,f,g:=crypto.GetCrypto()

    fmt.Printf("\n --- : %+v", a)
    fmt.Printf("\n --- : %+v", b)
    fmt.Printf("\n --- : %+v", c)
    fmt.Printf("\n --- : %+v", d())
    fmt.Printf("\n --- : %+v", e)
    fmt.Printf("\n --- : %+v", f)
    fmt.Printf("\n --- : %+v", g)   
    
    
    
     /*sendRelayData(addrServer[0],fileToSend+".pt1",addrClient,sess,sfcg,idc)
    
   sendRelayData(addrServer[1],fileToSend+".pt2",addrClient,sess,sfcg,idc)*/
    //SendAll(fileToSend,sess) 
    fmt.Printf("---------------- %+v",sess)
    
}
func WaitClientRequest() (string,quic.Session){

    listener,kex, err := quic.ListenAddr(addr, utils.GenerateTLSConfig(), quicConfig)
	utils.HandleError(err)
    fmt.Println(kex)
    fmt.Println("Server started! Waiting for streams from client...")

    sess, err := listener.Accept()
    utils.HandleError(err)
    stream, err := sess.AcceptStream()
    utils.HandleError(err)
    
    fmt.Println("session created: ", sess.RemoteAddr())
    
    fileToSend:="go.zip"
    
   
    sendRelayData(addrServer[0],fileToSend+".pt1",sess)

    stream.Close()
    stream.Close()

    //SendAll(fileToSend, sess)
    //sendRelayData(addrServer[1],fileToSend+".pt2",addres,sess,addres,idc)
   /*  stream, err := sess.AcceptStream()
    utils.HandleError(err)
    fmt.Println("stream created: ", stream.StreamID())

   
    fmt.Println("Connected to client, start receiving the file name ")
    filename := make([]byte,64)
    
    // var quic1 quic.Stream
    stream.Read(filename)
    filename1 := strings.Trim(string(filename), ":")
    path:=sess.GetPaths()
    _= sess.SendPing(path[0])
    //SendAll(filename1,sess)
    
    /*stream.Write(filename)
    
    fmt.Printf("\n %+s\n", sess)
    fmt.Println(sess.GetIdConn())*/
    //idc:=fmt.Sprintf("%v", sess.GetIdConn())
    
    

    
    return fileToSend,sess
        
}

func sendRelayData(relayaddr string,filename string,sess quic.Session){
    
    sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, quicConfig)
    utils.HandleError(err)

    fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())
    idcon,_,_:=sess.GetIdConnAndpacker()
    mystruct:=sess.InitializeMyStrut(idcon,sess.RemoteAddr())
    castStruct:= sess.ToGOB64(mystruct)
    streamServer, err := sessServer.OpenStream()
    utils.HandleError(err)
    
    b, err := json.Marshal(mystruct)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("-------> ",string(b))
    
    fmt.Printf("----  my struct: %s\n", castStruct)

   
    ipadd:=fmt.Sprintf("%s",sess.RemoteAddr())
    ipadre:=utils.FillString(ipadd, 20)
    idconn:=utils.FillString(string(b), 500)
    fileName := utils.FillString(filename, 64) // par defaut fileInfo.Name()import socket

    fmt.Println("session created: ", sessServer.RemoteAddr())

    fmt.Println("stream created...")
    fmt.Println("Client connected")
   
    streamServer.Write([]byte(fileName))
    
    streamServer.Write([]byte(idconn))
    streamServer.Write([]byte(ipadre))
    streamServer.Close()
    streamServer.Close()

   

}

func SendAll(fileToSend string,sess quic.Session) {
    
   

    //fmt.Println("Size file : ",fileInfo.Size())

    

    /*sess, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, quicConfig)
    utils.HandleError(err)
    fmt.Printf("\n struct server: %+v \n", sess)
    fmt.Println("--------------------------")
    
    paths:=sess.GetPaths()
    for index := range paths  {
		fmt.Println("Le path de ", index, "est", *paths[index])
	}
    */
    stream, err := sess.OpenStream()
    utils.HandleError(err) 
    fmt.Println("A client has connected!")
    
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
    
    
    SendData(stream,fileToSend,fileInfo.Size())

    
    /*stream.Close()
    stream.Close()*/
    
    
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

    stream.Close()
    stream.Close()
    time.Sleep(2 * time.Second)
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
