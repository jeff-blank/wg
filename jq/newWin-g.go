package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jq = jquery.NewJQuery

func main() {

	jq(".newWin").On(jquery.CLICK, func(e jquery.Event) {
		e.PreventDefault()
		href := jq(e.Target).Attr("href")
		js.Global.Get("window").Call("open", href, "", "")
	})

}

// vim:foldmethod=marker:
