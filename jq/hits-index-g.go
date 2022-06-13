package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jq = jquery.NewJQuery

func main() {

	jq(js.Global.Get("document")).Ready(func() {
		jq(".hitEdit").On(jquery.CLICK, func(e jquery.Event) {
			e.PreventDefault()
			href := jq(e.Target).Attr("href")
			js.Global.Get("window").Call("open", href, "HitEdit", "width=620,height=500,location=no,menubars=no,toolbars=no,scrollbars=yes")
		})
	})

}

// vim:foldmethod=marker:
