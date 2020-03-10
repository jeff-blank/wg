package controllers

import (
	"github.com/jeff-blank/wg/app/util"
	"github.com/revel/revel"
)

type Util struct {
	*revel.Controller
}

func (c Util) GetStatesProvinces() revel.Result {
	states := util.GetStates(c.Params.Get("country"))
	return c.RenderJSON(states)
}

func (c Util) GetHomeState() revel.Result {
	return c.RenderText(util.GetHomeRegion("state"))
}

func (c Util) GetCounties() revel.Result {
	counties := util.GetCounties(c.Params.Get("state"))
	return c.RenderJSON(counties)
}

func (c Util) GetHomeCounty() revel.Result {
	return c.RenderText(util.GetHomeRegion("county"))
}

// vim:foldmethod=marker:
