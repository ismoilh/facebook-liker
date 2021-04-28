package server

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/tebeka/selenium"
)

func (srv *Server) handleRoot(rw http.ResponseWriter, r *http.Request) {
	log.Logger.Info().Str("path", "/").Msgf("RECEIVED request")
	defer log.Logger.Info().Str("path", "/").Msgf("FINISHED request")

	err := srv.templates.ExecuteTemplate(rw, "like.html", nil)
	if err != nil {
		log.Logger.Err(err).Msg("FAILED to execute template like.html")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

var (
	csvPath = pflag.StringP("csv-path", "c", "/go/src/app/data/users.csv", "a path to a csv file that contains user data")
)

func (srv *Server) handleLike(rw http.ResponseWriter, r *http.Request) {
	log.Logger.Info().Str("path", "/like").Msgf("RECEIVED request")
	defer log.Logger.Info().Str("path", "/like").Msgf("FINISHED request")

	subLogger := log.Logger.With().Str("path", "/like").Logger()

	subLogger.Info().Msg("GETTING browser from pool")
	if err := r.ParseForm(); err != nil {
		log.Logger.Err(err).Msg("ParseForm() err: ")
		return
	}

	name := r.FormValue("url")
	likesValue, err := strconv.ParseInt(r.FormValue("likesCount")[0:], 10, 64)
	if err != nil {
		log.Logger.Err(err)
	}

	var (
		PostsURL   string = name
		likesCount int64  = likesValue
		count      int64  = 0
	)

	subLogger.Info().Msg("SUCCESSFULLY got browser from pool")

	users := readUsersFromCSV(*csvPath)
	for _, u := range users {
		if count != likesCount {
			wd, err := srv.scrp.Start()
			if err != nil {
				log.Logger.Err(err).Msg("Failed to start scrapper")
			}
			var otherHandle string
			wd.MaximizeWindow(otherHandle)
			// navigate to facebook login page
			log.Logger.Info().Msg("NAVIGATING to facebook login page")
			err = wd.Get("https://www.facebook.com")
			if err != nil {
				log.Logger.Err(err).Msg("failed to navigate to facebook login page: \n")
				continue
			}
			log.Logger.Info().Msg("SUCCESSFULLY navigated to facebook login page")

			// wait for the login page to load
			time.Sleep(time.Second * 5)

			// find accept cookies button
			elem, err := wd.FindElement(selenium.ByXPATH, `/html/body/div[3]/div[2]/div/div/div/div/div[3]/button[2]`)
			if err != nil {
				log.Logger.Err(err).Msg("failed to find accept cookies button: \n")
				continue
			}
			err = elem.Click()
			if err != nil {
				log.Logger.Err(err).Msg("failed to click on accept cookies button\n")
				continue
			}

			// wait for the accept cookies windows to disappear
			time.Sleep(time.Second * 2)

			// find email input box
			elem, err = wd.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div/div/div[2]/div/div[1]/form/div[1]/div[1]/input`)
			if err != nil {
				log.Logger.Err(err).Msg("failed to find email input box: \n")
				continue
			}

			// fill email input box
			err = elem.SendKeys(u.Email)
			if err != nil {
				log.Logger.Err(err).Msg("failed to fill email input box: \n")
				continue
			}

			// find password input
			elem, err = wd.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div/div/div[2]/div/div[1]/form/div[1]/div[2]/div/input`)
			if err != nil {
				log.Logger.Err(err).Msg("failed to find password input box: \n")
				continue
			}

			// fill password input box
			err = elem.SendKeys(u.Password)
			if err != nil {
				log.Logger.Err(err).Msg("failed to fill password input box: \n")
				continue
			}

			// click log in button
			elem, err = wd.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div/div/div[2]/div/div[1]/form/div[2]/button`)
			if err != nil {
				log.Logger.Err(err).Msg("failed to click log in button: \n")
				continue
			}

			// click on login button
			err = elem.Click()
			if err != nil {
				log.Logger.Err(err).Msg("failed to click on login button\n")
				continue
			}

			exists, _ := wd.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div[2]/div[1]/div[2]/div/div/div/div/div/div[2]/div[2]/div[1]/button`)

			if exists != nil {
				// click update Terms button
				elem, err = wd.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div[2]/div[1]/div[2]/div/div/div/div/div/div[2]/div[2]/div[1]/button`)
				if err != nil {
					log.Logger.Err(err).Msg("failed to find update Terms button: \n")
					continue
				}

				// click on login button
				err = elem.Click()
				if err != nil {
					log.Logger.Err(err).Msg("failed to click on update Terms button\n")
					continue
				}
			}
			exists1, _ := wd.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div[2]/div[1]/div[2]/div/div/div/div/div[2]/div[4]/button`)

			if exists1 != nil {
				// click close terms button
				elem, err = wd.FindElement(selenium.ByXPATH, `/html/body/div[1]/div[2]/div[1]/div/div[2]/div[1]/div[2]/div/div/div/div/div[2]/div[4]/button`)
				if err != nil {
					log.Logger.Err(err).Msg("failed to find close terms button: \n")
					continue
				}

				// click on close terms button
				err = elem.Click()
				if err != nil {
					log.Logger.Err(err).Msg("failed to click on close terms button\n")
					continue
				}

			}

			log.Logger.Info().Msg("NAVIGATING to posts url page")
			err = wd.Get(PostsURL)
			if err != nil {
				log.Logger.Err(err).Msg("failed to navigate to posts url page: \n")
				continue
			}
			log.Logger.Info().Msg("SUCCESSFULLY navigated to posts url page")

			time.Sleep(5 * time.Second)
			// makeScreenshot(wd, fmt.Sprintf("login-%s.png", u.Email))
			// click like button
			elem, err = wd.FindElement(selenium.ByXPATH, `/html/body/div[1]/div/div[1]/div/div[3]/div/div/div[1]/div[1]/div/div[2]/div/div/div/div[1]/div[2]/div/div[2]/div/div[1]`)

			if err != nil {
				log.Logger.Err(err).Msg("failed to find like button\n")
				continue
			}
			// click on like button
			err = elem.Click()
			if err != nil {
				log.Logger.Err(err).Msg("failed to click on like button\n")
				continue
			}
			count++
			time.Sleep(5 * time.Second)
			// makeScreenshot(wd, fmt.Sprintf("login-%s.png", u.Email))

			// close browser when we are done
			if err := wd.Close(); err != nil {
				log.Logger.Err(err).Msg("failed to close selenium wd\n")
				continue
			}
			// close browser when we are done
			if err := wd.Close(); err != nil {
				log.Logger.Err(err).Msg("failed to close selenium driver: \n")
				continue
			}
			wd.Quit()
		} else {
			break
		}
	}

	http.Redirect(rw, r, "/", http.StatusOK)
}

