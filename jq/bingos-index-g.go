package main

import (
	"fmt"
	s "strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jq = jquery.NewJQuery

func main() {
	regionColumn := "c_dtlRegion"
	hitColumn := "c_dtlHit"
	cellContainerId := "#d_bingoDtl"

	pagePct := 0.92
	topElements := [2]string{"#h1", "#p_add"}
	topExtra := 0
	headRowId := "#r_dtlHeadRow"
	headTableId := "#t_dtlHead"
	scrollerId := "#d_dtlScroll"
	lastColId := "#" + hitColumn

	jq(".detLink").On(jquery.CLICK, func(e jquery.Event) {
		e.PreventDefault()
		ic := jq(e.Target).Attr("id")
		sep := s.Index(ic, "_")
		rest := ic[sep+1:]
		sep = s.Index(rest, "_")
		bId := rest[:sep]
		bName := rest[sep+1:]

		dtlHeadTable := jq("<table/>").AddClass("bordered").SetAttr("id", headTableId[1:])
		jq(headTableId).Remove()
		jq(scrollerId).Remove()

		cell1Html := fmt.Sprintf(`%s Detail <span class="small">(<a href="bingos/%s/edit">edit</a>)</span>`, bName, bId)
		dtlHead1Cell := jq("<th/>").SetAttr("colspan", "2").AddClass("bordered").SetHtml(cell1Html)
		dtlHead2Cell1 := jq("<th/>").AddClass("bordered "+regionColumn).SetHtml("<strong>County</strong>").SetAttr("id", regionColumn)
		dtlHead2Cell2 := jq("<th/>").AddClass("bordered "+hitColumn).SetHtml("<strong>Hit?</strong>").SetAttr("id", hitColumn)
		dtlHeadRow2 := jq("<tr/>").SetAttr("id", headRowId[1:]).Append(dtlHead2Cell1, dtlHead2Cell2)

		dtlHeadTable.Append(jq("<tr/>").Append(dtlHead1Cell), dtlHeadRow2)

		jq(cellContainerId).Append(dtlHeadTable)

		dtlScroller := jq("<div>").AddClass("bordered scrollable").SetAttr("id", scrollerId[1:])
		dtlBodyTable := jq("<table/>")

		jquery.Get("bingos/"+bId, func(data interface{}) {
			for _, ent := range data.([]interface{}) {
				var hitsChar string
				region := ent.(map[string]interface{})["County"].(string) + ", " + ent.(map[string]interface{})["State"].(string)
				hits := ent.(map[string]interface{})["Hits"].(bool)
				if hits {
					hitsChar = "&#x2714;"
				} else {
					hitsChar = "&nbsp;"
				}
				regionCell := jq("<td/>").AddClass("bordered " + regionColumn).SetHtml(region)
				hitCell := jq("<td/>").AddClass("bordered center " + hitColumn).SetHtml(hitsChar)
				dtlBodyTable.Append(jq("<tr/>").Append(regionCell, hitCell))
			}
			jquery.When(jq("div" + cellContainerId).Append(dtlScroller.Append(dtlBodyTable))).Done(func() {
				js.Global.Get("tableAdjust").Call("ta", pagePct, topElements, topExtra, headRowId, headTableId, scrollerId, lastColId)
			})
		})

	})

}

// vim:foldmethod=marker:
