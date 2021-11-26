package main

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"trash-factory/pkg/models"
)

type DataBase struct {
	dbPath          string
	userDBPath      string
	containerDBPath string
}

func NewDataBase() *DataBase {
	db := DataBase{dbPath: "db/"}
	db.userDBPath = db.dbPath + "users/"
	db.containerDBPath = db.dbPath + "containers/"

	err := os.MkdirAll(db.userDBPath, 0777)
	if err != nil {
		log.Fatal("Can't create users folder: " + err.Error())
	}
	err = os.MkdirAll(db.containerDBPath, 0777)
	if err != nil {
		log.Fatal("Can't create containers folder: " + err.Error())
	}

	return &db
}

func (db *DataBase) GetUser(tokenKey string) (*models.User, error) {
	userInfo, err := os.ReadFile(db.userDBPath + tokenKey)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user, err := models.DeserializeUser(userInfo)
	if err != nil {
		return nil, err
	}
	fmt.Println(user.ContainersIds)
	return &user, nil
}

func (db *DataBase) SaveUser(user *models.User) error {
	filename := user.TokenKey
	data, err := user.Serialize()
	if err != nil {
		return err
	}
	err = os.WriteFile(db.userDBPath+filename, data, 0666)
	return err
}

func (db *DataBase) SaveContainer(tokenKey string, container *models.Container) error {
	data, err := container.Serialize()
	if err != nil {
		return err
	}

	if err := os.WriteFile(db.containerDBPath+container.ID, data, 0666); err != nil {
		return err
	}

	user, err := db.GetUser(tokenKey)
	if err != nil {
		return err
	}
	user.ContainersIds = append(user.ContainersIds, container.ID)
	return db.SaveUser(user)
}

func (db *DataBase) GetContainer(containerId string) (*models.Container, error) {
	containerInfo, err := os.ReadFile(db.containerDBPath + containerId)
	if err != nil {
		return nil, errors.New("container not found")
	}

	container, err := models.DeserializeContainer(containerInfo)
	if err != nil {
		return nil, err
	}
	fmt.Printf("id %s size %s\n", container.ID, container.Size)
	return &container, nil
}

func (db *DataBase) GetAllUsers() (*[]string, error) {
	dir, err := os.ReadDir(db.userDBPath)
	if err != nil {
		return nil, err
	}
	users := make([]string, len(dir))
	for i, entry := range dir {
		users[i] = entry.Name()
	}
	return &users, nil
}
