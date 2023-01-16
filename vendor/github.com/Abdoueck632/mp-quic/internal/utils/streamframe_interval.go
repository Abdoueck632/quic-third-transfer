package utils

import "github.com/Abdoueck632/mp-quic/internal/protocol"

// ByteInterval is an interval from one ByteCount to the other
// +gen linkedlist
type ByteInterval struct {
	Start protocol.ByteCount
	End   protocol.ByteCount
}
