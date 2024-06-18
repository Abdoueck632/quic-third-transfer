package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	quic "github.com/Abdoueck632/mp-quic"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
)

var dataString = make([]byte, 1000)
var ackImprimes = []uint64{0}

// var cpt = 1
var tabBuffer []int

type fileType struct {
	size int64
	data []byte
}

/*
	func main2() {
		var dataMigration config.DataMigration
		createConnectionToRelay("addrServer[1]", dataMigration)
	}
*/
func main() {
	savePath := os.Args[1]
	serverAddr := os.Args[2]

	//,
	sessClient, streamClient := acceptConnection(savePath, config.Addr)

	dataMigration := ReadDataMigration(streamClient)
	fmt.Printf("%+v", dataMigration)
	//sessServer, _ := acceptConnection(savePath, "0.0.0.0:4243")
	//ReadDataMigration(streamClient)

	_, streamServer := createConnexion(serverAddr)
	err := processFile(savePath, &dataMigration, sessClient, streamClient, streamServer)
	utils.HandleError(err)
	fmt.Printf("Data Migration: %+v\n", dataMigration)

}
func createConnexion(addr string) (quic.Session, quic.Stream) {
	sess, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	stream, err1 := sess.OpenStream()

	utils.HandleError(err1)
	fmt.Println("A server has connected!")

	stream.Write([]byte(utils.FillString("Ack", 10)))
	return sess, stream
}
func acceptConnection(savePath string, addr string) (quic.Session, quic.Stream) {
	fmt.Println("Saving file to: ", savePath)

	// Écoute des connexions entrantes
	listener, err := quic.ListenAddr(addr, utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)
	fmt.Println("Server started! Waiting for streams from client...")

	// Acceptation d'une nouvelle session
	sess, err := listener.Accept()
	utils.HandleError(err)
	fmt.Println("Session created: ", sess.RemoteAddr())

	// Acceptation d'un nouveau flux dans la session
	stream, err := sess.AcceptStream()
	utils.HandleError(err)
	fmt.Println("Stream created: ", stream.StreamID())
	fmt.Println("Connected to server, start receiving the file name and file size")

	return sess, stream
}
func contains(slice []uint64, x uint64) bool {
	for _, v := range slice {
		if v == x {
			return true
		}
	}

	return false
}
func processFile(savePath string, dataMigration *config.DataMigration, sess quic.Session, stream quic.Stream, streamServer quic.Stream) error {
	fmt.Println("Trying to connect to: ", dataMigration.IpAddr, "Filename ", dataMigration.FileName)
	dataMigration.TabBuffer = generateBufferIndices(dataMigration.TabBuffer[0], dataMigration.TabBuffer[1], config.BUFFERSIZE)
	name := savePath + dataMigration.FileName
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	SetCryptoSetup(sess, *dataMigration)
	stream.Setuint64(dataMigration.WritteOffset)

	fmt.Println("Sending filename and filesize!")
	stream.Setuint64(dataMigration.WritteOffset)
	time.Sleep(1 * time.Second)
	done := make(chan bool)
	//iteration := 0
	buffer := make([]byte, 100)
	// Fonction pour afficher les nouveaux ACK
	/*printNewAcks := func() {
		for {
			iteration++
			paquets := sess.GetAckPaquet()
			//fmt.Println(sess.GetAckPaquet())

			for _, paquet := range paquets {
				//streamServer.Write([]byte(utils.FillString("Ngagne SECK", 30)))
				if paquet.Ack == true {
					//fmt.Println(paquet.Offset)
					//fmt.Println(ackImprimes)
					if contains(ackImprimes, uint64(paquet.Offset)) == false {
						ackImprimes = append(ackImprimes, uint64(paquet.Offset))
						fmt.Println("1fhffffftgffttyddyttyy")

						ack, err := json.Marshal(config.Ack{Offset: uint64(paquet.Offset),
							IdRelay: dataMigration.IdRelay})
						if err != nil {
							log.Fatal(err)
						}
						streamServer.Write([]byte(utils.FillString(string(ack), 30)))

					}
				}
			}
		}
	}*/
	receivedck := func() {
		for {
			if size, _ := streamServer.Read(buffer); size > 0 {
				plageBuffer := config.PlageBuffer{}
				js := strings.Trim(string(buffer), ":")

				json.Unmarshal([]byte(js), &plageBuffer)
				fmt.Println("plageBuffer ", plageBuffer.TabBuffer)
				//fmt.Println("ma generation  ...", generateBufferIndices(plageBuffer.TabBuffer[0], plageBuffer.TabBuffer[1], config.BUFFERSIZE))
				tab := generateBufferIndices(plageBuffer.TabBuffer[0], plageBuffer.TabBuffer[1], config.BUFFERSIZE)
				//fmt.Println("Tab ", tab)

				dataMigration.TabBuffer = append(dataMigration.TabBuffer, tab...)
				fmt.Println("++")
				//fmt.Println("TabBuffer ", dataMigration.TabBuffer)

			}

		}
	}

	//go printNewAcks()
	go receivedck()
	go sendFile(stream, sess, dataMigration, file, done, streamServer)

	for {
		select {
		case <-done:
			fmt.Println("Transfer complete.")
			fmt.Println("la nouvelle structure", sess.GetAckPaquet())
			return nil
		}
	}

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
func generateBufferIndices(start, end, bufferSize int) []int {
	indices := []int{}
	for i := start; i <= end; i += bufferSize {
		indices = append(indices, i)
	}
	return indices
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

func sendFile(stream quic.Stream, sess quic.Session, dataMigration *config.DataMigration, file *os.File, done chan bool, streamServer quic.Stream) {
	defer close(done)

	//fileInfo, err := file.Stat()
	//utils.HandleError(err)

	sendBuffer := make([]byte, config.BUFFERSIZE)
	fmt.Println("Start sending file!\n")
	indice := 0
	var sentBytes int64
	//var c uint64
	start := time.Now()
	rateLimiter := time.NewTicker(config.THROTTLE_RATE) // Adjust THROTTLE_RATE for desired speed

	for {
		select {
		case <-done:
			return
		case <-rateLimiter.C: // Wait for rate limiter tick before reading/sending
			if indice < len(dataMigration.TabBuffer) {
				sentSize, err := file.ReadAt(sendBuffer, int64(dataMigration.TabBuffer[indice]))
				if sentSize == 0 {
					if err != nil {
						done <- false
						return
					}
					done <- true
					return
				}
				stream.Setuint64(uint64(74 + dataMigration.TabBuffer[indice]))
				stream.Write(sendBuffer)
				//dataMigration.StartAt += int64(sentSize) * int64(dataMigration.RelayNumber)
				//sentBytes += int64(sentSize)
				//_, _, c = stream.GetReadPosInFrame()
				//stream.Setuint64(c + uint64(sentSize))
				indice++
				//fmt.Printf("\033[2K\rSent: %d:  %d / %d  \n", cpt, dataMigration.StartAt, fileInfo.Size())
				//fmt.Println("la nouvelle structure", sess.GetAckPaquet())
			} else {
				//data := make([]byte, 10)
				//data = []byte(utils.FillString("sync", 10))
				//streamServer.Write(data)
				fmt.Println("Pas de données")
				fmt.Println("tp", dataMigration.TabBuffer)

			}
		}
	}

	elapsed := time.Since(start)
	fmt.Println("\nTransfer took: ", elapsed)

	fmt.Println("\n\n Thioune :File has been sent, closing stream! with ", sentBytes)
}
func sendBuffer(stream quic.Stream) {

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
	dataMigration.WritteOffset += config.BUFFERSIZE
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
	sessServer.SetIPAddress("127.0.0.1:4242", 1)
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

	err := sess.CreationRelayPath(dataMigration.IpAddr, fmt.Sprintf("%v", sess.LocalAddrById(1)), dataMigration.IdPathToCreate) // "10.0.2.2:4242"
	if err != nil {
		fmt.Println("Error ", err)
	}

	//fmt.Printf("%+v", sess.GetPathManager())

}
