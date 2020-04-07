package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

type Client struct {
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

func (client *Client) buildUrl(urlPath string) (url string) {
	return path.Join(client.host, client.magic, urlPath)
}

// Simple wrapper around Curl with the few args we care about
func (client *Client) Curl(pieces map[string]string) (reply string) {
	args := []string{"--silent"}
	if referrer, ok := pieces["referrer"]; ok {
		args = append(args, "-H")
		args = append(args, "Referrer: "+referrer)
	}
	if cookie, ok := pieces["cookie"]; ok {
		args = append(args, "-H")
		args = append(args, "Cookie: "+cookie)
	}
	if url, ok := pieces["url"]; ok {
		args = append(args, url)
	}
	cmd := exec.Command("curl", args...)
	log.Debugf("%+v", cmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Errorf("Error while trying to run Curl")
		panic(err)
	}
	return out.String()
}

// Interpret the arg as a relative path and insert magic + auth
func (client *Client) FetchPath(path string) (reply string, err error) {
	url := client.buildUrl(path)
	reply, err = client.fetchUrl(url)
	return
}

// Interpret the arg as a complete URL and just insert auth headers
func (client *Client) fetchUrl(url string) (reply string, err error) {
	reply = client.Curl(map[string]string{
		"url":      url,
		"referrer": client.host,
		"cookie":   client.cookie,
	})
	return
}

func (client *Client) parseRedirect(body string) (newUrl string, err error) {
	re := regexp.MustCompile(`window.parent.location.href\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(body)
	newUrl = string(matches[1])
	log.Info("newUrl=", newUrl)
	return
}

func NewClient(conf *Config) (client *Client, err error) {
	client = &Client{
		conf: conf,
		host: conf.Host,
	}

	// Try connecting, to see if the URL is maybe OK
	res := client.Curl(map[string]string{"url": client.host})
	log.Info("Get was OK")

	// Now log in
	err = client.generateCookie()
	if err != nil {
		return
	}
	loginUrl := fmt.Sprintf("%s/userRpm/LoginRpm.htm?Save=Save", client.host)
	res, err = client.fetchUrl(loginUrl)
	if err != nil {
		return
	}
	newUrl, err := client.parseRedirect(res)
	if err != nil {
		return
	}
	log.Info("Apparently logged in OK")

	// Parse out the magic string that goes in front of all paths now
	redirUrl, err := url.Parse(newUrl)
	if err != nil {
		return
	}
	client.host = "http://" + redirUrl.Host
	for magicHunt := redirUrl.Path; magicHunt != "/"; magicHunt = path.Dir(magicHunt) {
		client.magic = magicHunt
	}
	log.Debug("magic=", client.magic)
	log.Debug("Follow redirect to ", redirUrl)
	res, err = client.fetchUrl(redirUrl.String())
	if err != nil {
		return
	}
	if len(res) < 300 {
		err = errors.New("Body looks too short to be the correct page")
	}
	log.Info("Loaded main page")

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
