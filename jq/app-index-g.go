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

		if imgFile != "wghbs.svg" {
			newHtml = `<a onclick="window.open(this.href); return false" href="` + imgPath + s.ReplaceAll(s.ReplaceAll(imgFile, "wghbs", "wgushbc"), ".svg", ".png") + `">`
		}
		newHtml += `<img src="` + newImg + `" alt="` + newAlt + `" width="720" height="465">`
		if imgFile != "wghbs.png" {
			newHtml += `</a>`
		}

		main.SetHtml(newHtml)
	})

}

// vim:foldmethod=marker:
