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
	//table := `<table id="dataTable">`
	table := jq("<table/>").SetAttr("id", "dataTable")
	i := len(data.([]interface{}))
	cellNum := 0
	for _, ent := range data.([]interface{}) {
		entry := map[string]interface{}(ent.(map[string]interface{}))
		denom := strconv.Itoa(int(entry["Denom"].(float64)))
		serial := entry["Serial"].(string)
		if len(serial) == 10 {
			serial += "&nbsp;"
		}

		row := jq("<tr/>")
		jq("<td/>").SetAttr("id", "h_"+strconv.Itoa(cellNum)).AddClass("c_hit bordered").SetHtml(strconv.Itoa(i)).AppendTo(row)
		jq("<td/>").SetAttr("id", "d_"+strconv.Itoa(cellNum)).AddClass("c_denom aright bordered").SetHtml(denom).AppendTo(row)
		serialSeries := fmt.Sprintf("%s&nbsp;/&nbsp;%s", serial, entry["Series"].(string))
		jq("<td/>").SetAttr("id", "b_"+strconv.Itoa(cellNum)).AddClass("c_bill mono bordered").SetHtml(serialSeries).AppendTo(row)
		billUrl := "https://www.wheresgeorge.com/" + entry["RptKey"].(string)
		href := jq("<a/>").SetAttr("href", billUrl).AddClass("newWin").SetHtml(entry["EntDate"].(string))
		jq("<td/>").SetAttr("id", "l_"+strconv.Itoa(cellNum)).AddClass("c_date bordered").Append(href).AppendTo(row)
		jq("<td/>").SetAttr("id", "s_"+strconv.Itoa(cellNum)).AddClass("c_state bordered").SetHtml(entry["State"].(string)).AppendTo(row)
		jq("<td/>").SetAttr("id", "c_"+strconv.Itoa(cellNum)).AddClass("c_county bordered").SetHtml(entry["CountyCity"].(string)).AppendTo(row)

		table.Append(row)
		i--
		cellNum++
	}
	//table += "</table>"
	jq("#dataTable").Remove()
	jq("#scroller").RemoveAttr("style")
	jquery.When(jq("#scroller").Append(table)).Done(func() {
		jq(".newWin").On(jquery.CLICK, func(e jquery.Event) {
			linkObj := jq(e.Target)
			e.PreventDefault()
			href := linkObj.Attr("href")
			js.Global.Get("window").Call("open", href, "", "")
		})
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
