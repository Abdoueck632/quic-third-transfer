package congestion

import "github.com/Abdoueck632/mp-quic/internal/protocol"

type connectionStats struct {
	slowstartPacketsLost protocol.PacketNumber
	slowstartBytesLost   protocol.ByteCount
}
