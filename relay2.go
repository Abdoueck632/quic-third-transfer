package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"

	quic "github.com/Abdoueck632/mp-quic"
)

var cmpt = 1

func main2() {
	//sessLocalServer, streamLocalServer :=
	createNewConnection()
}
func main() {
	dataMigration := config.DataMigration{}

	savePath := os.Args[1]
	fmt.Println("Saving file to: ", savePath)

	fmt.Println("Attaching to: ", config.Addr)
	//sessLocalServer, streamLocalServer := createNewConnection()

	listener, err := quic.ListenAddr(config.Addr, utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("Server started! Waiting for streams from client...")

	sess, err := listener.Accept()
	utils.HandleError(err)

	stream, _ := sess.AcceptStream()
	dataMigration = ReadDataMigration1(stream)

	fmt.Println("Connected to server, start receiving the file name and file size")

	file, err := os.Open(savePath + dataMigration.FileName)
	utils.HandleError(err)

	fmt.Println("Trying to connect to: ", dataMigration.IpAddr, "Filename ", dataMigration.FileName)

	//use the first session with the client and this server
	fmt.Println("session from relay 1 created: ", dataMigration.IpAddr)
	SetCryptoSetup2(sess, dataMigration)
	stream.Setuint64(dataMigration.WritteOffset)

	time.Sleep(1 * time.Second)
	sendFile2(stream, dataMigration, file)

}
func ReadDataMigration1(stream quic.Stream) config.DataMigration {
	var data = make([]byte, 1000)
	stream.Read(data)

	return myTrim1(data)
}
func createNewConnection() (quic.Session, quic.Stream) {
	var dataString = make([]byte, 1000)
	listener, err := quic.ListenAddr("0.0.0.0:4242", utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)
	sess, err := listener.Accept()
	utils.HandleError(err)
	stream, err := sess.AcceptStream()
	utils.HandleError(err)
	fmt.Println("Server started! Waiting for streams from client...")

	//fmt.Println("session created: ", sess.RemoteAddr())

	stream.Read(dataString)
	for {
		buffer := make([]byte, config.BUFFERSIZE)

		/*if err != nil {
			// Gérer l'erreur de lecture
			log.Fatal(err)
		}

		*/
		if bytesRead, _ := stream.Read(buffer); bytesRead == 0 {
			fmt.Println("...")
		} else {
			fmt.Println("---ABDOU SECK----------------------", string(buffer))

		}

	}
	return sess, stream
}
func SendDataToRelayAfterInitialisation1(streamServer quic.Stream, dataMigration config.DataMigration) {

	dataByte, err := json.Marshal(dataMigration)
	if err != nil {
		log.Fatal(err)
	}
	streamServer.Write([]byte(utils.FillString(string(dataByte), 1000)))

}
func createNewNullConnection() {
	sessServer, err := quic.DialAddr("127.0.0.1:4242", &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())
	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)

	fmt.Println("stream created...")
	fmt.Println("Client connected")

	sendBuffer := make([]byte, config.BUFFERSIZE)
	sentBytes, err := streamServer.Write(sendBuffer)
	utils.HandleError(err)
	fmt.Printf("Sent: %d / %d  \n", sentBytes)

	go createDependentConnection()

	for {
		time.Sleep(1000 * time.Millisecond)
	}
}

func createDependentConnection() {
	fmt.Println("Starting dependent session func. waiting a bit...")
	//time.Sleep(10 * time.Second)
	sessServer, err := quic.DialAddr("127.0.0.1:4242", &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("dependent session created with secondary server: ", sessServer.RemoteAddr())
	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)

	fmt.Println("Dependent stream created...")
	fmt.Println("Dependent Client connected")

	sendBuffer := make([]byte, config.BUFFERSIZE)
	sentBytes, err := streamServer.Write(sendBuffer)
	utils.HandleError(err)
	fmt.Printf("Dependent Sent: %d / %d  \n", sentBytes)

}

