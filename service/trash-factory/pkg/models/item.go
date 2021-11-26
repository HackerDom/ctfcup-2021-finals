package models

import "trash-factory/pkg/serializeb"

type Item struct {
	Type        uint8
	Weight      uint8
	Description string
}

func (item *Item) Serialize() ([]byte, error) {
	writer := serializeb.NewWriter()
	item.SerializeNext(&writer)
	return writer.GetBytes()
}

func (item *Item) SerializeNext(writer *serializeb.Writer) {
	writer.WriteUint8(item.Type)
	writer.WriteUint8(item.Weight)
	writer.WriteString(item.Description)
}

func DeserializeItem(buf []byte) (Item, error) {
	reader := serializeb.NewReader(buf)
	return DeserializeNextItem(reader)
}

func DeserializeNextItem(reader serializeb.Reader) (Item, error) {
	itemType, err := reader.ReadUint8()
	if err != nil {
		return Item{}, err
	}
	weight, err := reader.ReadUint8()
	if err != nil {
		return Item{}, err
	}

	description, err := reader.ReadString()
	if err != nil {
		return Item{}, err
	}

	return Item{
		Type:        itemType,
		Weight:      weight,
		Description: description,
	}, nil
}
