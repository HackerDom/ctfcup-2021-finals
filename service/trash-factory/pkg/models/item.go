package models

import "trash-factory/pkg/serializeb"

type Item struct {
	Type        uint8
	Weight      uint8
	Description string
}

func (item *Item) SerializeNew() ([]byte, error)  {
	writer := serializeb.NewWriter()
	item.Serialize(&writer)
	return writer.GetBytes()
}

func (item *Item) Serialize(writer *serializeb.Writer) {
	writer.WriteUint8(item.Type)
	writer.WriteUint8(item.Weight)
	writer.WriteString(item.Description)
}

func DeserializeItemNew(buf []byte) (Item, error) {
	reader := serializeb.NewReader(buf)
	return DeserializeItem(reader)
}


func DeserializeItem(reader serializeb.Reader) (Item, error) {
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
		Type: itemType,
		Weight: weight,
		Description: description,
	}, nil
}