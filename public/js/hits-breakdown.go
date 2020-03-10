package main

import (
	"fmt"
	s "strings"

	"github.com/gopherjs/jquery"
)

const (
	BREAKDOWN_START = `
      <table class="bordered">
        <tr>
          <th class="bordered" colspan="2"><a href="%s">%s</a> Breakdown</th>
        </tr>
        <tr>
          <th class="bordered">%s</th>
          <th class="bordered">Hits</th>
        </tr>
`

	BREAKDOWN_REGION_CELL = `<td class="bordered"><a href="%s">%s</a></td>`
	BREAKDOWN_COUNT_CELL  = `<td class="bordered aright">%d</td></tr>`
)

var jq = jquery.NewJQuery

func main() {

	jq(".brklnk").On(jquery.CLICK, func(e jquery.Event) {
		var subRegionType string

		e.PreventDefault()
		cs := jq(e.Target).Attr("id")
		sep := s.Index(cs, "_")
		country := cs[0:sep]
		state := cs[sep+1:]
		brkPath := "showbrk?country=" + country + "&region=" + state

		regionHitsFilterPath := `hits?country=` + country
		tableRegionName := state
		hitsFilterPath := `hits?country=` + country
		if country == "US" {
			regionHitsFilterPath += `&state=` + state
			hitsFilterPath += `&state=` + state
			subRegionType = "County"
		} else if country == "Canada" {
			regionHitsFilterPath += `&state=` + state
			hitsFilterPath += `&state=` + state
			subRegionType = "City"
		} else {
			tableRegionName = country
			subRegionType = "City"
		}
		cell := fmt.Sprintf(BREAKDOWN_START, regionHitsFilterPath, tableRegionName, subRegionType)

		jquery.Get(brkPath, func(data interface{}) {
			for _, ent := range data.([]interface{}) {
				region := ent.(map[string]interface{})["Region"].(string)
				count := int(ent.(map[string]interface{})["Count"].(float64))
				uri := hitsFilterPath + `&county=` + region
				cell += fmt.Sprintf(`<tr>`+BREAKDOWN_REGION_CELL, uri, region)
				cell += fmt.Sprintf(BREAKDOWN_COUNT_CELL+`</tr>`, count)
			}
			jq("div#showbrk").SetHtml(cell + "</table>")
		})

	})

}

// vim:foldmethod=marker:
