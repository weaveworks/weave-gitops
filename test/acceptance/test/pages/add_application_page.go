package pages

import (
	"github.com/sclevine/agouti"
)

type AddAppPageElements struct {
	AddAppHeader         *agouti.Selection
	AppName              *agouti.Selection
	AppNamespace         *agouti.Selection
	AppRepoURL           *agouti.Selection
	ConfigRepoURL        *agouti.Selection
	PathToManifests      *agouti.Selection
	Branch               *agouti.Selection
	AutoMergeCheck       *agouti.Selection
	SubmitButton         *agouti.Selection
	AuthenticationButton *agouti.Selection
}

func GetAddAppPageElements(webDriver *agouti.Page) *AddAppPageElements {
	addApplicationPage := AddAppPageElements{
		AddAppHeader:         webDriver.FindByXPath(`//*[@id="app"]//nav//h2[text()="Add Application"]`),
		AppName:              webDriver.FindByID("name"),
		AppNamespace:         webDriver.FindByID("namespace"),
		AppRepoURL:           webDriver.FindByID("url"),
		ConfigRepoURL:        webDriver.FindByID("configRepo"),
		PathToManifests:      webDriver.FindByID("path"),
		Branch:               webDriver.FindByID("branch"),
		AutoMergeCheck:       webDriver.FindByXPath(`//input[@type="checkbox" and contains(@class, 'MuiSwitch-input')]`),
		SubmitButton:         webDriver.FindByXPath(`//button[@type="submit" and //@span=(text()="Submit")]`),
		AuthenticationButton: webDriver.FindByXPath(`//button[@type="button"]/span[text()="Authenticate with Github"]`)}

	return &addApplicationPage
}
