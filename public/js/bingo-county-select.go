package main

import (
	"fmt"
	"log"
	s "strings"

	"github.com/gopherjs/jquery"
)

var jq = jquery.NewJQuery

func initializeForm() {
	jquery.Get("/private/wg/rvd/util/StatesProvinces?country=US", func(data interface{}) {
		sel := jq("#s_state")
		for _, state := range data.([]interface{}) {
			sel.Append(jq(`<option>`).SetText(state.(string)))
		}
	})
}

func clearSelect(objId string) {
	sel := jq(objId)
	for len(sel.Children("option").ToArray()) > 0 {
		sel.Children("option").Last().Remove()
	}
}

func populateCounties(e jquery.Event) {
	state := jq(e.Target).Val()
	added := jq("#s_counties")
	clearSelect("#pickCounties")
	jquery.Get("/private/wg/rvd/util/Counties?id=true&state="+state, func(data interface{}) {
		sel := jq("#pickCounties")
		for _, county := range data.([]interface{}) {
			countyName := county.(map[string]interface{})["Region"]
			countyId := county.(map[string]interface{})["Id"]
			newOpt := jq("<option>")
			newOpt.SetText(countyName)
			newOpt.SetVal(countyId)
			filter := fmt.Sprintf(`option[value="%d"]`, int(countyId.(float64)))
			if len(added.Children(filter).ToArray()) > 0 {
				newOpt.SetAttr("disabled", "disabled")
			}
			sel.Append(newOpt)
		}
	})
}

func addCounties() {
	dest := jq("#s_counties")
	state := jq("#s_state").Val()
	sel := jq("#pickCounties")
	selected := s.Split(sel.Val(), ",")
	log.Printf("%#v", selected)
	for _, idStr := range selected {
		opt := sel.Children(fmt.Sprintf(`option[value="%s"]`, idStr)).First()
		dest.Append(opt.Clone().SetText(opt.Text() + ", " + state))
		opt.SetAttr("disabled", "disabled")
	}
}

func removeCounties() {
	dest := jq("#s_counties")
	sel := jq("#pickCounties")
	selected := s.Split(dest.Val(), ",")
	log.Printf("%#v", selected)
	for _, idStr := range selected {
		opt := dest.Children(fmt.Sprintf(`option[value="%s"]`, idStr)).First()
		opt.Remove()
		opt = sel.Children(fmt.Sprintf(`option[value="%s"]`, idStr))
		if len(opt.ToArray()) > 0 {
			opt.First().RemoveAttr("disabled")
		} else {
			log.Printf("no value '%s' visible", idStr)
		}
	}
}

func selectAllSubmit(e jquery.Event) {
	if jq("#ftitle").Val() == "" {
		// no title
		e.PreventDefault()
		return
	}
	opts := jq("#s_counties").Children("option")
	if len(opts.ToArray()) < 1 {
		// no counties
		e.PreventDefault()
		return
	}
	selected := ""
	opts.Each(func(i int, elem interface{}) {
		selected += jq(elem).Val() + ","
	})
	jq("#h_counties").SetVal(selected)
}

func main() {

	jq("#form").Ready(initializeForm)
	jq("#s_state").On(jquery.CHANGE, populateCounties)
	jq("#b_add").On(jquery.CLICK, addCounties)
	jq("#b_remove").On(jquery.CLICK, removeCounties)
	jq("#form").On(jquery.SUBMIT, selectAllSubmit)

}

// vim:foldmethod=marker:
