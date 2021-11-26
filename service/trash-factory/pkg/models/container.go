package models

import "trash-factory/pkg/serializeb"

type Container struct {
	ID          string
	Size        uint8
	Items       []Item
	Description string
}

func (container *Container) Serialize() ([]byte, error) {
	writer := serializeb.NewWriter()
	container.SerializeNext(&writer)
	return writer.GetBytes()
}

func (container *Container) SerializeNext(writer *serializeb.Writer) {
	writer.WriteString(container.ID)
	writer.WriteUint8(container.Size)
	writer.WriteArray(serializeb.ToGenericArray(container.Items),
		func(item interface{}, writer *serializeb.Writer) {
			i := item.(Item)
			i.SerializeNext(writer)
		})
}

func DeserializeContainer(buf []byte) (Container, error) {
	reader := serializeb.NewReader(buf)
	return DeserializeContainerNext(reader)
}

func DeserializeContainerNext(reader serializeb.Reader) (Container, error) {
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
		items[i], err = DeserializeNextItem(reader)
		if err != nil {
			return Container{}, err
		}

	}

	description, err := reader.ReadString()
	if err != nil {
		return Container{}, err
	}

	return Container{
		ID:          id,
		Size:        size,
		Items:       items,
		Description: description,
	}, nil
}
