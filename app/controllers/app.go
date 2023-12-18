package controllers

import (
	s "strings"

	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/routes"
	"github.com/revel/revel"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	var resPrefix, imgPrefix string

	if s.Index(revel.AppName, "(dev)") >= 0 {
		app.Environment = "dev"
	} else {
		app.Environment = "prod"
		revel.AppLog.Infof("revel.AppName='%s' -> Environment=prod", revel.AppName)
	}

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

	revel.AppLog.Debugf("env = %s", app.Environment)
	if app.Environment != "" && app.Environment != "prod" {
		imgPrefix = app.Environment + "-"
	}

	return c.Render(links, resPrefix, imgPrefix)
}
