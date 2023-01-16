package utils

import (
	"encoding/binary"

	"github.com/Abdoueck632/mp-quic/internal/protocol"
)

// GenerateConnectionID generates a connection ID using cryptographic not random, fix connectionID for all session
func GenerateConnectionID() (protocol.ConnectionID, error) {
	b := make([]byte, 8)
	/*_, err := rand.Read(b)
	if err != nil {
		return 0, err
	}

	*/
	b = []byte{245, 27, 40, 248, 54, 246, 198, 61}
	return protocol.ConnectionID(binary.LittleEndian.Uint64(b)), nil
}
