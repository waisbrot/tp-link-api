package lib

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	http   *http.Client
	req    *http.Request
	conf   *Config
	host   string
	magic  string
	cookie string
}

func (client *Client) generateCookie() (err error) {
	log.Infof("pass=%s", client.conf.Password)
	passMd5 := md5.Sum([]byte(client.conf.Password))
	log.Infof("passMd5=%s", string(passMd5[:]))
	passHex := hex.EncodeToString(passMd5[:])
	log.Infof("passhex=%s", passHex)
	userPass := fmt.Sprintf("%s:%s", client.conf.Username, passHex)
	log.Infof("userPass=%s", userPass)
	userPassB64 := base64.StdEncoding.EncodeToString([]byte(userPass))
	log.Infof("userPassB64=%s", userPassB64)
	authString := "Basic " + userPassB64
	// Neither Query nor Path escape but something stupid in between
	replacer := strings.NewReplacer(
		" ", "%20",
		"+", "%2B",
		"/", "%2F",
		"=", "%3D",
	)
	authStringEscaped := replacer.Replace(authString)
	client.cookie = "Authorization=" + authStringEscaped
	return
}

func (client *Client) blindFetch(path string) (err error) {
	client.setUrl(path)
	body, err := client.doHttp()
	if err != nil {
		return
	}
	if len(body) < 300 {
		log.Errorf("Body looks too short: %s", body)
	}
	return
}

func (client *Client) blindFetches(paths ...string) (err error) {
	for _, path := range paths {
		if err := client.blindFetch(path); err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) parseRedirect(body []byte) (newUrl string, err error) {
	re := regexp.MustCompile(`window.parent.location.href\s*=\s*"([^"]+)"`)
	matches := re.FindSubmatch(body)
	newUrl = string(matches[1])
	log.Info("newUrl=", newUrl)
	return
}

func (client *Client) setUrl(newUrl string) (err error) {
	client.req.URL, err = url.Parse("http://" + client.host + client.magic + newUrl)
	return
}

func (client *Client) doHttp() (body []byte, err error) {
	log.WithFields(log.Fields{"url": client.req.URL, "headers": client.req.Header}).Debug("HTTP request")
	res, err := client.http.Do(client.req)
	if err != nil {
		return
	}
	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	return
}

func NewClient(conf *Config) (client *Client, err error) {
	client = &Client{
		conf: conf,
	}
	client.http = &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     false,
			DisableCompression:    true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// Try connecting, to see if the URL is maybe OK
	res, err := client.http.Get(conf.Host)
	if err != nil {
		return
	}
	if res.StatusCode != 200 {
		err = errors.New("Non-200 response")
		return
	}
	log.Info("Get was OK")

	// Now log in
	err = client.generateCookie()
	if err != nil {
		return
	}
	loginUrl := fmt.Sprintf("%s/userRpm/LoginRpm.htm?Save=Save", conf.Host)
	client.req, err = http.NewRequest("GET", loginUrl, nil)
	if err != nil {
		return
	}
	client.req.Header.Set("User-Agent", "curl/7.58.0")
	client.req.Header.Set("Accept", "*/*")
	client.req.Header.Set("Referrer", conf.Host+"/")
	client.req.Header.Set("Cookie", client.cookie)
	client.req.Header.Del("Accept-Encoding")
	body, err := client.doHttp()
	if err != nil {
		return
	}
	newUrl, err := client.parseRedirect(body)
	if err != nil {
		return
	}
	log.Info("Apparently logged in OK")
	redirUrl, err := url.Parse(newUrl)
	if err != nil {
		return
	}
	client.host = redirUrl.Host
	for magicHunt := redirUrl.Path; magicHunt != "/"; magicHunt = path.Dir(magicHunt) {
		client.magic = magicHunt
	}
	log.Debug("magic=", client.magic)
	log.Debug("Follow redirect to ", redirUrl)
	client.req.URL = redirUrl
	body, err = client.doHttp()
	if err != nil {
		return
	}
	if len(body) < 300 {
		err = errors.New("Body looks too short to be the correct page")
	}
	log.Info("Loaded main page")

	err = client.blindFetches(
		"/localiztion/char_set.js",
		"/frames/top.htm",
		"/userRpm/MenuRpm.htm",
	)
	if err != nil {
		return
	}
	return
}

// Hypertext Transfer Protocol
//     GET /userRpm/LoginRpm.htm?Save=Save HTTP/1.1\r\n
//     Host: 192.168.0.1\r\n
//     User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0\r\n
//     Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8\r\n
// Accept-Language: en-US,en;q=0.5\r\n
// Accept-Encoding: gzip, deflate\r\n
// Connection: keep-alive\r\n
// Referer: http://192.168.0.1/\r\n
// Cookie: Authorization=Basic%20YWRtaW46MjEyMzJmMjk3YTU3YTVhNzQzODk0YTBlNGE4MDFmYzM%3D\r\n
// Upgrade-Insecure-Requests: 1\r\n