func sendFile2(stream quic.Stream, dataMigration config.DataMigration, file *os.File) (uint64, int64) {

	/*stream, err := sess.OpenStream()
	utils.HandleError(err)
	fmt.Println("A client has connected!")

	*/

	//fileSize := utils.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	//sfileName := utils.FillString(fileInfo.Name(), 64)
	//stream.Read(nilbuffer)
	fmt.Println("Sending filename and filesize!")

	//stream.Write([]byte(fileSize))
	//stream.Write([]byte(fileName))

	sendBuffer := make([]byte, config.BUFFERSIZE)
	fmt.Println("Start sending file!\n")
	var c uint64
	var sentBytes int64
	start := time.Now()

	for {
		/*if sentBytes == 300*config.BUFFERSIZE {
			break
		}

		*/
		fmt.Println("yvvghbujnuygvfgv")

		sentSize, err := file.ReadAt(sendBuffer, dataMigration.StartAt)

		if sentSize == 0 {
			if err != nil {
				return 0, -1
			}

		}
		stream.Write(sendBuffer)

		sentBytes += int64(sentSize)

		fmt.Printf("\033[2K\rSent: %d:: %d -> %d  \n", cmpt, dataMigration.StartAt, dataMigration.StartAt+config.BUFFERSIZE)
		dataMigration.StartAt += int64(sentSize) + config.BUFFERSIZE
		_, _, c = stream.GetReadPosInFrame()
		stream.Setuint64(c + config.BUFFERSIZE)
		fmt.Printf("-------->>>> chaine %s \n ", string(sendBuffer))
		cmpt++
	}
	elapsed := time.Since(start)
	fmt.Println("\nTransfer took: ", elapsed)

	fmt.Println("\n\nFile has been sent, closing stream!")
	fmt.Println("\n\n Size Send ", dataMigration.StartAt)
	cmpt++
	return c, dataMigration.StartAt

}

/*
func sendRelayData2(sess quic.Session, streamServer quic.Stream, streamRelay quic.Stream, dataMigration config.DataMigration) (config.DataMigration, quic.Stream) {

	sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())

	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)



	fmt.Println("stream created...")
	fmt.Println("Client connected")

	dataByte, err := json.Marshal(dataMigration)
	if err != nil {
		log.Fatal(err)
	}
	streamRelay.Write([]byte(utils.FillString(string(dataByte), 1000)))

	for {

		if size, _ := streamRelay.Read(dataStrings); size > 0 {
			//streamServer.Read(dataStrings)
			break
		}
		fmt.Println("here ...")
	}

	return myTrim1(dataStrings), streamRelay

	/*
		dataByte, err := json.Marshal(dataMigration)
				if err != nil {
					log.Fatal(err)
				}
				stream.Write([]byte(utils.FillString(string(dataByte), 1000)))
				sess.ClosePath(2)
				for {

					if size, _ := stream.Read(dataStrings); size > 0 {

						*dataMigration = myTrim(dataString)
						break
					}
					fmt.Println("here ...")
				}
				return stream

}
*/
func myTrim1(dataString []byte) config.DataMigration {
	dataMigration := config.DataMigration{}
	js := strings.Trim(string(dataString), ":")

	err := json.Unmarshal([]byte(js), &dataMigration)
	if err != nil {
		log.Fatal(err)
	}
	dataMigration.FileName = strings.Trim(string(dataMigration.FileName), ":")
	dataMigration.IpAddr = strings.Trim(string(dataMigration.IpAddr), ":")
	return dataMigration
}

func SetCryptoSetup2(sess quic.Session, dataMigration config.DataMigration) {
	for {
		if sess.GetLenPaths() >= 2 {

			break
		}
	}

	sess.SetDerivateKey(dataMigration.CrytoKey[0], dataMigration.CrytoKey[1], dataMigration.CrytoKey[2], dataMigration.CrytoKey[3])
	//sess.SetIPAddress(dataMigration.IpAddr, 1)
	err := sess.CreationRelayPath(dataMigration.IpAddr, fmt.Sprintf("%v", sess.LocalAddrById(1)), 4)

	if err != nil {
		fmt.Println("Error ", err)
	}

	//fmt.Printf("%+v", sess.GetPathManager())

}
func createSession1(add string) (quic.Session, quic.Stream, error) {
	sess, err := quic.DialAddr(add, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sess.RemoteAddr())

	stream, err := sess.OpenStream()
	utils.HandleError(err)
	return sess, stream, err
}

/*
	func SetParamter2(sess quic.Session, dataMigration config.DataMigration) {
		sess.GetpathsAndLen().GetpacketNumberGenerator().SetPacketNumber(dataMigration.PacketNumber["peek"])
		sess.GetpathsAndLen().SetlastSentPacketNumber(dataMigration.PacketNumber["packetsSent"], dataMigration.PacketNumber["lastSentPacketNumberSend"], dataMigration.PacketNumber["largestReceivedPacketWithAckSend"], dataMigration.PacketNumber["LargestAckedSend"], dataMigration.PacketNumber["lastRcvdPacketNumberPath"], dataMigration.PacketNumber["largestRcvdPacketNumberPath"])
		sess.GetpathsAndLen().SetlastRcvdPacketNumber(dataMigration.PacketNumber["largestObservedRcv"], dataMigration.PacketNumber["lowerLimitRcv"], dataMigration.PacketNumber["packetsRcv"], dataMigration.PacketNumber["LowerlastAckRcv"], dataMigration.PacketNumber["LarglastAckRcv"])

		//sess.SetIPAddress(dataMigration.IpAddr, 1)
	}
*/
