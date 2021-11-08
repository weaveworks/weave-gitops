package pages

import "github.com/sclevine/agouti"

type AppDetailsPageElements struct {
	LogoImage *agouti.Selection
}

func GetAppDetailsPageElements(webDriver *agouti.Page) *AppDetailsPageElements {
	appDetailsPage := AppDetailsPageElements{
		LogoImage: webDriver.FindByXPath(`//*[@id="app"]/div//img`)}
		
	return &appDetailsPage
}
