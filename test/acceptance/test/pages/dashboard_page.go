package pages

import (
	"github.com/sclevine/agouti"
)

type DashboardWebDriver struct {
	LogoImage      *agouti.Selection
	ApplicationTab *agouti.Selection
}

func Dashboard(webDriver *agouti.Page) DashboardWebDriver {
	dashboard := DashboardWebDriver{
		LogoImage:      webDriver.FindByXPath(`//*[@id="app"]/div//img`),
		ApplicationTab: webDriver.FindByXPath(`//*[@id="app"]//div/a/span/span[1]`)}

	return dashboard
}
