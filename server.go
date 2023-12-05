package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	quic "github.com/Abdoueck632/mp-quic"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
)

var CLIENTADDR = "10.0.3.2:4242"

func main() {
	AddrServer := os.Args[1]
	dataMigration := config.DataMigration{}
	filename := make([]byte, 64)
	listener, err := quic.ListenAddr(config.Addr, utils.GenerateTLSConfig(), config.QuicConfig)
	utils.HandleError(err)
	f, err := os.Create("serveur_SSLKEYLOGFILE.bin")
	if err != nil {
		utils.HandleError(err)
	} else {
		defer f.Close()
	}

	fmt.Println("Server started! Waiting for streams from client...")

	sess, err := listener.Accept()
	utils.HandleError(err)

	stream, err := sess.AcceptStream()
	utils.HandleError(err)

	fmt.Println("session created: ", sess.RemoteAddr())

	//read filename of the client
	stream.Read(filename)
	//sess.ClosePath(0)

	dataMigration.FileName = strings.Trim(string(filename), ":")

	if err != nil {
		log.Fatalf("loadDerivedKeys: %s", err)
	}
	//time.Sleep(10 * time.Second)
	//send to the first server relay

	lines, err := loadDerivedKeys("/derivateK.in.json")
	dataMigration.CrytoKey = lines
	fmt.Println(dataMigration)
	//	name := "./storage-server/" + dataMigration.FileName
	//file, err := os.Open(name)

	//fileInfo, err := file.Stat()

	//fileSize := utils.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	//fileName := utils.FillString(fileInfo.Name(), 64)
	//stream.Write([]byte(fileSize))
	//stream.Write([]byte(fileName))
	//dataMigration.WritteOffset = 74
	SendRelayData(AddrServer, dataMigration, sess)
	//dataMigration.StartAt = config.BUFFERSIZE
	//dataMigration.WritteOffset += config.BUFFERSIZE
	//SendRelayData(AddrServer[1], dataMigration, sess)
	//sess.ClosePath(0)
	//	sess.AdvertiseAddress(AddrServer[0])

	//time.Sleep(2 * time.Second)

	/*time.Sleep(10 * time.Second)
	send to the second server relay
	sendRelayData(addrServer[1], filename1+".pt2", ipadd, newBytes)
	*/

}

func SendRelayData(relayaddr string, dataMigration config.DataMigration, sess quic.Session) {

	dataMigration.Once, dataMigration.Obit, dataMigration.Id = sess.GetCryptoSetup().GetOncesObitID()

	sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())

	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)
	fmt.Printf(" œœœœœœœœœœœœœœœœœœœœœœœœœ %v", sess.RemoteAddrById(1))
	for {
		if sess.GetLenPaths() == 2 {
			break
		}
	}
	dataMigration.IpAddr = fmt.Sprintf("%v", sess.RemoteAddrById(1))
	dataMigration.IpAddr = utils.FillString(dataMigration.IpAddr, 20)

	dataMigration.FileName = utils.FillString(dataMigration.FileName, 64) // par defaut fileInfo.Name()import socket

	//fmt.Println("session created: ", sess.RemoteAddr())

	fmt.Println("stream created...")
	fmt.Println("Client connected")

	if verifyOrder(sess, dataMigration.CrytoKey[2]) != true {
		fmt.Println("False in verification")
		dataMigration.CrytoKey[0], dataMigration.CrytoKey[1] = inverseByte(dataMigration.CrytoKey[0], dataMigration.CrytoKey[1])
		dataMigration.CrytoKey[2], dataMigration.CrytoKey[3] = inverseByte(dataMigration.CrytoKey[2], dataMigration.CrytoKey[3])
	}

	dataByte, err := json.Marshal(dataMigration)
	if err != nil {
		log.Fatal(err)
	}

	streamServer.Write([]byte(utils.FillString(string(dataByte), 1000)))
	fmt.Println("%+v", dataMigration)

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

// writeLines writes the lines to the given file.
func saveDerivedKeys(data [][]byte, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Parcourt le tableau et écrit chaque élément dans le fichier
	dataByte, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write(dataByte)

	return err
}

func loadDerivedKeys(path string) ([][]byte, error) {
	datas, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	//defer datas.Close()

	// Sépare le fichier en lignes
	var derivedKeys [][]byte
	json.Unmarshal(datas, &derivedKeys)

	fmt.Printf("%v\n", derivedKeys)
	return derivedKeys, nil
}
func stringTobytes(line string) []byte {
	return []byte(line)
}
func stringTobytes2(tab []string) [][]byte {
	var s [][]byte
	for _, mybte := range tab {
		s = append(s, stringTobytes(mybte))
	}
	fmt.Println(s)
	return s
}
func convertStringSliceToByteSliceSlice(s []string) [][]byte {
	var result [][]byte
	for _, str := range s {
		var bytes []byte
		for _, r := range []rune(str) {
			buf := make([]byte, utf8.RuneLen(r))
			utf8.EncodeRune(buf, r)
			bytes = append(bytes, buf...)
		}
		result = append(result, bytes)
	}
	return result
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
