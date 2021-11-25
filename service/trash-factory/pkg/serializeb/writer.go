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

func NewWriter() Writer  {
	return Writer{
		buffer: new(bytes.Buffer),
	}
}

func (writer *Writer)WriteUint8(value uint8) {
	writer.buffer.WriteByte(value)
}

func (writer *Writer)WriteUint32(value int) {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(value))
	writer.buffer.Write(bs)
}

func (writer *Writer)WriteString(value string) {
	stringb := []byte(value)
	writer.WriteUint32(len(stringb))
	writer.buffer.Write(stringb)
}

func (writer *Writer)WriteBytes(value []byte) {
	writer.WriteUint32(len(value))
	writer.buffer.Write(value)
}

func (writer *Writer)WriteArray(items []interface{}, serializeItem func(item interface{}, writer *Writer))  {
	writer.WriteUint32(len(items))
	for _, item := range items {
		serializeItem(item, writer)
	}
}

func (writer Writer) GetBytes() ([]byte, error) {
	return writer.GetBytes()
}

func ToGenericArray(arr ...interface{}) []interface{} {
	return arr
}