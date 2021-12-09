package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"reflect"
	"regexp"
	"strings"
	"time"
	"trash-factory/pkg/api"
	"trash-factory/pkg/models"
)

const (
	OK            = 101 //# operation has been finished sucessfully
	CORRUPT       = 102 // # service is working, but there is no correct flag (only for "get" ops)
	MUMBLE        = 103 // # service is working incorrect (iex: not responding to the protocol)
	DOWN          = 104 //# service not working (iex: no tcp connection can be initialized)
	CHECKER_ERROR = 110 // # something gone wrong with args or with remote part of checker
)

type Verdict struct {
	Code   int
	Reason string
}

type CheckerError struct {
	Verdict Verdict
	Err     error
}

func (e *CheckerError) Unwrap() error { return e.Err }

func (e *CheckerError) Error() string {
	return e.Verdict.Reason
}

func NewVerdict(code int, reason string) *CheckerError {
	return &CheckerError{
		Verdict: Verdict{
			Code:   code,
			Reason: reason,
		},
		Err: errors.New(reason),
	}
}

type Endpoints struct {
	Addr    string
	WebPort string
	CPPort  string
}

func (e Endpoints) GetCPUrl() string {
	return e.Addr + ":" + e.CPPort
}

func (e Endpoints) GetWebUrl() string {
	return e.Addr + ":" + e.WebPort
}

func main() {
	//
	rand.Seed(time.Now().UnixMilli())
	adrr := flag.String("addr", "", "backend url")
	command := flag.String("command", "", "command : check, put1, put2, get1, get2")
	data := flag.String("data", "", "data string")
	flag.Parse()

	v := Run(adrr, command, data)
	fmt.Println(fmt.Sprintf("VERDICT_CODE:%d", v.Code))
	fmt.Println(fmt.Sprintf("VERDICT_REASON:%s", v.Reason))
}

func Run(adrr *string, command *string, data *string) Verdict {
	cpPort := "9090"
	webPort := "8080"
	host := *adrr
	if strings.Contains(*adrr, ":") {
		parts := strings.Split(*adrr, ":")
		host = parts[0]
		port := parts[1]
		cpPort = "21" + port[2:]
		webPort = "11" + port[2:]
	}

	endpoint := Endpoints{
		CPPort:  cpPort,
		WebPort: webPort,
		Addr:    host,
	}

	switch *command {
	case "check":
		err := Check(&endpoint)
		verdict, failed := EnsureSuccess(err)
		if failed {
			return verdict
		}
		return Verdict{
			Code:   OK,
			Reason: "",
		}
	case "put1":
		tokenKey, token, err := Put_User(&endpoint)
		verdict, failed := EnsureSuccess(err)
		if failed {
			return verdict
		}
		return Verdict{
			Code:   OK,
			Reason: fmt.Sprintf("%s:%s", tokenKey, token),
		}
	case "get1":
		parts := strings.Split(*data, ":")
		err := Get_User(&endpoint, parts[0], parts[1])
		verdict, failed := EnsureSuccess(err)
		if failed {
			return verdict
		}
		return Verdict{
			Code:   OK,
			Reason: "",
		}
	case "put2":
		tokenKey, token, containerId, err := Put_Item(&endpoint, *data)
		verdict, failed := EnsureSuccess(err)
		if failed {
			return verdict
		}
		return Verdict{
			Code:   OK,
			Reason: fmt.Sprintf("%s:%s:%s:%s", tokenKey, token, containerId, *data),
		}
	case "get2":
		parts := strings.Split(*data, ":")
		flag := parts[3]
		err := Get_Item(&endpoint, parts[0], parts[1], parts[2], flag)
		verdict, failed := EnsureSuccess(err)
		if failed {
			return verdict
		}
		return Verdict{
			Code:   OK,
			Reason: "",
		}
	}

	return Verdict{
		Code:   CHECKER_ERROR,
		Reason: "Unknown operation",
	}
}

func EnsureSuccess(err error) (Verdict, bool) {
	if err != nil {
		if oerr, ok := err.(*net.OpError); ok {
			return Verdict{
				Code:   DOWN,
				Reason: oerr.Error(),
			}, true
		}
		if cerr, ok := err.(*CheckerError); ok {
			return cerr.Verdict, true
		}

		return Verdict{
			Code:   MUMBLE,
			Reason: err.Error(),
		}, true
	}
	return Verdict{}, false
}

func Check(endpoints *Endpoints) error {

	tokenKey, token, err := CreateUser(endpoints.GetWebUrl())
	if err != nil {
		log.Error(err)
		return err
	}
	client := api.NewClient(endpoints.GetCPUrl(), tokenKey, token)

	log.Info("Try get user")
	user, err := client.GetUser(tokenKey)
	if err != nil {
		return err
	}

	if user.TokenKey != tokenKey {
		return errors.New("Not this user")
	}

	size := 5
	description := GetContainerDescription()
	log.Info("Try create container")
	containerId, err := client.CreateContainer(size, description)
	if err != nil {
		return err
	}

	expectedContainer := models.Container{
		ID:          containerId,
		Description: description,
		Size:        uint8(size),
		Items:       []models.Item{},
	}

	for i := 0; i < size; i++ {
		expectedItem := ItemsGenerator()
		expectedContainer.Items = append(expectedContainer.Items, expectedItem)
		log.Info("Try put item")
		err = client.PutItem(expectedItem, containerId)
		if err != nil {
			return err
		}
		log.Info("Try get item")
		actualItem, err := client.GetItem(containerId, i)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(actualItem, expectedItem) {
			return errors.New("Actual items not equal expected")
		}
	}
	log.Info("Try get container")
	container, err := client.GetContainerInfo(containerId)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(container, expectedContainer) {
		return errors.New("Actual container not equal expected")
	}
	log.Info("Check OK")
	return nil
}

func Put_User(endpoints *Endpoints) (string, string, error) {
	log.Info("Try Put User")
	return CreateUser(endpoints.GetWebUrl())
}

func Get_User(e *Endpoints, tokenKey string, token string) error {
	log.Info("Try Get User")
	client := api.NewClient(e.GetCPUrl(), tokenKey, token)
	user, err := client.GetUser(tokenKey)

	if err != nil {
		return err
	}
	tokenBytes, err := hex.DecodeString(token)
	if user.TokenKey != tokenKey || string(user.Token) != string(tokenBytes) {
		return error(NewVerdict(CORRUPT, "Flag corrupted"))
	}
	log.Info("Get User OK")
	return nil
}

func Put_Item(e *Endpoints, flag string) (string, string, string, error) {
	log.Info("Try Put Item")
	tokenKey, token, err := CreateUser(e.GetWebUrl())
	client := api.NewClient(e.GetCPUrl(), tokenKey, token)
	containerId, err := client.CreateContainer(3, GetContainerDescription())
	if err != nil {
		return "", "", "", err
	}

	item := ItemsGenerator()
	item.Description = flag
	err = client.PutItem(item, containerId)
	if err != nil {
		return "", "", "", err
	}
	log.Info("Put Item OK")
	return tokenKey, token, containerId, err
}

func Get_Item(e *Endpoints, tokenKey string, token string, containerId, flag string) error {
	log.Info("Try Get Item")
	client := api.NewClient(e.GetCPUrl(), tokenKey, token)
	item, err := client.GetItem(containerId, 0)
	if err != nil {
		return err
	}
	if item.Description != flag {
		return error(NewVerdict(CORRUPT, "Flag corrupted"))
	}
	log.Info("Get Item OK")
	return nil
}

func CreateUser(addr string) (string, string, error) {
	var tokenKey string
	var token string
	c := colly.NewCollector()

	var lastE *colly.HTMLElement
	// Find and visit all links
	c.OnHTML("a[href=\"/token\"]", func(e *colly.HTMLElement) {
		err := e.Request.Visit(e.Attr("href"))
		if err != nil {
			return
		}
	})

	c.OnHTML("div[class=\"token-info\"]", func(e *colly.HTMLElement) {
		lastE = e
		keyPattern := regexp.MustCompile("TOKEN KEY: (.*)")
		tokkenPattern := regexp.MustCompile("TOKEN: (.*)")
		r := keyPattern.FindStringSubmatch(e.Text)
		if r != nil {
			tokenKey = r[1]
			r = tokkenPattern.FindStringSubmatch(e.Text)
			if r != nil {
				token = r[1]
			}
		}
	})

	err := c.Visit(fmt.Sprintf("http://%s", addr))

	if tokenKey == "" || token == "" {
		log.Error(fmt.Sprintf("%s. Key %s Token %s", "Can't parse user", tokenKey, token))
		log.Error(lastE.Text)
		return "", "", error(NewVerdict(MUMBLE, "Can't register user"))
	}

	return tokenKey, token, err
}

func ItemsGenerator() models.Item {
	return models.Item{
		Type:        1,
		Description: "test item",
		Weight:      10,
	}
}

func GetContainerDescription() string {
	return "test container"
}
