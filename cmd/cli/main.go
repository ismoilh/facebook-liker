package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

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
	csvPath         = pflag.StringP("csv-path", "c", "./data/users.csv", "a path to a csv file that contains user data")
)

func main() {
	pflag.Parse()

	seleniumURL := fmt.Sprintf("http://%s/wd/hub", *seleniumAddr) // configure a url to which golang app will send requests

	// configure which browser to use and add headless mode to it
	caps := selenium.Capabilities{"browserName": *seleniumBrowser, "maxInstances": 10, "maxSessions": 10}
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

	// chdir to /go/src/app so we can save screenshots there
	err := os.Chdir("/go/src/app/")
	if err != nil {
		log.Fatalf("failed to chdir: %v\n", err)
	}

	// read users email + password from csv file
	users := readUsersFromCSV(*csvPath)
	for _, u := range users {
		// start the browser
		log.Printf("CREATING new selenium driver, browser: %s\n", *seleniumBrowser)
		driver, err := selenium.NewRemote(caps, seleniumURL)
		if err != nil {
			log.Fatalf("failed to create new selenium driver: %v\n", err)
		}
		log.Printf("SUCCESSFULLY created new selenium driver, browser: %s", *seleniumBrowser)

		// close browser(driver) before log.Fatal
		defer func() {
			driver.Close()
			driver.Quit()
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
			log.Fatalf("failed to find accept cookies button: %v\n", err)
		}

		// click on accept cookies button
		err = elem.Click()
		if err != nil {
			log.Fatalf("failed to click on accept cookies button\n")
		}

		// wait for the accept cookies windows to disappear
		time.Sleep(time.Second * 2)

		// find email input box
		elem, err = driver.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div/div/div[2]/div/div[1]/form/div[1]/div[1]/input`)
		if err != nil {
			log.Fatalf("failed to find email input box: %v\n", err)
		}

		// fill email input box
		err = elem.SendKeys(u.Email)
		if err != nil {
			log.Fatalf("failed to fill email input box: %v\n", err)
		}

		// find password input
		elem, err = driver.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div/div/div[2]/div/div[1]/form/div[1]/div[2]/div/input`)
		if err != nil {
			log.Fatalf("failed to find password input box: %v\n", err)
		}

		// fill password input box
		err = elem.SendKeys(u.Password)
		if err != nil {
			log.Fatalf("failed to fill password input box: %v\n", err)
		}

		// click log in button
		elem, err = driver.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div/div/div[2]/div/div[1]/form/div[2]/button`)
		if err != nil {
			log.Fatalf("failed to click log in button: %v\n", err)
		}

		// click on login button
		err = elem.Click()
		if err != nil {
			log.Fatalf("failed to click on login button\n")
		}

		time.Sleep(5 * time.Second)

		makeScreenshot(driver, fmt.Sprintf("login-%s.png", u.Email))

		// close browser when we are done
		if err := driver.Close(); err != nil {
			log.Fatalf("failed to close selenium driver: %v\n", err)
		}
		driver.Quit()
	}
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
	Email    string `csv:"email,omitempty"`
	Password string `csv:"password,omitempty"`
}

// readUsersFromCSV reads all users from specified csv file,
// and returns their emails + passwords
func readUsersFromCSV(csvPath string) []*user {
	// open csv file for parsing
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("failed to open csv file: %v\n", err)
	}

	// close csv file when we are done with it
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close csv file: %v\n", err)
		}
	}()

	// decode csv file
	csvr := csv.NewReader(f)
	rr, err := csvr.ReadAll()
	if err != nil {
		log.Fatalf("failed to read all csv file: %v\n", err)
	}

	// skip header of csv file
	rr = rr[1:]

	// read every line of record from csv file and save it to users struct
	users := make([]*user, 0, 1)
	for _, r := range rr {
		u := &user{
			Email:    r[0],
			Password: r[1],
		}

		users = append(users, u)
	}

	return users
}
