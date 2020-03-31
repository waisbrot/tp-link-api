package lib

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Client(conf *Config) (client string, err error) {
	tpClient := &http.Client{
		Timeout: time.Second * 5,
	}
	res, err := tpClient.Get(conf.Host)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		err = errors.New("Non-200 response")
		return
	}
	log.Info("Get was OK")

	passMd5 := md5.Sum([]byte(conf.Password))
	passHex := hex.EncodeToString(passMd5[:])
	passUpper := strings.ToUpper(passHex)

	loginValues := url.Values{}
	loginValues.Set("username", conf.Username)
	loginValues.Set("password", passUpper)

	res, err = tpClient.PostForm(conf.Host, loginValues)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		err = errors.New(res.Status)
	}
	log.Info("Post was OK")

	urlWithTimeQuery := fmt.Sprintf("%s/data/monitor.client.client.json?operation=load&_=%d", conf.Host, time.Now().Unix()*1000)
	res, err = tpClient.Get(urlWithTimeQuery)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		err = errors.New(res.Status)
	}
	results, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return
	}
	log.Infof("Fetched %v", string(results))

	client = res.Status
	return
}
