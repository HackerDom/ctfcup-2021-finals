package models

import "trash-factory/pkg/serializeb"

type Container struct {
	ID          string
	Size        uint8
	Items       []Item
	Description string
}

func (container *Container) SerializeNew() ([]byte, error) {
	writer := serializeb.NewWriter()
	container.Serialize(&writer)
	return writer.GetBytes()
}

func (container *Container) Serialize(writer *serializeb.Writer) {
	writer.WriteString(container.ID)
	writer.WriteUint8(container.Size)
	writer.WriteArray(serializeb.ToGenericArray(container.Items),
		func(item interface{}, writer *serializeb.Writer) {
			i := item.(Item)
			i.Serialize(writer)
		})
}

func DeserializeContainerNew(buf []byte) (Container, error) {
	reader := serializeb.NewReader(buf)
	return DeserializeContainer(reader)
}

func DeserializeContainer(reader serializeb.Reader) (Container, error) {
	id, err := reader.ReadString()
	if err != nil {
		return Container{}, err
	}
	size, err := reader.ReadUint8()
	if err != nil {
		return Container{}, err
	}

	itemCount, err := reader.ReadArraySize()
	if err != nil {
		return Container{}, err
	}

	items := make([]Item, itemCount)
	for i := 0; i < itemCount; i++ {
		items[i], err = DeserializeItem(reader)
		if err != nil {
			return Container{}, err
		}

	}

	description, err := reader.ReadString()
	if err != nil {
		return Container{}, err
	}

	return Container{
		ID: id,
		Size: size,
		Items: items,
		Description: description,
	}, nil
}