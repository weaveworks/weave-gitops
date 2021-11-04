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
		AddAppHeader:         webDriver.FindByXPath(`//*[@id="app"]//nav//h2`),
		AppName:              webDriver.FindByID("name"),
		AppNamespace:         webDriver.FindByID("namespace"),
		AppRepoURL:           webDriver.FindByID("url"),
		ConfigRepoURL:        webDriver.FindByID("configUrl"),
		PathToManifests:      webDriver.FindByID("path"),
		Branch:               webDriver.FindByID("branch"),
		AutoMergeCheck:       webDriver.FindByXPath(`//input[@type="checkbox"]`),
		SubmitButton:         webDriver.FindByXPath(`//button[@type="submit"]`),
		AuthenticationButton: webDriver.FindByXPath(`//button[@type="submit"]`)}

	return &addApplicationPage
}
