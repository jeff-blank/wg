package main

import (
	s "strings"

	"github.com/gopherjs/jquery"
)

var jq = jquery.NewJQuery

func main() {

	jq("img.mapThumb").On(jquery.CLICK, func(e jquery.Event) {
		var newHtml string

		main := jq("#hbsMain")
		clicked := jq(e.Target)
		newImg := clicked.Attr("src")
		newAlt := clicked.Attr("alt")
		lastSlash := s.LastIndex(newImg, "/")
		imgPath := newImg[:lastSlash+1]
		imgFile := newImg[lastSlash+1:]

		if imgFile != "wghbs.png" {
			newHtml = `<a onclick="window.open(this.href); return false" href="` + imgPath + s.ReplaceAll(imgFile, "wghbs", "wgushbc") + `">`
		}
		newHtml += `<img src="` + newImg + `" alt="` + newAlt + `">`
		if imgFile != "wghbs.png" {
			newHtml += `</a>`
		}

		main.SetHtml(newHtml)
	})

}

// vim:foldmethod=marker:
