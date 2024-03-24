package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	quic "github.com/Abdoueck632/mp-quic"
	"github.com/Abdoueck632/quic-third-transfer/config"
	"github.com/Abdoueck632/quic-third-transfer/utils"
)

func main() {
	AddrRelay1 := os.Args[1]
	AddrRelay2 := os.Args[2]
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
	for {
		if sess.GetLenPaths() == 2 {
			break
		}
	}
	lines, err := loadDerivedKeys("/derivateK.in.json")
	dataMigration.CrytoKey = lines
	name := "./storage-server/" + dataMigration.FileName
	file, err := os.Open(name)

	fileInfo, err := file.Stat()

	fileSize := utils.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := utils.FillString(fileInfo.Name(), 64)
	stream.Write([]byte(fileSize))
	stream.Write([]byte(fileName))
	dataMigration.WritteOffset = 74
	SendRelayData(AddrRelay1, dataMigration, sess, 2)

	dataMigration.WritteOffset += config.BUFFERSIZE
	dataMigration.StartAt = config.BUFFERSIZE
	SendRelayData(AddrRelay2, dataMigration, sess, 4)

}

func SendRelayData(relayaddr string, dataMigration config.DataMigration, sess quic.Session, idpath int) {

	dataMigration.IpAddr = fmt.Sprintf("%v", sess.RemoteAddrById(1))

	dataMigration.Once, dataMigration.Obit, dataMigration.Id = sess.GetCryptoSetup().GetOncesObitID()
	dataMigration.RelayNumber = 2
	sessServer, err := quic.DialAddr(relayaddr, &tls.Config{InsecureSkipVerify: true}, config.QuicConfig)
	utils.HandleError(err)

	fmt.Println("session created with secondary server: ", sessServer.RemoteAddr())

	streamServer, err := sessServer.OpenStream()
	utils.HandleError(err)

	dataMigration.IpAddr = utils.FillString(dataMigration.IpAddr, 20)

	dataMigration.FileName = utils.FillString(dataMigration.FileName, 64) // par defaut fileInfo.Name()import socket
	dataMigration.IdPathToCreate = idpath
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
