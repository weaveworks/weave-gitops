package pages

import (
	"github.com/sclevine/agouti"
)

type DashboardPageElements struct {
	LogoImage          *agouti.Selection
	ApplicationsHeader *agouti.Selection
	AddAppButton       *agouti.Selection
}

type AppListElements struct {
	AppList *agouti.Selection
}

func GetDashboardPageElements(webDriver *agouti.Page) *DashboardPageElements {
	dashboard := DashboardPageElements{
		LogoImage:          webDriver.FindByXPath(`//*[@id="app"]/div//img`),
		ApplicationsHeader: webDriver.FindByXPath(`//*[@id="app"]//div/h2`),
		AddAppButton:       webDriver.FindByXPath(`//*[@id="app"]//button`)}

	return &dashboard
}

func GetAppListElements(webDriver *agouti.Page, appName string) *AppListElements {
	appList := AppListElements{
		AppList: webDriver.FindByXPath(`//*[@id="app"]//tbody/tr/td/span/a[contains(@href,'` + appName + `')]`)}
	return &appList
}
