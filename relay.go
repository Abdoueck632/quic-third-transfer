package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	quic "github.com/Abdoueck632/mp-quic"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var addrServer = [2]string{"10.0.2.2:4242", "10.0.2.3:4242"}

type fileType struct {
	size int64
	data []byte
}

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

	fmt.Println("stream created: ", stream.StreamID())
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
	SetCryptoSetup(sess, dataMigration)

	fmt.Println("Trying to connect to: ", dataMigration.IpAddr, "Filename ", dataMigration.FileName)

	//use the first session with the client and this server
	fmt.Println("session created: ", sess.RemoteAddr())
	fmt.Println("stream created...")
	fmt.Println("Client connected")

	name := savePath + dataMigration.FileName
	dataMigration.WritteOffset = sendFile(stream, name)

	dataMigration.PacketNumber = GetPacketNumber(sess)
	sess.SetIPAddress("10.0.5.2:4243")

	fmt.Printf(" \n Packet Number %+v \n ", dataMigration)

	sendRelayData(addrServer[1], &dataMigration)

	/*time.Sleep(1 * time.Second)
	SetParamter(sess, dataMigration)
	//fmt.Println(GetPacketNumber(sess))

	sendFile3(stream, name, int64(300))

	*/

	//sendFile(sess, savePath+"groot.jpg")

	//sendRelayData(addrServer[1], "GoLand.zip", addrclient1, newBytes, once, obit, id, js)

}

func sendFile(stream quic.Stream, fileToSend string) uint64 {

	//stream, err := sess.OpenStream()
	//utils.HandleError(err)
	fmt.Println("A client has connected!")

	file, err := os.Open(fileToSend)
	utils.HandleError(err)

	fileInfo, err := file.Stat()
	utils.HandleError(err)

	fileSize := utils.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := utils.FillString(fileInfo.Name(), 64)
	//stream.Read(nilbuffer)
	fmt.Println("Sending filename and filesize!")

	stream.Write([]byte(fileSize))
	stream.Write([]byte(fileName))

	sendBuffer := make([]byte, config.BUFFERSIZE)
	fmt.Println("Start sending file!\n")

	var sentBytes int64
	var c uint64
	start := time.Now()

	for {
		if sentBytes == int64(300) {
			break
		}

		sentSize, err := file.Read(sendBuffer)

		if err != nil {
			break
		}

		stream.Write(sendBuffer)

		sentBytes += int64(sentSize)
		_, _, c = stream.GetReadPosInFrame()
		fmt.Println("°°°°°°°°°°°°°°°°°°°°°°°°°°°° ", c)
		fmt.Printf("-------->>>> chaine %s \n ", string(sendBuffer))
		fmt.Printf("\033[2K\rSent: %d / %d  \n", sentBytes, fileInfo.Size())

	}
	elapsed := time.Since(start)
	fmt.Println("\nTransfer took: ", elapsed)

	fmt.Println("\n\nFile has been sent, closing stream! with ", sentBytes)

	return c

}
func sendRelayData(relayaddr string, dataMigration *config.DataMigration) {

	sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())

	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)

	dataMigration.StartAt = int64(300)

	fmt.Println("stream created...")
	fmt.Println("Client connected")

	dataByte, err := json.Marshal(dataMigration)
	if err != nil {
		log.Fatal(err)
	}
	streamServer.Write([]byte(utils.FillString(string(dataByte), 1000)))

}
func GetPacketNumber(sess quic.Session) map[string]uint64 {
	path := sess.GetpathsAndLen()
	packetSent, a, b, c, d, e := path.GetlastSentPacketNumber()
	largesObse, lowerLimit, packetRcv, lastAck := path.GetRcvPacketNumber()
	LowerlastAckRcv, larglastAckRcv, _ := lastAck.GetAckedFrame()
	packetNumber := make(map[string]uint64)
	packetNumber["lastSentPacketNumberSend"] = uint64(a)
	packetNumber["LargestAckedSend"] = uint64(b)
	packetNumber["largestReceivedPacketWithAckSend"] = uint64(c)
	packetNumber["lastRcvdPacketNumberPath"] = uint64(d)
	packetNumber["largestRcvdPacketNumberPath"] = uint64(e)
	packetNumber["largestObservedRcv"] = uint64(largesObse)
	packetNumber["lowerLimitRcv"] = uint64(lowerLimit)
	packetNumber["packetsRcv"] = packetRcv
	packetNumber["packetsSent"] = packetSent
	packetNumber["LarglastAckRcv"] = uint64(larglastAckRcv)
	packetNumber["LowerlastAckRcv"] = uint64(LowerlastAckRcv)

	packetNumber["peek"] = uint64(path.GetpacketNumberGenerator().Peek())

	return packetNumber
}

func sendFile3(stream quic.Stream, fileToSend string, sizeSend int64) {

	/*stream, err := sess.OpenStream()
	utils.HandleError(err)
	fmt.Println("A client has connected!")

	*/

	file, err := os.Open(fileToSend)
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

		sentSize, err := file.ReadAt(sendBuffer, sizeSend)
		if sentSize == 0 {
			if err != nil {
				break
			}
			return
		}
		sizeSend += int64(sentSize)
		sentBytes += int64(sentSize)

		stream.Write(sendBuffer)

		fmt.Printf("\033[2K\rSent: %d / %d  \n", sentBytes, fileInfo.Size())
		fmt.Printf("-------->>>> chaine %s \n ", string(sendBuffer))

	}
	elapsed := time.Since(start)
	fmt.Println("\nTransfer took: ", elapsed)

	fmt.Println("\n\nFile has been sent, closing stream!")
	fmt.Println("\n\n Size Send ", sizeSend)
	stream.Close()
	stream.Close()

}
func fileToBytes(filename string) []fileType {

	var filemap []fileType
	file, err := os.Open(filename)
	utils.HandleError(err)

	fileInfo, err := file.Stat()
	utils.HandleError(err)
	sendBuffer := make([]byte, config.BUFFERSIZE)
	fmt.Println("Start saving file!\n")

	var sentBytes int64

	for {

		sentSize, err := file.Read(sendBuffer)
		if err != nil {
			break
		}

		filemap = append(filemap, fileType{int64(sentSize), sendBuffer})

		sentBytes += int64(sentSize)
		fmt.Printf("\033[2K\rSaving: %d / %d \n", sentBytes, fileInfo.Size())
	}
	//fmt.Println(filemap)
	return filemap
}
func SetParamter(sess quic.Session, dataMigration config.DataMigration) {
	sess.GetpathsAndLen().GetpacketNumberGenerator().SetPacketNumber(dataMigration.PacketNumber["peek"])
	sess.GetpathsAndLen().SetlastSentPacketNumber(dataMigration.PacketNumber["packetsSent"], dataMigration.PacketNumber["lastSentPacketNumberSend"], dataMigration.PacketNumber["largestReceivedPacketWithAckSend"], dataMigration.PacketNumber["LargestAckedSend"], dataMigration.PacketNumber["lastRcvdPacketNumberPath"], dataMigration.PacketNumber["largestRcvdPacketNumberPath"])
	sess.GetpathsAndLen().SetlastRcvdPacketNumber(dataMigration.PacketNumber["largestObservedRcv"], dataMigration.PacketNumber["lowerLimitRcv"], dataMigration.PacketNumber["packetsRcv"], dataMigration.PacketNumber["LowerlastAckRcv"], dataMigration.PacketNumber["LarglastAckRcv"])
	sess.SetIPAddress(dataMigration.IpAddr)
}
func SetCryptoSetup(sess quic.Session, dataMigration config.DataMigration) {
	sess.SetDerivateKey(dataMigration.CrytoKey[0], dataMigration.CrytoKey[1], dataMigration.CrytoKey[2], dataMigration.CrytoKey[3])
	sess.GetCryptoSetup().SetOncesObitID(dataMigration.Once, dataMigration.Obit, dataMigration.Id)
	sess.SetIPAddress(dataMigration.IpAddr)
}
