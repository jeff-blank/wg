package main

import (
	"fmt"
	"strconv"
	s "strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/dom"
)

const HITS_PATH = "/util/GetFirstHits?type="

var jq = jquery.NewJQuery

func getShowHits(data interface{}) {
	//log.Printf("%#v", data)
	table := `<table id="dataTable">`
	i := len(data.([]interface{}))
	cellNum := 0
	for _, ent := range data.([]interface{}) {
		entry := map[string]interface{}(ent.(map[string]interface{}))
		denom := strconv.Itoa(int(entry["Denom"].(float64)))
		serial := entry["Serial"].(string)
		if len(serial) == 10 {
			serial += "&nbsp;"
		}
		row := `<tr>`
		row += `<td class="c_hit bordered" id="h_` + strconv.Itoa(cellNum) + `">` + strconv.Itoa(i) + `</td>`
		row += `<td class="c_denom aright bordered" id="d_` + strconv.Itoa(cellNum) + `">` + denom + `</td>`
		row += fmt.Sprintf(`<td class="c_bill mono bordered" id="b_`+strconv.Itoa(cellNum)+`">%s&nbsp;/&nbsp;%s</td>`,
			serial,
			entry["Series"].(string),
		)
		row += `<td class="c_date bordered" id="l_` + strconv.Itoa(cellNum) + `"><a href="https://www.wheresgeorge.com/` +
			entry["RptKey"].(string) +
			`" class="newWinHack">` +
			entry["EntDate"].(string) +
			`</a></td>`
		row += `<td class="c_state bordered" id="s_` + strconv.Itoa(cellNum) + `">` + entry["State"].(string) + `</td>`
		row += `<td class="c_county bordered" id="c_` + strconv.Itoa(cellNum) + `">` + entry["CountyCity"].(string) + `</td>`
		table += row
		i--
		cellNum++
	}
	table += "</table>"
	jq("#dataTable").Remove()
	jq("#scroller").RemoveAttr("style")
	jq("#scroller").SetHtml(table)
	jq(".newWinHack").Each(func(i int, elem interface{}) {
		jq(elem).RemoveClass("newWinHack").AddClass("newWin")
	})
}

func winSize() map[string]int {
	dimensions := make(map[string]int)
	dimensions["x"] = dom.GetWindow().InnerWidth()
	dimensions["y"] = dom.GetWindow().InnerHeight()
	return dimensions
}

func throbberShow(dimmer, throbber jquery.JQuery) {
	winXY := winSize()
	imgLeft := int(float64(winXY["x"])/2) - throbber.Width()/2
	throbber.SetCss("left", imgLeft)
	imgTop := int(float64(winXY["y"])/2) - throbber.Height()/2
	throbber.SetCss("top", imgTop)
	dimmer.Show()
	throbber.Show()
}

func setActiveTab(tabName string) {
	jq("#c_" + tabName).RemoveClass("tab-inactive").AddClass("tab-active")
	jq("#" + tabName).RemoveClass("tab-inactive").AddClass("tab-active")
	if tabName == "counties" {
		jq("#c_states").RemoveClass("tab-active").AddClass("tab-inactive")
		jq("#states").RemoveClass("tab-active").AddClass("tab-inactive")
	} else {
		jq("#c_counties").RemoveClass("tab-active").AddClass("tab-inactive")
		jq("#counties").RemoveClass("tab-active").AddClass("tab-inactive")
	}
}

func main() {
	var hitsPath string
	throbber := jq("#throbber")
	dimmer := jq("#dimmer")

	jq(js.Global.Get("document")).Ready(func() {
		throbberShow(dimmer, throbber)

		url := js.Global.Get("location").String()
		typeOffset := s.Index(url, "#")
		hitsPath = HITS_PATH
		if typeOffset < 0 || url[typeOffset+1:] != "counties" {
			hitsPath = HITS_PATH + "state"
			setActiveTab("states")
		} else {
			hitsPath = HITS_PATH + "county"
			setActiveTab("counties")
		}
		jquery.Get(hitsPath, func(data interface{}) {
			getShowHits(data)
			js.Global.Get("tableFix").Call("tf", "c_county")
			throbber.Hide()
			dimmer.Hide()
		})
	})

	jq(".tab-click").On(jquery.CLICK, func(e jquery.Event) {
		throbberShow(dimmer, throbber)
		tabClick := jq(e.Target).Attr("id")
		setActiveTab(tabClick)
		url := js.Global.Get("location").String()
		typeOffset := s.Index(url, "#")
		if tabClick == "counties" {
			hitsPath = HITS_PATH + "county"
			if typeOffset < 0 {
				url += "#counties"
			} else {
				url = url[:typeOffset] + "#counties"
			}
		} else {
			hitsPath = HITS_PATH + "state"
			if typeOffset < 0 {
				url += "#states"
			} else {
				url = url[:typeOffset] + "#states"
			}
		}
		js.Global.Set("location", url)

		jquery.Get(hitsPath, func(data interface{}) {
			getShowHits(data)
			js.Global.Get("tableFix").Call("tf", "c_county")
			throbber.Hide()
			dimmer.Hide()
		})
	})
}

// vim:foldmethod=marker:
