package wire

import "github.com/Abdoueck632/mp-quic/internal/protocol"

// AckRange is an ACK range
type AckRange struct {
	First protocol.PacketNumber
	Last  protocol.PacketNumber
}
