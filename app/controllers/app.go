package controllers

import (
	"github.com/jeff-blank/wg/app/routes"
	"github.com/revel/revel"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	qs := ""
	filterCountry := c.Params.Get("country")
	filterState := c.Params.Get("state")
	filterCounty := c.Params.Get("county")
	filterYear := c.Params.Get("year")
	filterSort := c.Params.Get("sort")
	if filterYear != "" {
		qs += "&year=" + filterYear
	}
	if filterCountry != "" {
		qs += "&country=" + filterCountry
	}
	if filterState != "" {
		qs += "&state=" + filterState
	}
	if filterCounty != "" {
		qs += "&county=" + filterCounty
	}
	if filterSort != "" {
		qs += "&sort=" + filterSort
	}
	if qs != "" {
		qs = "?" + qs[1:]
	}
	return c.Redirect(routes.Hits.Index() + qs)
}
