package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

const (
	LEAP_DAY_MONTH = time.February
	LEAP_DAYS      = 1
)

var jq = jquery.NewJQuery

func stateProvinceSelect(country, defState, defCounty string) {
	var homeState string

	clearSelect("#sstate", false)
	if country == "US" {
		jquery.Get("/util/GetHomeState", func(data interface{}) {
			homeState = data.(string)
		})
	} else if country != "Canada" {
		jq("#cstate").SetHtml(`<div>unexpected country "` + country + `"</div>`)
		homeState = "--"
	}

	jquery.When(jquery.Get("/util/StatesProvinces?country="+country, func(data interface{}) {
		sel := jq("#sstate")
		for _, state := range data.([]interface{}) {
			sel.Append(jq(`<option>`).SetText(state.(string)))
			if state == defState || (defState == "" && state == homeState) {
				sel.Children("option").Last().SetAttr("selected", "selected")
			}
		}
	})).Done(func() {
		state := jq("#sstate").Val()
		countySelect(state, defCounty)
	})
}

func countySelect(defState, defCounty string) {

	clearSelect("#scounty", false)

	jquery.Get("/util/GetHomeState", func(data interface{}) {
		if defState == "" {
			defState = data.(string)
		}
		jquery.When(jquery.Get("/util/GetHomeCounty", func(data interface{}) {
			if defCounty == "" {
				defCounty = data.(string)
			}
		})).Done(func() {
			jquery.Get("/util/Counties?state="+defState, func(data interface{}) {
				sel := jq("#scounty")
				for _, county := range data.([]interface{}) {
					countyRec := county.(map[string]interface{})
					sel.Append(jq(`<option>`).SetText(countyRec["Region"]))
					if countyRec["Region"] == defCounty {
						sel.Children("option").Last().SetAttr("selected", "selected")
					}
				}
			})
		})
	})
}

func datesOfMonth(e jquery.Event, date string) {
	var (
		year          int
		yearStr       string
		day_in        string
		selectedMonth string
		monthStr      string
	)

	monthDays := map[time.Month]int{
		time.January:   31,
		time.February:  28,
		time.March:     31,
		time.April:     30,
		time.May:       31,
		time.June:      30,
		time.July:      31,
		time.August:    31,
		time.September: 30,
		time.October:   31,
		time.November:  30,
		time.December:  31,
	}

	ySel := jq("#syear")
	mSel := jq("#smonth")
	dSel := jq("#sday")

	if date != "undefined" {
		yearStr = date[:4]
		selectedMonth = date[5:7]
		day_in = date[8:]
	} else {
		yearStr = ySel.Val()
		selectedMonth = mSel.Val()
	}
	year, _ = strconv.Atoi(yearStr)

	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		monthDays[LEAP_DAY_MONTH] += LEAP_DAYS
	}

	if selectedMonth[:1] == "0" {
		monthStr = selectedMonth[1:]
	} else {
		monthStr = selectedMonth
	}
	month, _ := strconv.Atoi(monthStr)

	if date != "undefined" {
		ySel.Children("option").Each(func(i int, elem interface{}) {
			if jq(elem).Val() == yearStr {
				jq(elem).SetAttr("selected", "selected")
			} else {
				jq(elem).RemoveAttr("selected")
			}
		})
		mSel.Children("option").Each(func(i int, elem interface{}) {
			if jq(elem).Val() == selectedMonth {
				jq(elem).SetAttr("selected", "selected")
			} else {
				jq(elem).RemoveAttr("selected")
			}
		})
	}

	clearSelect("#sday", false)
	dMax := monthDays[time.Month(month)]
	dSel.Children("option").First().Remove()
	for d := 1; d <= dMax; d++ {
		dayStr := fmt.Sprintf("%02d", d)
		dSel.Append(jq(`<option>`).SetText(dayStr))
		if dayStr == day_in {
			dSel.Children("option").Last().SetAttr("selected", "selected")
		}
	}
}

func residenceSelect(res string) {
	var currentResidence string

	sel := jq("#sresidence")
	jquery.Get("/util/GetResidences", func(rlData interface{}) {
		jquery.Get("/util/GetCurrentResidence", func(crData interface{}) {
			currentResidence = crData.(string)
			for _, residence := range rlData.([]interface{}) {
				sel.Append(jq("<option>").SetText(residence.(string)))
				if residence == res || (res == "" && residence == currentResidence) {
					sel.Children("option").Last().SetAttr("selected", "selected")
				}
			}
		})
	})
}

func initializeForm() {
	var (
		state      string
		county     string
		country    string
		residence  string
		date       string
		hitStrings map[string]string
	)

	if jq("#fkey").Val() != "" {
		jquery.When(jquery.Get("/util/GetHitById?id="+jq("#h_hitid").Val(), func(data interface{}) {
			var hit map[string]interface{}
			hit = data.(map[string]interface{})
			state = hit["State"].(string)
			county = hit["CountyCity"].(string)
			country = hit["Country"].(string)
			residence = hit["Residence"].(string)
			date = hit["EntDate"].(string)
			hitStrings = map[string]string{
				"country":   country,
				"state":     state,
				"county":    county,
				"residence": residence,
				"entdate":   date,
			}
		})).Done(func() {
			stateProvinceSelect(hitStrings["country"], hitStrings["state"], hitStrings["county"])
			usCounty(hitStrings["county"])
			datesOfMonth(jquery.Event{}, hitStrings["entdate"])
			residenceSelect(hitStrings["residence"])
			jq("#sresidence").SetAttr("disabled", true)
			jq("#r_delete").Show()
		})
	} else {
		stateProvinceSelect("US", "", "")
		usCounty("")
		residenceSelect("")
		jq("#fkey").Focus()
	}
}

func nonUsCity() {
	jq("#scounty").SetAttr("name", "_scounty")
	jq("#scounty").Hide()
	jq("#fcounty").SetAttr("name", "county")
	jq("#fcounty").Show()
	jq("#lcounty").SetHtml("Hit City")
}

func usCounty(county string) {
	jq("#fcounty").SetAttr("name", "_fcounty")
	jq("#fcounty").Hide()
	jq("#scounty").SetAttr("name", "county")
	jq("#scounty").Show()
	jq("#lcounty").SetHtml("Hit County")
}

func clearSelect(objId string, firstSelected bool) {
	sel := jq(objId)
	for len(sel.Children("option").ToArray()) > 0 {
		sel.Children("option").Last().Remove()
	}
	sel.Append(jq(`<option>`).SetText("--"))
	if firstSelected {
		sel.Children("option").First().SetAttr("selected", "selected")
	}
}

func denomSeries(e jquery.Event) {
	var seriesCode, frb string

	dSel := jq("#sdenom")
	denom0 := dSel.Children(`option[value="0"]`)
	denom1 := dSel.Children(`option[value="1"]`)
	denom2 := dSel.Children(`option[value="2"]`)
	serial := jq(e.Target).Val()
	if len(serial) > 0 {
		seriesCode = serial[:1]
		frb = serial[1:2]
	}
	if len(serial) == 11 && seriesCode >= "A" && seriesCode <= "Z" && frb >= "A" && frb <= "L" {
		jq("#fseries").SetVal("")
		jq("#fseries").SetAttr("disabled", true)
		if denom1.Attr("selected") == "selected" || denom2.Attr("selected") == "selected" {
			denom1.RemoveAttr("selected")
			denom2.RemoveAttr("selected")
			denom0.Select()
		}
		denom1.SetAttr("disabled", true)
		denom2.SetAttr("disabled", true)
	} else {
		jq("#fseries").RemoveAttr("disabled")
		denom1.RemoveAttr("disabled")
		denom2.RemoveAttr("disabled")
	}
}

func billFill(e jquery.Event) {
	jquery.Get("/entries/"+jq(e.Target).Val(), func(data interface{}) {
		var billData map[string]interface{} = data.(map[string]interface{})
		if billData["Id"].(float64) == 0 {
			return
		}
		jq("#fserial").SetVal(billData["Serial"])
		jq("#fseries").SetVal(billData["Series"])
		jq("#sdenom").Children("option").Each(func(i int, elem interface{}) {
			if jq(elem).Val() == strconv.Itoa(int(billData["Denomination"].(float64))) {
				jq(elem).SetAttr("selected", "selected")
			} else if jq(elem).Attr("selected") == "selected" {
				jq(elem).RemoveAttr("selected")
			}
		})
		jq("#sresidence").Children("option").Each(func(i int, elem interface{}) {
			if jq(elem).Val() == billData["Residence"] {
				jq(elem).SetAttr("selected", "selected")
			} else if jq(elem).Attr("selected") == "selected" {
				jq(elem).RemoveAttr("selected")
			}
		})
	})
}

func main() {

	jq(js.Global.Get("document")).Ready(initializeForm)

	jq("#fkey").On(jquery.CHANGE, billFill)

	jq("#sstate").On(jquery.CHANGE, func(e jquery.Event) {
		country := jq("#fcountry").Val()
		if country == "US" {
			state := jq(e.Target).Val()
			countySelect(state, "")
		}
	})

	jq("#fcountry").On(jquery.CHANGE, func(e jquery.Event) {
		country := jq(e.Target).Val()
		if country == "US" {
			// show state and county pickers
			jq("#stateProvince").Show()
			jq("#lstate").SetHtml("Hit State")
			usCounty("")
			stateProvinceSelect("US", "", "")
		} else if country == "Canada" {
			// show province picker and city text-input
			jq("#stateProvince").Show()
			jq("#lstate").SetHtml("Hit Province")
			nonUsCity()
			stateProvinceSelect("Canada", "", "")
		} else {
			// hide state/province row; replace all options with single option "--"
			jq("#stateProvince").Hide()
			clearSelect("#sstate", true)
			nonUsCity()
		}
	})

	jq("#smonth").On(jquery.CHANGE, datesOfMonth)
	jq("#syear").On(jquery.CHANGE, datesOfMonth)

	jq("#fserial").On(jquery.CHANGE, denomSeries)

}

// vim:foldmethod=marker:
