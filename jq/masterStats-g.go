package main

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/dom"
)

const (
	GRAPH_WIDTH_DEFAULT = 0.8
	GRAPH_WIDTH_MAX     = 1024.0
)

var jq = jquery.NewJQuery

func main() {

	graphWidth := GRAPH_WIDTH_DEFAULT
	pagePct := 0.94
	topElements := [1]string{"#h1"}
	topExtra := 0
	headRowId := "#headRow"
	headTableId := "#hTable"
	scrollerId := "#scroller"
	lastColId := "#c_yearHits"

	jq(js.Global.Get("document")).Ready(func() {
		js.Global.Get("tableAdjust").Call("ta", pagePct, topElements, topExtra, headRowId, headTableId, scrollerId, lastColId)
	})

	jq(".graphLink").On(jquery.CLICK, func(e jquery.Event) {
		e.PreventDefault()
		imgHref := jq(e.Target).Attr("href")

		graph := jq("#graph")
		dimmer := jq("#dimmer")
		popup := jq("#graphContainer")
		dismiss := jq("#graphDismiss")

		graph.SetHtml(`<img style="width: inherit" src="` + imgHref + `">`)

		ww := dom.GetWindow().InnerWidth()
		wh := dom.GetWindow().InnerHeight()
		iw := graph.Width()

		graph.SetWidth(fmt.Sprintf("%d", int(graphWidth*float64(ww))))

		iw = graph.Width()

		if float64(iw) > GRAPH_WIDTH_MAX {
			graphWidth = GRAPH_WIDTH_MAX / float64(ww)
			graph.SetWidth(fmt.Sprintf("%d", int(graphWidth*float64(ww))))
			iw = graph.Width()
		}

		imgLeft := int(float64(ww)/2) - iw/2
		graph.SetCss("left", imgLeft)
		graph.SetCss("top", int(float64(wh)*0.2))

		marginPct := (1 - graphWidth) / 2
		dismiss.SetCss("left", int((1-marginPct)*float64(ww))+8)
		dismiss.SetCss("top", int(float64(wh)*0.2))

		dimmer.Show()
		popup.Show()

	})

	dom.GetWindow().AddEventListener("keydown", false, func(e dom.Event) {
		ke := e.(*dom.KeyboardEvent)
		if ke.KeyCode == 27 {
			jq("#graphContainer").Hide()
			jq("#dimmer").Hide()
		}
	})

	jq("#dimmer").On(jquery.CLICK, func(e jquery.Event) {
		jq("#graphContainer").Hide()
		jq("#dimmer").Hide()
	})

	jq("#graphDismiss").On(jquery.CLICK, func(e jquery.Event) {
		jq("#graphContainer").Hide()
		jq("#dimmer").Hide()
	})

	jq(".billEnts").On(jquery.CLICK, func(e jquery.Event) {
		e.PreventDefault()
		href := jq(e.Target).Attr("href")
		js.Global.Get("window").Call("open", href, "BillEnts", "width=620,height=200,location=no,menubars=no,toolbars=no,scrollbars=yes")
	})

}

// vim:foldmethod=marker:
