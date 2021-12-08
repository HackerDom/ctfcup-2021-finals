package serializeb

import (
	"bytes"
	"encoding/binary"
)

type Reader struct {
	buffer *bytes.Buffer
}

func NewReader(buf []byte) Reader{
	return Reader{
		buffer: bytes.NewBuffer(buf),
	}
}

func (reader *Reader) ReadUint8() (uint8, error) {
	readByte, err := reader.buffer.ReadByte()
	if err != nil {
		return 0, err
	}
	return readByte, nil
}

func (reader *Reader) ReadUint32() (int, error) {
	bs := make([]byte, 4)
	reader.buffer.Write(bs)
	_, err := reader.buffer.Read(bs)
	if err != nil {
		return 0, err
	}
	value := binary.LittleEndian.Uint32(bs)
	return int(value), nil
}

func (reader *Reader) ReadString() (string, error) {
	length, err := reader.ReadUint32()
	if err != nil {
		return "", err
	}

	stringb := make([]byte, length)
	_, err = reader.buffer.Read(stringb)
	if err != nil {
		return "", err
	}

	return string(stringb), nil
}

func (reader *Reader) ReadBytes() ([]byte, error) {
	length, err := reader.ReadUint32()
	if err != nil {
		return nil, err
	}

	bytes := make([]byte, length)
	_, err = reader.buffer.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (reader *Reader) ReadArraySize() (int, error) {
	size, err := reader.ReadUint32()
	if err != nil {
		return -1, err
	}

	return size, nil
}