package pages

import (
	"github.com/sclevine/agouti"
)

type DashboardPageElements struct {
	LogoImage          *agouti.Selection
	ApplicationsHeader *agouti.Selection
	AddAppButton       *agouti.Selection
}

type Apps struct {
	AppLinks *agouti.MultiSelection
}

func GetDashboardPageElements(webDriver *agouti.Page) *DashboardPageElements {
	dashboard := DashboardPageElements{
		LogoImage:          webDriver.FindByXPath(`//*[@id="app"]/div//img`),
		ApplicationsHeader: webDriver.FindByXPath(`//*[@id="app"]//div/h2`),
		AddAppButton:       webDriver.FindByXPath(`//*[@id="app"]//button`)}
	return &dashboard
}

func GetApps(webDriver *agouti.Page) *Apps {
	apps := Apps{
		AppLinks: webDriver.AllByXPath(`//*[@id="app"]//table/tbody/tr`)}
	return &apps
}


