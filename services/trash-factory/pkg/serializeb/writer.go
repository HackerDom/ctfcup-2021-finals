package serializeb

import (
	"bytes"
	"encoding/binary"
)

type BinarySerializable interface {
	Serialize(writer *Writer)
	SerializeNew(writer *Writer)
}

type Writer struct {
	buffer *bytes.Buffer
}

func NewWriter() Writer {
	return Writer{
		buffer: new(bytes.Buffer),
	}
}

func (writer *Writer) WriteUint8(value uint8) *Writer {
	writer.buffer.WriteByte(value)
	return writer
}

func (writer *Writer) WriteUint32(value int) *Writer {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(value))
	writer.buffer.Write(bs)
	return writer
}

func (writer *Writer) WriteString(value string) *Writer {
	stringb := []byte(value)
	writer.WriteUint32(len(stringb))
	writer.buffer.Write(stringb)
	return writer
}

func (writer *Writer) WriteBytes(value []byte) *Writer {
	writer.WriteUint32(len(value))
	writer.buffer.Write(value)
	return writer
}

func (writer *Writer) WriteArraySize(size int) *Writer {
	writer.WriteUint32(size)
	return writer
}

func (writer Writer) GetBytes() []byte {
	return writer.buffer.Bytes()
}
