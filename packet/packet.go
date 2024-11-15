package packet

import (
	"errors"
	"io"
)

const MAGIC_NUMBER = 0xa1

func Pack(buffer []byte, connID uint16, packetID uint16) []byte {
	packed := make([]byte, 0, 7+len(buffer))
	packed = append(packed, MAGIC_NUMBER)
	packed = append(packed, byte(len(buffer)))
	packed = append(packed, byte(len(buffer)>>8))
	packed = append(packed, byte(connID))
	packed = append(packed, byte(connID>>8))
	packed = append(packed, byte(packetID))
	packed = append(packed, byte(packetID>>8))
	packed = append(packed, buffer...)

	return packed
}

func WritePacket(writer io.Writer, buffer []byte, connID uint16, packetID uint16) (n int, err error) {
	packed := Pack(buffer, connID, packetID)

	n = 0
	for n < len(packed) {
		n_, err := writer.Write(packed[n:])
		if err != nil {
			return n + n_, err
		}
		n += n_
	}
	n -= 7
	return
}

func Unpack(buffer []byte) (connID uint16, packetID uint16, data []byte, err error) {
	if buffer[0] != MAGIC_NUMBER {
		return 0, 0, nil, errors.New("invalid magic number")
	}
	length := int(buffer[1]) | int(buffer[2])<<8
	connID = uint16(buffer[3]) | uint16(buffer[4])<<8
	packetID = uint16(buffer[5]) | uint16(buffer[6])<<8
	data = buffer[7:]
	if length != len(data) {
		return 0, 0, nil, errors.New("invalid packet length")
	}
	return
}

func ReadPacket(reader io.Reader, buffer []byte) (length int, connID uint16, packetID uint16, err error) {
	header := make([]byte, 7)
	n, err := reader.Read(header)
	if err != nil {
		return
	}
	if n < 7 {
		return 0, 0, 0, errors.New("invalid packet")
	}

	if header[0] != MAGIC_NUMBER {
		return 0, 0, 0, errors.New("invalid magic number")
	}

	length = int(header[1]) | int(header[2])<<8
	connID = uint16(header[3]) | uint16(header[4])<<8
	packetID = uint16(header[5]) | uint16(header[6])<<8
	pt := 0
	for pt < length {
		n, err = reader.Read((buffer)[pt:length])
		if err != nil {
			return
		}
		pt += n
	}
	return
}
