package main

import (
	"fmt"
	s "strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jq = jquery.NewJQuery

func numberStates() {
	jq("#t_usBody").Children("tbody").Children("tr").Each(func(i int, row interface{}) {
		jq(row).Children("td").First().SetHtml(fmt.Sprintf("%d.", i+1))
	})

}

func main() {

	pagePct := 0.92
	topElements := [3]string{
		"#topNav",
		"#topHr",
		"#h1",
	}
	topExtra := 0

	jq(js.Global.Get("document")).Ready(func() {
		numberStates()
		headRowId := "#r_usHeadRow"
		headTableId := "#t_usHead"
		scrollerId := "#d_usHits"
		lastColId := "#c_stateHits"
		js.Global.Get("tableAdjust").Call("ta", pagePct, topElements, topExtra, headRowId, headTableId, scrollerId, lastColId)
	})

	jq(".brklnk").On(jquery.CLICK, func(e jquery.Event) {
		var subRegionType string

		headRowId := "#r_detailHeadRow"
		headTableId := "#t_detailHead"
		scrollerId := "#d_dtlScroll"
		lastColId := "#c_dtlHits"

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

		brkHeadTable := jq("<table/>").SetAttr("id", "t_detailHead").AddClass("bordered")
		jq("#t_detailHead").Remove()
		jq("#d_dtlScroll").Remove()
		topHdrTxt := fmt.Sprintf(`<a href="%s">%s</a> Breakdown`, regionHitsFilterPath, tableRegionName)
		topHdrRow := jq("<tr/>").Append(jq("<th/>").SetAttr("colspan", "3").AddClass("bordered").SetHtml(topHdrTxt))
		rankHdr := jq("<th/>").AddClass("bordered c_dtlRank").SetAttr("id", "c_dtlRank").SetHtml("#")
		regionHdr := jq("<th/>").AddClass("bordered c_dtlRegion").SetAttr("id", "c_dtlRegion").SetHtml(subRegionType)
		countHdr := jq("<th/>").AddClass("bordered c_dtlHits").SetAttr("id", "c_dtlHits").SetHtml("Hits")
		brkHeadTable.Append(topHdrRow)
		brkHeadTable.Append(jq("<tr/>").SetAttr("id", headRowId[1:]).Append(rankHdr, regionHdr, countHdr))
		jq("div#showbrk").Append(brkHeadTable)

		dtlScroller := jq("<div/>").SetAttr("id", "d_dtlScroll").AddClass("scrollable bordered")
		dtlTable := jq("<table/>")

		jquery.Get(brkPath, func(data interface{}) {
			rank := 1
			for _, ent := range data.([]interface{}) {
				region := ent.(map[string]interface{})["Region"].(string)
				count := int(ent.(map[string]interface{})["Count"].(float64))
				uri := hitsFilterPath + `&county=` + region
				rankCell := jq("<td/>").AddClass("bordered aright c_dtlRank").SetHtml(fmt.Sprintf(`%d.`, rank))
				rank++
				regionCell := jq("<td/>").AddClass("bordered c_dtlRegion").SetHtml(fmt.Sprintf(`<a href="%s">%s</a>`, uri, region))
				countCell := jq("<td/>").AddClass("bordered aright c_dtlHits").SetHtml(fmt.Sprintf(`%d`, count))
				dtlTable.Append(jq("<tr/>").Append(rankCell, regionCell, countCell))
			}
			jquery.When(jq("div#showbrk").Append(dtlScroller.Append(dtlTable))).Done(func() {
				js.Global.Get("tableAdjust").Call("ta", pagePct, topElements, topExtra, headRowId, headTableId, scrollerId, lastColId)
			})
		})

	})

}

// vim:foldmethod=marker:
