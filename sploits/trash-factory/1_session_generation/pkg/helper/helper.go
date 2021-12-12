package helper

import (
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

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

func CreateUser(addr string) (string, string, error) {
	for i := 0; i < 5; i++ {
		key, token, err := CreateUserBase(addr)
		if err == nil {
			return key, token, err
		} else {
			log.Error(err)
		}
	}
	return "", "", errors.New("Can't register user")
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
		return "", "", errors.New("Can't register user")
	}

	return tokenKey, token, err
}