func (srv *Server) handleCSV(rw http.ResponseWriter, r *http.Request) {
	log.Logger.Info().Str("path", "/csv").Msgf("RECEIVED request")
	defer log.Logger.Info().Str("path", "/csv").Msgf("FINISHED request")
	err := srv.templates.ExecuteTemplate(rw, "csv.html", nil)
	if err != nil {
		log.Logger.Err(err).Msg("FAILED to execute template csv.html")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// makeScreenshot makes screenshot of current window and saves it in specified path
func makeScreenshot(wd selenium.WebDriver, screenshotName string) {
	// make screenshot
	log.Logger.Info().Msg("MAKING screenshot")
	bs, err := wd.Screenshot()
	if err != nil {
		log.Logger.Err(err).Msg("failed to make screenshot: \n")
	}
	log.Logger.Info().Msg("SUCCESSFULLY made screenshot")

	// save screenshot to screenshot folder
	log.Logger.Info().Msg("SAVING screenshot")
	err = os.WriteFile(fmt.Sprintf("/go/src/app/screenshots/%s", screenshotName), bs, 0777)
	if err != nil {
		log.Logger.Err(err).Msg("failed to save screenshot: \n")
	}
	log.Logger.Info().Msg("SUCCESSFULLY saved screenshot")
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
		log.Logger.Err(err).Msg("failed to open csv file: \n")
	}

	// close csv file when we are done with it
	defer func() {
		if err := f.Close(); err != nil {
			log.Logger.Err(err).Msg("failed to close csv file: \n")
		}
	}()

	// decode csv file
	csvr := csv.NewReader(f)
	rr, err := csvr.ReadAll()
	if err != nil {
		log.Logger.Err(err).Msg("failed to read all csv file: \n")
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
