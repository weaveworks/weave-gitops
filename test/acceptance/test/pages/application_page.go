package pages

import "github.com/sclevine/agouti"

type AppDetailsPageElements struct {
	ApplicationsHeader      *agouti.Selection
	NameSubheader           *agouti.Selection
	DeploymentTypeSubheader *agouti.Selection
	URLSubheader            *agouti.Selection
	PathSubheader           *agouti.Selection
	AppName                 *agouti.Selection
}

type NamedElements struct {
	AppNameHeader           *agouti.Selection
	AppName                 *agouti.Selection
	AppType                 *agouti.Selection
	AppURL                  *agouti.Selection
	AppPathToManifests      *agouti.Selection
	HelmSuccessMessage      *agouti.Selection
	KustomizeSuccessMessage *agouti.Selection
}

func GetAppDetailsPageElements(webDriver *agouti.Page) *AppDetailsPageElements {
	appDetailsPage := AppDetailsPageElements{
		ApplicationsHeader:      webDriver.FindByXPath(`//*[@id="app"]//nav/ol/li/a/span/h2[text()="Applications"]`),
		NameSubheader:           webDriver.FindByXPath(`//table/tbody/tr/td[1]/div[text()="Name"]`),
		DeploymentTypeSubheader: webDriver.FindByXPath(`//table/tbody/tr/td[2]/div[text()="Deployment Type"]`),
		URLSubheader:            webDriver.FindByXPath(`//table/tbody/tr/td[3]/div[text()="URL"]`),
		PathSubheader:           webDriver.FindByXPath(`//table/tbody/tr/td[4]/div[text()="Path"]`)}

	return &appDetailsPage
}

func GetAppNameElements(webDriver *agouti.Page, appName string) *NamedElements {
	appNameElements := NamedElements{
		AppNameHeader: webDriver.FindByXPath(`//*[@id="app"]//div/nav/ol/li/h2[text()="` + appName + `"]`),
		AppName:       webDriver.FindByXPath(`//table/tbody/tr/td[1]/div[text()="` + appName + `"]`)}

	return &appNameElements
}

func GetAppTypeElement(webDriver *agouti.Page, appType string) *NamedElements {
	appTypeElement := NamedElements{
		AppType: webDriver.FindByXPath(`//table/tbody/tr/td[2]/div[text()="` + appType + `"]`)}
	return &appTypeElement
}

func GetURLElement(webDriver *agouti.Page, appURL string) *NamedElements {
	appURLElement := NamedElements{
		AppURL: webDriver.FindByXPath(`//table/tbody/tr/td[3]/div[text()="` + appURL + `"]`)}
	return &appURLElement
}

func GetPathElement(webDriver *agouti.Page, pathToManifests string) *NamedElements {
	appPathElement := NamedElements{
		AppPathToManifests: webDriver.FindByXPath(`//table/tbody/tr/td[4]/div[text()="` + pathToManifests + `"]`)}
	return &appPathElement
}

func GetMessageElements(webDriver *agouti.Page, msg string) *NamedElements {
	successMsgElement := NamedElements{
		HelmSuccessMessage:      webDriver.FindByXPath(`//table//tbody/tr/td[text()="` + msg + ` install succeeded"]`),
		KustomizeSuccessMessage: webDriver.FindByXPath(`//table/tbody/tr/td[text()="` + msg + `"]`)}

	return &successMsgElement
}
