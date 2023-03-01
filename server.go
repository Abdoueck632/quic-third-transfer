package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
	"io"
	"log"
	"math"
	"os"
	"strings"

	quic "github.com/Abdoueck632/mp-quic"
	"strconv"
)

var CLIENTADDR = "10.0.3.2:4242"

var AddrServer = [2]string{"10.0.2.2:4242", "10.0.2.3:4242"}

func main() {
	dataMigration := config.DataMigration{}
	filename := make([]byte, 64)
	listener, err := quic.ListenAddr(config.Addr, utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("Server started! Waiting for streams from client...")

	sess, err := listener.Accept()
	utils.HandleError(err)

	stream, err := sess.AcceptStream()
	utils.HandleError(err)

	fmt.Println("session created: ", sess.RemoteAddr())

	dataMigration.IpAddr = fmt.Sprintf("%v", sess.RemoteAddr())

	//read filename of the client
	stream.Read(filename)
	dataMigration.FileName = strings.Trim(string(filename), ":")
	//lecture du fichier de sauvegarde pour les clés
	lines, err := readLines("/derivateK.in.txt")
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	//recupération des clés dans un tableau de bytes à 2 dimensions
	dataMigration.CrytoKey = stringTobytes2(lines)
	fmt.Printf("-----------> after conversion : %v", dataMigration.CrytoKey)

	//Récupération des des clés once obiet et id du serveur pour les envoyés au relay
	dataMigration.Once, dataMigration.Obit, dataMigration.Id = sess.GetCryptoSetup().GetOncesObitID()
	//send to the first server relay
	SendRelayData(AddrServer[0], dataMigration)

	/*time.Sleep(10 * time.Second)
	send to the second server relay
	sendRelayData(addrServer[1], filename1+".pt2", ipadd, newBytes)
	*/

	fmt.Printf("\n %+v \n ", dataMigration)
}

func SendRelayData(relayaddr string, dataMigration config.DataMigration) {

	sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())

	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)

	dataMigration.IpAddr = utils.FillString(dataMigration.IpAddr, 20)

	dataMigration.FileName = utils.FillString(dataMigration.FileName, 64) // par defaut fileInfo.Name()import socket

	//fmt.Println("session created: ", sess.RemoteAddr())

	fmt.Println("stream created...")
	fmt.Println("Client connected")

	if verifyOrder(sessServer, dataMigration.CrytoKey[2]) == true {
		dataMigration.CrytoKey[0], dataMigration.CrytoKey[2] = inverseByte(dataMigration.CrytoKey[0], dataMigration.CrytoKey[2])
		dataMigration.CrytoKey[1], dataMigration.CrytoKey[3] = inverseByte(dataMigration.CrytoKey[1], dataMigration.CrytoKey[3])
	}
	dataByte, err := json.Marshal(dataMigration)
	if err != nil {
		log.Fatal(err)
	}
	streamServer.Write([]byte(utils.FillString(string(dataByte), 1000)))
}
func verifyOrder(sess quic.Session, otherIV []byte) bool {
	forw, _, _ := sess.GetCryptoSetup().GetAEADs()
	if bytes.Equal(forw.GetOtherIV(), otherIV) == true {
		return true
	}
	return false
}
func inverseByte(first, second []byte) ([]byte, []byte) {
	return second, first
}
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
func stringTobytes(line string) []byte {
	return []byte(line)
}
func stringTobytes2(tab []string) [][]byte {
	var s [][]byte
	for _, mybte := range tab {
		s = append(s, stringTobytes(mybte))
	}

	return s
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
