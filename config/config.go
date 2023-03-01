package config

import (
	quic "github.com/Abdoueck632/mp-quic"
)

// BUFFERSIZE is the
// size of max packet size
const BUFFERSIZE = 5

// PORT the default port for communication
const PORT = "4242"

// const SERVER_ADDR = "192.168.43.148:" + PORT
const Addr = "0.0.0.0:" + PORT
const Threshold = 5 * 1024 // 1KB
var QuicConfig = &quic.Config{
	CreatePaths: false,
}

type DataMigration struct {
	CrytoKey     [][]byte
	IpAddr       string
	FileName     string
	FileSize     int
	Once         []byte
	Obit         []byte
	Id           []byte
	PacketNumber map[string]uint64
	StartAt      int64
	WritteOffset uint64
}
