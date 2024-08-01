package config

import (
	"fmt"
	"os"
	"time"

	quic "github.com/Abdoueck632/mp-quic"
)

// BUFFERSIZE is the
// size of max packet size
const BUFFERSIZE = 1000
const THROTTLE_RATE = 200 * time.Millisecond

// PORT the default port for communication
const PORT = "4242"

// const SERVER_ADDR = "192.168.43.148:" + PORT
const Addr = "0.0.0.0:" + PORT
const Threshold = 5 * 1024 // 1KB
var QuicConfig = &quic.Config{
	CreatePaths: true,
	//	CacheHandshake: true,
	//IdleTimeout:      10000 * time.Hour,
	//HandshakeTimeout: 10000 * time.Hour,
	//MaxReceiveConnectionFlowControlWindow: uint64(80000),
	//MaxReceiveStreamFlowControlWindow:     uint64(9000),
}

//var QuicConfigServer = &quic.Config{
//	CreatePaths:    true,
//	CacheHandshake: false,
//	IdleTimeout:    10000 * time.Hour,
//}

type DataMigration struct {
	CrytoKey        [][]byte
	IpAddr          string
	FileName        string
	FileSize        int
	Once            []byte
	Obit            []byte
	Id              []byte
	PacketNumber    map[string]uint64
	StartAt         int64
	WritteOffset    uint64
	MustSynchronise bool
	RelayNumber     int
	DataBeforeSend  uint64
	IdPathToCreate  int
	TabBuffer       []int
	IdRelay         int
}
type Ack struct {
	Offset  uint64
	IdRelay int
}
type PlageBuffer struct {
	TabBuffer []int
}

func WriteFile(fileName string, data string) {
	//fileName := "example.txt"

	// Ouvrir le fichier en mode écriture (création si n'existe pas, vidage si existe)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture du fichier : %v\n", err)
		return
	}
	defer file.Close()

	// Données à écrire dans le fichier
	//data := "Bonjour, ceci est un exemple d'écriture dans un fichier en Go.\n"

	// Écrire les données dans le fichier
	_, err = file.WriteString(data)
	if err != nil {
		fmt.Printf("Erreur lors de l'écriture dans le fichier : %v\n", err)
		return
	}

	fmt.Println("Écriture réussie dans le fichier.")
}
