package controllers

import (
	"github.com/jeff-blank/wg/app/util"
	"github.com/revel/revel"
)

type Util struct {
	*revel.Controller
}

func (c Util) GetStatesProvinces() revel.Result {
	states, _ := util.GetStates(c.Params.Get("country"))
	return c.RenderJSON(states)
}

func (c Util) GetHomeState() revel.Result {
	return c.RenderText(util.GetHomeRegion("state"))
}

func (c Util) GetCounties() revel.Result {
	counties, _ := util.GetCounties(c.Params.Get("state"))
	return c.RenderJSON(counties)
}

func (c Util) GetHomeCounty() revel.Result {
	return c.RenderText(util.GetHomeRegion("county"))
}

func (c Util) GetFirstHits() revel.Result {
	regionType, _ := util.GetFirstHits(c.Params.Get("type"))
	return c.RenderJSON(regionType)
}

func (c Util) GetCurrentResidence() revel.Result {
	r, _ := util.GetCurrentResidence()
	return c.RenderText(r)
}

func (c Util) GetResidences() revel.Result {
	residences, _ := util.GetResidences()
	return c.RenderJSON(residences)
}

func (c Util) GetHitById() revel.Result {
	id := c.Params.Get("id")
	revel.AppLog.Errorf("Util.GetHitById(): id=%s", id)
	h, _ := util.GetHits("and h.id = '" + id + "'")
	if len(h) != 1 {
		return c.RenderText("bill with id '" + id + "' not found or too many results")
	}
	return c.RenderJSON(h[0])
}

// vim:foldmethod=marker:
