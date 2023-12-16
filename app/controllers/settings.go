package controllers

import (
	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/util"
	"github.com/revel/revel"
)

type Settings struct {
	*revel.Controller
}

func (c Settings) Index() revel.Result {
	var (
		userTZ string
		prefs  app.UserPrefs
	)

	tzList, err := util.GetTZList()
	if err != nil {
		return c.RenderError(err)
	}

	prefs_a, err := util.GetPrefs()
	if err != nil {
		return c.RenderError(err)
	}
	if prefs_a != nil {
		prefs = prefs_a[0]
	}

	userTZ = prefs.TZString

	return c.Render(tzList, userTZ)
}
