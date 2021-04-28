package scrapper

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"github.com/tebeka/selenium/firefox"
)

// Scrapper holds information needed to construct browser
type Scrapper struct {
	seleniumHubAddr string

	caps selenium.Capabilities
}

// options for Browser
type options struct {
	browserName  string
	browserArgs  []string
	maxInstances int
	maxSessions  int
}

// Option is a an option for configuring browser
type Option func(o options) options

// New returns new configuration needed to start browser
func New(seleniumAddr string, opts ...Option) (*Scrapper, error) {
	var err error

	defer errors.Wrap(err, "scrapper.New")

	// default options if they are not specified
	defaultOpts := options{
		browserName:  "chrome",
		browserArgs:  []string{"--headless", "--window-size=375x850"},
		maxInstances: 10,
		maxSessions:  10,
	}

	// override default options if they are specified
	for _, o := range opts {
		defaultOpts = o(defaultOpts)
	}

	caps := selenium.Capabilities{
		"browserName":  defaultOpts.browserName,
		"maxInstances": defaultOpts.maxInstances,
		"maxSessions":  defaultOpts.maxSessions,
	}

	// add browser args
	if defaultOpts.browserName == "firefox" {
		firefoxCaps := firefox.Capabilities{}
		firefoxCaps.Args = append(firefoxCaps.Args, defaultOpts.browserArgs...)

		caps.AddFirefox(firefoxCaps)
	} else if defaultOpts.browserName == "chrome" {
		chromeCaps := chrome.Capabilities{}
		chromeCaps.Args = append(chromeCaps.Args, defaultOpts.browserArgs...)

		caps.AddChrome(chromeCaps)
	} else {
		return nil, errors.Errorf("unsupported browser: %s", defaultOpts.browserName)
	}

	browser := &Scrapper{
		caps:            caps,
		seleniumHubAddr: fmt.Sprintf("http://%s/wd/hub", seleniumAddr),
	}

	return browser, nil
}

// Start starts browser instance
func (b *Scrapper) Start() (selenium.WebDriver, error) {
	// start a new browser instance
	wd, err := selenium.NewRemote(b.caps, b.seleniumHubAddr)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create new remote")
	}

	return wd, nil
}

// OptBrowserName ...
func OptBrowserName(browserName string) Option {
	return func(o options) options {
		o.browserName = browserName

		return o
	}
}

// OptBrowserArgs ...
func OptBrowserArgs(browserArgs []string) Option {
	return func(o options) options {
		o.browserArgs = browserArgs

		return o
	}
}

// OptMaxInstances ...
func OptMaxInstances(maxInstances int) Option {
	return func(o options) options {
		o.maxInstances = maxInstances

		return o
	}
}

// OptMaxSessions ...
func OptMaxSessions(maxSessions int) Option {
	return func(o options) options {
		o.maxSessions = maxSessions

		return o
	}
}
