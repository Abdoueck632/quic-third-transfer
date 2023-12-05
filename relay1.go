package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	quic "github.com/Abdoueck632/mp-quic"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
)

var addrServer = [3]string{"10.144.208.212:4242", "10.144.208.213:4242", "10.144.208.213:4243"}
var dataString = make([]byte, 1000)
var cpt = 1

type fileType struct {
	size int64
	data []byte
}

func main2() {
	var dataMigration config.DataMigration
	createConnectionToRelay(addrServer[1], dataMigration)
}
func main() {
	dataMigration := config.DataMigration{}

	savePath := os.Args[1]
	fmt.Println("Saving file to: ", savePath)

	fmt.Println("Attaching to: ", config.Addr)
	listener, err := quic.ListenAddr(config.Addr, utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)
	/*f, err := os.Create("relay1_SSLKEYLOGFILE.bin")
	if err != nil {
		utils.HandleError(err)
	} else {
		defer f.Close()
	}

	*/

	fmt.Println("Server started! Waiting for streams from client...")

	sess, err := listener.Accept()
	utils.HandleError(err)

	fmt.Println("session created: ", sess.RemoteAddr())

	stream, err := sess.AcceptStream()
	utils.HandleError(err)

	fmt.Println("stream created: ", stream.StreamID())
	fmt.Println("Connected to server, start receiving the file name and file size")

	//sess.ClosePath(1)

	dataMigration = ReadDataMigration(stream)

	fmt.Printf(" \n dataMigration %+v \n ", dataMigration)

	fmt.Println("Trying to connect to: ", dataMigration.IpAddr, "Filename ", dataMigration.FileName)

	name := savePath + dataMigration.FileName
	file, err := os.Open(name)
	utils.HandleError(err)

	fileInfo, err := file.Stat()
	utils.HandleError(err)

	// Reconfigure the existing connection

	SetCryptoSetup(sess, dataMigration)
	//stream.Setuint64(dataMigration.WritteOffset)

	fileSize := utils.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := utils.FillString(fileInfo.Name(), 64)

	fmt.Println("Sending filename and filesize!")
	stream.Write([]byte(fileSize))
	stream.Write([]byte(fileName))
	//stream.Setuint64(dataMigration.WritteOffset)
	//_, _, dataMigration.WritteOffset = stream.GetReadPosInFrame()
	//dataMigration.StartAt = config.BUFFERSIZE
	createConnectionToRelay(addrServer[1], dataMigration)
	time.Sleep(1 * time.Second)
	sendFile(stream, dataMigration, file)

}
func activeListening(stream quic.Stream) {
	for {
		buffer := make([]byte, config.BUFFERSIZE)
		bytesRead, err := stream.Read(buffer)
		if err != nil {
			// Gérer l'erreur de lecture
			log.Fatal(err)
		}
		fmt.Println(string(bytesRead))
	}
}

func offsetManager(offset uint64) uint64 {
	return offset + uint64(config.BUFFERSIZE)
}
func ReadDataMigration(stream quic.Stream) config.DataMigration {
	var data = make([]byte, 1000)
	stream.Read(data)

	return myTrim(data)
}
func createNewLocalConnection() {
	fmt.Println("createNewLocalConnection")

	sessServer, err := quic.DialAddr("10.0.2.3:14242", &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with local server: ", sessServer.RemoteAddr())
	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)
	streamServer.Write(dataString)
	fmt.Println("stream created with local server...")
	fmt.Println("Client connected with local server")

	/*for {
		if sessServer.GetLenPaths() >= 2 {

			break
		}
	}
	dataByte, err := json.Marshal(migration)
	if err != nil {
		log.Fatal(err)
	}

	//sendBuffer := make([]byte, config.BUFFERSIZE)
	sentBytes, err := streamServer.Write([]byte(utils.FillString(string(dataByte), 1000)))
	utils.HandleError(err)

	*/
	//fmt.Printf("Sent to local server: %d / %d  \n", sentBytes)

	//	for {
	//		time.Sleep(1000 * time.Millisecond)
	//	}

}
func myTrim(dataString []byte) config.DataMigration {
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

func sendFile(stream quic.Stream, dataMigration config.DataMigration, file *os.File) (uint64, int64) {

	//stream, err := sess.OpenStream()
	//utils.HandleError(err)
	fmt.Println("A client has connected!")
	fileInfo, err := file.Stat()
	utils.HandleError(err)

	sendBuffer := make([]byte, config.BUFFERSIZE)
	fmt.Println("Start sending file!\n")

	var sentBytes int64
	var c uint64
	start := time.Now()

	for {

		sentSize, err := file.ReadAt(sendBuffer, dataMigration.StartAt)

		if sentSize == 0 {
			if err != nil {
				return 0, -1
			}

		}
		stream.Write(sendBuffer)

		dataMigration.StartAt += int64(sentSize) // + config.BUFFERSIZE
		sentBytes += int64(sentSize)
		_, _, c = stream.GetReadPosInFrame()
		//stream.Setuint64(c + uint64(sentSize))

		//fmt.Println("°°°°°°°°°°°°°°°°°°°°°°°°°°°° ", c)
		//fmt.Printf("-------->>>> chaine %s \n ", string(sendBuffer))
		fmt.Printf("\033[2K\rSent: %d:  %d / %d  \n", cpt, dataMigration.StartAt, fileInfo.Size())
	}

	elapsed := time.Since(start)
	fmt.Println("\nTransfer took: ", elapsed)

	fmt.Println("\n\n Thioune :File has been sent, closing stream! with ", sentBytes)

	return c, dataMigration.StartAt

}
func createSession(add string) (quic.Session, quic.Stream, error) {
	sess, err := quic.DialAddr(add, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sess.RemoteAddr())

	stream, err := sess.OpenStream()
	utils.HandleError(err)
	return sess, stream, err
}

func createConnectionToRelay(relayaddr string, dataMigration config.DataMigration) (quic.Stream, quic.Session) {
	sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	//fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())

	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)
	dataMigration.StartAt = config.BUFFERSIZE
	dataMigration.WritteOffset += config.BUFFERSIZE + 74
	dataByte, err := json.Marshal(dataMigration)
	data := []byte(utils.FillString(string(dataByte), 1000))
	streamServer.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	//time.Sleep(time.Second * 20)
	//	streamServer.Write(dataString)
	//streamServer.Write(dataString)
	fmt.Println("stream created...")
	fmt.Println("Client connected")

	return streamServer, sessServer
}
func SendDataToRelayAfterInitialisation(streamServer quic.Stream, dataMigration config.DataMigration) {

	dataByte, err := json.Marshal(dataMigration)
	if err != nil {
		log.Fatal(err)
	}
	streamServer.Write([]byte(utils.FillString(string(dataByte), 1000)))

}

func sendDataToRelay(streamToRelay quic.Stream, dataMigration config.DataMigration) {
	dataByte, err := json.Marshal(dataMigration)
	if err != nil {
		log.Fatal(err)
	}
	streamToRelay.Write([]byte(utils.FillString(string(dataByte), 1000)))

}

/*
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

*/

func sendFile3(stream quic.Stream, dataMigration config.DataMigration, name string) {

	/*stream, err := sess.OpenStream()
	utils.HandleError(err)
	fmt.Println("A client has connected!")

	*/

	file, err := os.Open(name)
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
		//fmt.Printf("-------->>>> chaine %s \n ", string(sendBuffer))

	}
	elapsed := time.Since(start)
	fmt.Println("\nTransfer took: ", elapsed)

	fmt.Println("\n\nFile has been sent, closing stream!")
	fmt.Println("\n\n Size Send ", dataMigration.StartAt)

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

/*
	func SetParameter(sess quic.Session, dataMigration config.DataMigration) {
		sess.GetpathsAndLen().GetpacketNumberGenerator().SetPacketNumber(dataMigration.PacketNumber["peek"])
		sess.GetpathsAndLen().SetlastSentPacketNumber(dataMigration.PacketNumber["packetsSent"], dataMigration.PacketNumber["lastSentPacketNumberSend"], dataMigration.PacketNumber["largestReceivedPacketWithAckSend"], dataMigration.PacketNumber["LargestAckedSend"], dataMigration.PacketNumber["lastRcvdPacketNumberPath"], dataMigration.PacketNumber["largestRcvdPacketNumberPath"])
		sess.GetpathsAndLen().SetlastRcvdPacketNumber(dataMigration.PacketNumber["largestObservedRcv"], dataMigration.PacketNumber["lowerLimitRcv"], dataMigration.PacketNumber["packetsRcv"], dataMigration.PacketNumber["LowerlastAckRcv"], dataMigration.PacketNumber["LarglastAckRcv"])
		//sess.SetIPAddress(dataMigration.IpAddr, 1)
		//sess.CreationRelayPath(dataMigration.IpAddr)

}
*/
func SetCryptoSetup(sess quic.Session, dataMigration config.DataMigration) {
	sess.SetDerivateKey(dataMigration.CrytoKey[0], dataMigration.CrytoKey[1], dataMigration.CrytoKey[2], dataMigration.CrytoKey[3])
	//sess.GetCryptoSetup().SetOncesObitID(dataMigration.Once, dataMigration.Obit, dataMigration.Id)
	//sess.SetIPAddress(dataMigration.IpAddr, 1)
	//sess.ClosePath(1)
	//sess.OpenPath(1)

	err := sess.CreationRelayPath(dataMigration.IpAddr, "10.0.2.2:4242", 2)
	if err != nil {
		fmt.Println("Error ", err)
	}

	//fmt.Printf("%+v", sess.GetPathManager())

}
