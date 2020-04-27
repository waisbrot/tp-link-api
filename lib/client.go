package lib

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	selog "github.com/tebeka/selenium/log"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Client struct {
	conf    *Config
	host    string
	magic   string
	cookie  string
	driver  selenium.WebDriver
	service *selenium.Service
}

func (client *Client) Exit() {
	defer client.service.Stop()
	defer client.driver.Quit()
}

func (client *Client) ClickElement(selector string) {
	elem := client.FindElementBySelector(selector)
	moveToElement(elem, 0, 0)
	err := client.driver.Click(selenium.LeftButton)
	if err != nil {
		panic(err)
	}
}

func (client *Client) FindElementBySelector(selector string) (elem selenium.WebElement) {
	elem, err := client.driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		panic(err)
	}
	return
}

func (client *Client) FindElementsBySelector(selector string) (elems []selenium.WebElement) {
	elems, err := client.driver.FindElements(selenium.ByCSSSelector, selector)
	if err != nil {
		panic(err)
	}
	return
}

func getTextFromElement(element selenium.WebElement) (text string) {
	text, err := element.Text()
	if err != nil {
		panic(err)
	}
	return
}

func sendKeysToElement(element selenium.WebElement, keys string) {
	err := element.SendKeys(keys)
	if err != nil {
		panic(err)
	}
}

func moveToElement(element selenium.WebElement, x, y int) {
	err := element.MoveTo(x, y)
	if err != nil {
		panic(err)
	}
}

func NewClient(conf *Config) (client *Client, err error) {
	client = &Client{
		conf: conf,
		host: conf.Host,
	}
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exeBasePath, err := filepath.Abs(filepath.Dir(exe))
	if err != nil {
		panic(err)
	}
	seleniumBasePath := path.Join(exeBasePath, "linux-amd64")
	seleniumPath := path.Join(seleniumBasePath, "selenium-server-standalone-3.141.59.jar")
	chromeDriverPath := path.Join(seleniumBasePath, "chromedriver")
	port := 8080
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),             // Start an X frame buffer for the browser to run in.
		selenium.ChromeDriver(chromeDriverPath), // Specify the path to ChromeDriver in order to use Chrome.
		selenium.Output(os.Stderr),              // Output debug information to STDERR.
	}
	selenium.SetDebug(false)
	log.Debugf("seleniumPath=%s chromeDriverPath=%s", seleniumPath, chromeDriverPath)
	client.service, err = selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		panic(err)
	}

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"no-sandbox",
			// "verbose",
			"headless",
			"disable-dev-shm-usage"},
	})
	seleniumLoglevel := selog.Info
	caps.AddLogging(selog.Capabilities{
		selog.Server:      seleniumLoglevel,
		selog.Browser:     seleniumLoglevel,
		selog.Client:      seleniumLoglevel,
		selog.Driver:      seleniumLoglevel,
		selog.Performance: seleniumLoglevel,
		selog.Profiler:    seleniumLoglevel,
	})
	client.driver, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}

	// Go to a test page
	if err := client.driver.Get(conf.Host); err != nil {
		panic(err)
	}
	title, err := client.driver.Title()
	if err != nil {
		panic(err)
	}
	log.Info("Title=" + title)

	usernameField := client.FindElementBySelector("#userName")
	sendKeysToElement(usernameField, conf.Username)
	passwordField := client.FindElementBySelector("#pcPassword")
	sendKeysToElement(passwordField, conf.Password)
	client.ClickElement("#loginBtn")
	time.Sleep(500)
	url, err := client.driver.CurrentURL()
	log.Info("Got to " + url)

	return
}
