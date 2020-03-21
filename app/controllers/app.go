package controllers

import (
	"github.com/jeff-blank/wg/app/routes"
	"github.com/revel/revel"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	var resPrefix string

	residence := c.Params.Get("res")
	if residence == "pdx" || residence == "cmx" {
		resPrefix = residence + "_"
	}
	links := make(map[string]string)
	links = map[string]string{
		"reports":   routes.Reports.Index(),
		"hits":      routes.Hits.Index() + "?year=current",
		"hitAdd":    routes.Hits.New(),
		"hitsBreak": routes.Hits.Breakdown(),
		"bingos":    routes.Bingos.Index(),
		"logout":    routes.Login.Logout(),
		"newrelic":  "https://rpm.newrelic.com/accounts/1720615/applications/",
	}
	if revel.RunMode == "prod" {
		links["newrelic"] += "252898343"
	} else {
		links["newrelic"] += "45030346"
	}

	return c.Render(links, resPrefix)
}
