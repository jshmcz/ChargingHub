package device

import "errors"

// 帧格式（一期最小实现）：
// 0x68 | cmd(1) | deviceNoLen(1) | deviceNo(ASCII) | dataLen(1) | data | CRC16(2, 低字节在前)
// CRC16 采用 Modbus 多项式 0xA001，初值 0xFFFF，计算范围为帧头到 data 末尾。

type Frame struct {
	Cmd      byte
	DeviceNo string
	Data     []byte
}

func crc16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 == 1 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func buildFrame(cmd byte, deviceNo string, data []byte) []byte {
	body := []byte{0x68, cmd, byte(len(deviceNo))}
	body = append(body, []byte(deviceNo)...)
	body = append(body, byte(len(data)))
	body = append(body, data...)
	c := crc16(body)
	return append(body, byte(c&0xFF), byte(c>>8))
}

func parseFrame(raw []byte) (Frame, error) {
	if len(raw) < 7 || raw[0] != 0x68 {
		return Frame{}, errors.New("invalid frame header")
	}
	cmd := raw[1]
	devLen := int(raw[2])
	if len(raw) < 3+devLen+3 {
		return Frame{}, errors.New("truncated device number")
	}
	deviceNo := string(raw[3 : 3+devLen])
	pos := 3 + devLen
	dataLen := int(raw[pos])
	pos++
	if len(raw) < pos+dataLen+2 {
		return Frame{}, errors.New("truncated data")
	}
	data := append([]byte(nil), raw[pos:pos+dataLen]...)
	pos += dataLen
	got := uint16(raw[pos]) | uint16(raw[pos+1])<<8
	if crc16(raw[:pos]) != got {
		return Frame{}, errors.New("crc mismatch")
	}
	return Frame{Cmd: cmd, DeviceNo: deviceNo, Data: data}, nil
}
