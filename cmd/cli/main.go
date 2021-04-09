package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jszwec/csvutil"

	"github.com/tebeka/selenium/chrome"
	"github.com/tebeka/selenium/firefox"

	"log"

	"github.com/tebeka/selenium"

	"github.com/spf13/pflag"
)

const facebookLoginPageURL = "https://www.facebook.com"

var (
	seleniumAddr    = pflag.StringP("selenium-addr", "a", "localhost:4444", "address of selenium server")
	seleniumBrowser = pflag.StringP("selenium-browser", "b", "chrome", "browser to be used by selenium")
	csvPath = pflag.StringP("csv-path", "c", "./data/users.csv", "a path to a csv file that contains user data")
)

func main() {
	pflag.Parse()

	seleniumURL := fmt.Sprintf("http://%s/wd/hub", *seleniumAddr) // configure a url to which golang app will send requests

	// configure which browser to use and add headless mode to it
	caps := selenium.Capabilities{"browserName": *seleniumBrowser} 
	if *seleniumBrowser == "firefox" {
		firefoxCaps := firefox.Capabilities{}
		firefoxCaps.Args = append(firefoxCaps.Args, "--headless")

		caps.AddFirefox(firefoxCaps)
	} else if *seleniumBrowser == "chrome" {
		chromeCaps := chrome.Capabilities{}
		chromeCaps.Args = append(chromeCaps.Args, "--headless")

		caps.AddChrome(chromeCaps)
	} else {
		log.Fatalf("%s browser is not supported", *seleniumBrowser)
	}

	// start the browser
	log.Printf("CREATING new selenium driver, browser: %s\n", *seleniumBrowser)
	driver, err := selenium.NewRemote(caps, seleniumURL)
	if err != nil {
		log.Fatalf("failed to create new selenium driver: %v\n", err)
	}
	log.Printf("SUCCESSFULLY created new selenium driver, browser: %s", *seleniumBrowser)

	// close browser(driver) when we are done
	defer func() {
		if err := driver.Close(); err != nil {
			log.Fatalf("failed to close selenium driver: %v\n", err)
		}
	}()

	// maximize browser window
	log.Println("MAXIMIZING browser window")
	err = driver.MaximizeWindow("")
	if err != nil {
		log.Fatalf("failed to maximize browser window: %v\n", err)
	}
	log.Println("SUCCESSFULLY maximized browser window")

	// navigate to facebook login page
	log.Println("NAVIGATING to facebook login page")
	err = driver.Get(facebookLoginPageURL)
	if err != nil {
		log.Fatalf("failed to navigate to facebook login page: %v\n", err)
	}
	log.Println("SUCCESSFULLY navigated to facebook login page")

	// wait for the login page to load then make screenshot
	time.Sleep(time.Second * 5)

	// find accept cookies button
	elem, err := driver.FindElement(selenium.ByXPATH, "/html/body/div[3]/div[2]/div/div/div/div/div[3]/button[2]")
	if err != nil {
		log.Fatalf("failed to find accept cookies button: %v", err)
	}

	// click on accept cookies button
	err = elem.Click()
	if err != nil {
		log.Fatalf("failed to click on accept cookies button")
	}

	// wait for the accept cookies windows to disappear
	time.Sleep(time.Second * 2)

	makeScreenshot(driver, "accept-cookies.png")
}

// makeScreenshot makes screenshot of current window and saves it in specified path
func makeScreenshot(wd selenium.WebDriver, screenshotName string) {
	// make screenshot
	log.Println("MAKING screenshot")
	bs, err := wd.Screenshot()
	if err != nil {
		log.Fatalf("failed to make screenshot: %v\n", err)
	}
	log.Println("SUCCESSFULLY made screenshot")

	// chdir to /go/src/app so we can save screenshots there
	err = os.Chdir("/go/src/app/")
	if err != nil {
		log.Fatalf("failed to chdir: %v\n", err)
	}

	// save screenshot to screenshot folder
	log.Println("SAVING screenshot")
	err = os.WriteFile(fmt.Sprintf("./screenshots/%s", screenshotName), bs, 0777)
	if err != nil {
		log.Fatalf("failed to save screenshot: %v\n", err)
	}
	log.Println("SUCCESSFULLY saved screenshot")
}

// user ...
type user struct {
	Email string `csv:"email,omitempty"`
	Password string `csv:"password,omitempty"`
}

// readUsersFromCSV reads all users from specified csv file,
// and returns their emails + passwords
func readUsersFromCSV(csvPath string) []*user {
	// open csv file for parsing
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("failed to open csv file: %v", err)
	}

	// close csv file when we are done with it
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close csv file: %v", err)
		}
	}()

	// decode csv file
	dec, err := csvutil.NewDecoder(csv.NewReader(f), "")
	if err != nil {
		log.Fatalf("failed to decode csv file: %v", err)
	}

	// read every line of record from csv file and save it to users struct
	users := make([]*user, 0, 1)
	for {
		u := &user{}
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		users = append(users, u)
	}

	return users
}