package models

import "trash-factory/pkg/serializeb"

type UserStatistic struct {
	TokenKey string
	Total    int
	ByType   map[uint8]uint8
}

type Statistic struct {
	byUsers map[string]*UserStatistic
	Users   []*UserStatistic
}

func NewStatistic() *Statistic {
	return &Statistic{
		Users:   []*UserStatistic{},
		byUsers: map[string]*UserStatistic{},
	}
}

func (stats *Statistic) AddItem(tokenKey string, item Item) {
	byUsers := stats.byUsers
	if _, ok := byUsers[tokenKey]; !ok {
		newUser := &UserStatistic{
			TokenKey: tokenKey,
			ByType:   map[uint8]uint8{},
			Total:    0,
		}
		byUsers[tokenKey] = newUser
		users := stats.Users
		stats.Users = append(users, newUser)
	}
	userStats := byUsers[tokenKey]
	userStats.Total += int(item.Weight)
	userStats.ByType[item.Type] += item.Weight
}

func (stats *Statistic) Serialize() []byte {
	writer := serializeb.NewWriter()
	stats.SerializeNext(&writer)
	return writer.GetBytes()
}

func (stats *Statistic) SerializeNext(writer *serializeb.Writer) {
	writer.WriteArraySize(len(stats.Users))
	for _, userStats := range stats.Users {
		writer.WriteString(userStats.TokenKey)
		writer.WriteUint32(userStats.Total)
		writer.WriteArraySize(len(userStats.ByType))
		for trashType, weight := range userStats.ByType {
			writer.WriteUint8(trashType)
			writer.WriteUint8(weight)
		}
	}
}

func DeserializeStatistic(buf []byte) (Statistic, error) {
	reader := serializeb.NewReader(buf)
	return DeserializeNextStatistic(reader)
}

func DeserializeNextStatistic(reader serializeb.Reader) (Statistic, error) {
	userCount, err := reader.ReadArraySize()
	stats := make([]*UserStatistic, 0)
	if err != nil {
		return Statistic{}, err
	}
	for i := 0; i < userCount; i++ {
		tokenKey, err2 := reader.ReadString()
		if err2 != nil {
			return Statistic{}, err2
		}

		totalWeight, err := reader.ReadUint32()
		if err != nil {
			return Statistic{}, err
		}

		typeCount, err := reader.ReadArraySize()
		if err != nil {
			return Statistic{}, err
		}
		userStats := &UserStatistic{
			Total:    totalWeight,
			TokenKey: tokenKey,
			ByType:   map[uint8]uint8{},
		}
		for i := 0; i < typeCount; i++ {
			trashType, err := reader.ReadUint8()
			if err != nil {
				return Statistic{}, err
			}
			weight, err := reader.ReadUint8()
			if err != nil {
				return Statistic{}, err
			}
			userStats.ByType[trashType] = weight
		}
		stats = append(stats, userStats)
	}
	return Statistic{
		Users: stats,
	}, nil
}
