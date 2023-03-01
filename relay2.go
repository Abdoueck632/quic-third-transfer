package main

import (
	"encoding/json"
	"fmt"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
	"log"
	"os"
	"strings"
	"time"

	quic "github.com/Abdoueck632/mp-quic"
)

// var addrServer = [2]string{"10.0.2.2:4242", "10.0.2.3:4242"}

func main() {

	dataMigration := config.DataMigration{}
	dataString := make([]byte, 1000)

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

	//mt.Println("stream created: ", stream.StreamID())
	fmt.Println("Connected to server, start receiving the file name and file size")

	stream.Read(dataString)
	js := strings.Trim(string(dataString), ":")

	err = json.Unmarshal([]byte(js), &dataMigration)
	if err != nil {
		log.Fatal(err)
	}
	dataMigration.FileName = strings.Trim(string(dataMigration.FileName), ":")
	dataMigration.IpAddr = strings.Trim(string(dataMigration.IpAddr), ":")

	fmt.Printf(" \n dataMigration %+v \n ", dataMigration)
	SetCryptoSetup2(sess, dataMigration)

	fmt.Println("Trying to connect to: ", dataMigration.IpAddr, "Filename ", dataMigration.FileName)

	//use the first session with the client and this server
	fmt.Println("session created: ", sess.RemoteAddr())
	fmt.Println("stream created...")
	fmt.Println("Client connected")

	dataMigration.FileName = savePath + dataMigration.FileName

	fmt.Printf(" \n Packet Number %+v \n ", dataMigration.PacketNumber)

	SetParamter2(sess, dataMigration)
	//fmt.Println(GetPacketNumber(sess))
	stream.Setuint64(dataMigration.WritteOffset)
	sendFile2(stream, dataMigration)

}
func sendFile2(stream quic.Stream, dataMigration config.DataMigration) {

	/*stream, err := sess.OpenStream()
	utils.HandleError(err)
	fmt.Println("A client has connected!")

	*/

	file, err := os.Open(dataMigration.FileName)
	utils.HandleError(err)

	fileInfo, err := file.Stat()
	utils.HandleError(err)

	//fileSize := utils.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	//fileName := utils.FillString(fileInfo.Name(), 64)
	//stream.Read(nilbuffer)
	fmt.Println("Sending filename and filesize!")

	//stream.Write([]byte(fileSize))
	//stream.Write([]byte(fileName))

	sendBuffer := make([]byte, config.BUFFERSIZE)
	fmt.Println("Start sending file!\n")

	var sentBytes int64
	start := time.Now()

	for {

		sentSize, err := file.ReadAt(sendBuffer, dataMigration.StartAt)
		if sentSize == 0 {
			if err != nil {
				break
			}
			return
		}
		dataMigration.StartAt += int64(sentSize)
		sentBytes += int64(sentSize)

		stream.Write(sendBuffer)

		fmt.Printf("\033[2K\rSent: %d / %d  \n", sentBytes, fileInfo.Size())
		fmt.Printf("-------->>>> chaine %s \n ", string(sendBuffer))

	}
	elapsed := time.Since(start)
	fmt.Println("\nTransfer took: ", elapsed)

	fmt.Println("\n\nFile has been sent, closing stream!")
	fmt.Println("\n\n Size Send ", dataMigration.StartAt)

}
func SetParamter2(sess quic.Session, dataMigration config.DataMigration) {
	sess.GetpathsAndLen().GetpacketNumberGenerator().SetPacketNumber(dataMigration.PacketNumber["peek"])
	sess.GetpathsAndLen().SetlastSentPacketNumber(dataMigration.PacketNumber["packetsSent"], dataMigration.PacketNumber["lastSentPacketNumberSend"], dataMigration.PacketNumber["largestReceivedPacketWithAckSend"], dataMigration.PacketNumber["LargestAckedSend"], dataMigration.PacketNumber["lastRcvdPacketNumberPath"], dataMigration.PacketNumber["largestRcvdPacketNumberPath"])
	sess.GetpathsAndLen().SetlastRcvdPacketNumber(dataMigration.PacketNumber["largestObservedRcv"], dataMigration.PacketNumber["lowerLimitRcv"], dataMigration.PacketNumber["packetsRcv"], dataMigration.PacketNumber["LowerlastAckRcv"], dataMigration.PacketNumber["LarglastAckRcv"])

	sess.SetIPAddress(dataMigration.IpAddr)
}
func SetCryptoSetup2(sess quic.Session, dataMigration config.DataMigration) {
	sess.SetDerivateKey(dataMigration.CrytoKey[0], dataMigration.CrytoKey[1], dataMigration.CrytoKey[2], dataMigration.CrytoKey[3])
	sess.GetCryptoSetup().SetOncesObitID(dataMigration.Once, dataMigration.Obit, dataMigration.Id)
	sess.SetIPAddress(dataMigration.IpAddr)
}
