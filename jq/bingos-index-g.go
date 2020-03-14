package main

import (
	"fmt"
	s "strings"

	"github.com/gopherjs/jquery"
)

const (
	DETAIL_START = `
      <table class="bordered">
        <tr>
          <th class="bordered" colspan="2">%s Detail%s</th>
        </tr>
        <tr>
          <th class="bordered">County</th>
          <th class="bordered">Hit?</th>
        </tr>
`

	DETAIL_REGION_CELL = `<td class="bordered">%s</td>`
	DETAIL_COUNT_CELL  = `<td class="bordered center">%s</td></tr>`
)

var jq = jquery.NewJQuery

func main() {

	jq(".detLink").On(jquery.CLICK, func(e jquery.Event) {
		e.PreventDefault()
		ic := jq(e.Target).Attr("id")
		sep := s.Index(ic, "_")
		//bType := ic[:sep]
		rest := ic[sep+1:]
		sep = s.Index(rest, "_")
		bId := rest[:sep]
		bName := rest[sep+1:]
		// name_for_link := bType == "custom" ? "&id=" + bId : "&state" + bState

		html := fmt.Sprintf(DETAIL_START, bName, fmt.Sprintf(` <span class="small">(<a href="bingos/%s/edit">edit</a>)</span>`, bId))
		jq("div#showbrk").SetHtml(html + "</table>")

		brkPath := "bingos/" + bId
		jquery.Get(brkPath, func(data interface{}) {
			for _, ent := range data.([]interface{}) {
				var hitsChar string
				region := ent.(map[string]interface{})["County"].(string) + ", " + ent.(map[string]interface{})["State"].(string)
				hits := ent.(map[string]interface{})["Hits"].(bool)
				if hits {
					hitsChar = "&#x2714;"
				} else {
					hitsChar = "&nbsp;"
				}
				html += fmt.Sprintf(`<tr>`+DETAIL_REGION_CELL, region)
				html += fmt.Sprintf(DETAIL_COUNT_CELL+`</tr>`, hitsChar)
			}
			jq("div#showbrk").SetHtml(html + "</table>")
		})

	})

}

// vim:foldmethod=marker:
