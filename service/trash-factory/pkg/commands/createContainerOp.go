package commands

import "trash-factory/pkg/serializeb"

const DescriptionLimit = 50

type CreateContainerOp struct {
	Size        uint8
	Description string
}

func (op CreateContainerOp) Serialize() ([]byte, error) {
	writer := serializeb.NewWriter()
	writer.WriteUint8(op.Size)
	writer.WriteString(op.Description)
	return writer.GetBytes()
}

func DeserializeCreateContainerOp(buf []byte) (CreateContainerOp, error) {
	reader := serializeb.NewReader(buf)
	size, err := reader.ReadUint8()
	if err != nil {
		return CreateContainerOp{}, err
	}
	description, err := reader.ReadString()
	if err != nil {
		return CreateContainerOp{}, err
	}

	descSize := DescriptionLimit
	if descSize > len(description) {
		descSize = len(description)
	} //TODO: possibly should log if not

	return CreateContainerOp{
		Size:        size,
		Description: description[:descSize],
	}, nil
}
