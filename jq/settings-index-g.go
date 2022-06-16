package main

import (
	"log"
	"time"

	"github.com/gopherjs/jquery"
)

var jq = jquery.NewJQuery

func main() {

	jq("#s_tz").On(jquery.CHANGE, func() {
		jquery.Get(`/util/SetTimeZone?tz=`+jq("#s_tz").Val(), func(data interface{}) {
			jq("#d_saveStatus").Hide()
			if data.(string) == "ok" {
				jq("#d_saveStatus").RemoveClass("notSavedBlip")
				jq("#d_saveStatus").AddClass("savedBlip")
				jq("#d_saveStatus").SetHtml("Saved")
			} else {
				jq("#d_saveStatus").RemoveClass("SavedBlip")
				jq("#d_saveStatus").AddClass("notSavedBlip")
				jq("#d_saveStatus").SetHtml("Error; not saved")
				log.Print(data.(string))
			}
			jquery.Get(`/util/GetWGCredsStatus`, func(dataExpire interface{}) {
				jq("#credsStatus").SetHtml(dataExpire.(string))
			})
			jq("#d_saveStatus").FadeIn(500)
			time.AfterFunc(time.Second*3, func() {
				jq("#d_saveStatus").FadeOut(1000)
			})
		})
	})
}

// vim:foldmethod=marker:
