package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jq = jquery.NewJQuery

func main() {

	jq(".newWin").On(jquery.CLICK, func(e jquery.Event) {
		var linkObj jquery.JQuery

		if e.Target.String() == "[object HTMLImageElement]" {
			linkObj = jq(e.Target).Parent()
		} else {
			linkObj = jq(e.Target)
		}
		e.PreventDefault()
		href := linkObj.Attr("href")
		js.Global.Get("window").Call("open", href, "", "")
	})

}

// vim:foldmethod=marker:
