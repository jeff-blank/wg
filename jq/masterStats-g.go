package main

import (
	"fmt"
	"log"

	"github.com/gopherjs/jquery"
	"honnef.co/go/js/dom"
)

const GRAPH_WIDTH = 0.8

var jq = jquery.NewJQuery

func main() {

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

		graph.SetWidth(fmt.Sprintf("%d", int(GRAPH_WIDTH*float64(ww))))

		iw = graph.Width()
		log.Printf("%#v %#v %#v", ww, iw, GRAPH_WIDTH*float64(ww))

		imgLeft := int(float64(ww)/2) - iw/2
		graph.SetCss("left", imgLeft)
		graph.SetCss("top", int(float64(wh)*0.2))

		marginPct := (1 - GRAPH_WIDTH) / 2
		dismiss.SetCss("left", int((1-marginPct)*float64(ww))+1)
		dismiss.SetCss("top", int(float64(wh)*0.2))
		//log.Printf("%#v %#v", ww, (1-marginPct) * float64(ww))

		dimmer.Show()
		popup.Show()

	})

	jq("#graphDismiss").On(jquery.CLICK, func(e jquery.Event) {
		jq("#graphContainer").Hide()
		jq("#dimmer").Hide()
	})

}

// vim:foldmethod=marker:
