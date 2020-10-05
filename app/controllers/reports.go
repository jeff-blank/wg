package controllers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/routes"
	"github.com/jeff-blank/wg/app/util"
	"github.com/revel/revel"
)

const (
	// SQL queries {{{
	Q_HITS_CALENDAR = `
		select
			substr(entdate::text, 6) as date,
			count(1)
		from
			hits
		group by
			date
	`
	// }}}

	DATE_LIST_YEAR   = 2000
	DATE_LIST_LAYOUT = `2006-01-02`
)

type Reports struct {
	*revel.Controller
}

func (c Reports) Index() revel.Result {
	links := make(map[string]string)
	links["master"] = routes.Reports.MasterStats()
	links["calendar"] = routes.Reports.HitsCalendar()
	links["first"] = routes.Reports.FirstHits()
	links["last50"] = routes.Reports.Last50Counts()
	return c.Render(links)
}

func (c Reports) HitsCalendar() revel.Result {
	// date list is the outer array to ease columnizing dates in template
	var (
		calendar     [31]app.DayOfMonth
		totalsRow    app.DayOfMonth
		missingRow   app.DayOfMonth
		monthTotal   int
		missing      int
		missingTotal int
	)

	dates := make(map[string]int)

	rows, err := app.DB.Query(Q_HITS_CALENDAR)
	if err != nil {
		revel.AppLog.Errorf("query hits calendar: %#v", err)
		return c.RenderError(err)
	}

	defer rows.Close()
	for rows.Next() {
		var date string
		var count int

		err = rows.Scan(&date, &count)
		if err != nil {
			revel.AppLog.Errorf("read hits calendar: %#v", err)
			return c.RenderError(err)
		}

		dates[date] = count
	}

	checkDate := time.Date(DATE_LIST_YEAR, time.January, 1, 0, 0, 0, 0, time.UTC)
	for checkDate.Year() == DATE_LIST_YEAR {
		month0 := checkDate.Month()
		monthDay := checkDate.Day()
		dateStr := checkDate.Format(DATE_LIST_LAYOUT)[5:]
		calendar[monthDay-1].Label = strconv.Itoa(monthDay)
		if count, found := dates[dateStr]; found {
			calendar[monthDay-1].Months[month0-1] = count
			monthTotal += count
		} else {
			missing++
			missingTotal++
		}
		checkDate = checkDate.AddDate(0, 0, 1)
		month1 := checkDate.Month()
		if month0 != month1 {
			if monthDay < 31 {
				for i := monthDay + 1; i <= 31; i++ {
					calendar[i-1].Months[month0-1] = -1
				}
			}
			totalsRow.Months[month0-1] = monthTotal
			missingRow.Months[month0-1] = missing
			monthTotal = 0
			missing = 0
		}
	}

	for d := 0; d < len(calendar); d++ {
		monthDayTotal := 0
		for m := 0; m < len(calendar[d].Months); m++ {
			if calendar[d].Months[m] >= 0 {
				monthDayTotal += calendar[d].Months[m]
			}
		}
		calendar[d].Total = strconv.Itoa(monthDayTotal)
	}

	missingRow.Total = strconv.Itoa(missingTotal)
	return c.Render(calendar, totalsRow, missingRow)
}

func (c Reports) MasterStats() revel.Result {
	var tableIn []map[string]interface{}

	tableIn = util.StatsData("table").([]map[string]interface{})

	for m, monthData := range tableIn {
		month := monthData["month"]
		ents := monthData["monthBills"]
		monthData["monthBills"] = map[string]interface{}{"month": month, "entries": ents}
		tableIn[m] = monthData
	}

	// reverse table for display
	tableOut := make([]map[string]interface{}, 0)
	for m := len(tableIn) - 1; m >= 0; m-- {
		tableOut = append(tableOut, tableIn[m])
	}

	entsLink := routes.Entries.Edit()
	graphLinks := make(map[string]string)
	graphLinks["score"] = routes.Charts.Grapher("score")
	graphLinks["hits"] = routes.Charts.Grapher("hits")
	graphLinks["bills"] = routes.Charts.Grapher("bills")
	return c.Render(tableOut, entsLink, graphLinks)
}

func (c Reports) FirstHits() revel.Result {
	return c.Render()
}

func (c Reports) Last50Counts() revel.Result {
	denomData := make([][2]int, 0)
	seriesData := make([][2]interface{}, 0)
	stateData := make([][2]interface{}, 0)
	countyData := make([][3]interface{}, 0)

	query := fmt.Sprintf(app.Q_LAST_50_DENOM_SERIES, "denomination", "denomination", "denomination")
	rows, err := app.DB.Query(query)
	if err != nil {
		revel.AppLog.Errorf("query last 50 hits' denominations: %#v", err)
		return c.RenderError(err)
	}
	for rows.Next() {
		var (
			denom int
			count int
		)
		err := rows.Scan(&denom, &count)
		if err != nil {
			revel.AppLog.Errorf("read last 50 hits' denominations: %#v", err)
			return c.RenderError(err)
		}
		denomData = append(denomData, [2]int{denom, count})
	}

	query = fmt.Sprintf(app.Q_LAST_50_DENOM_SERIES, "series", "series", "series") + " desc"
	rows, err = app.DB.Query(query)
	if err != nil {
		revel.AppLog.Errorf("query last 50 hits' series: %#v", err)
		return c.RenderError(err)
	}
	for rows.Next() {
		var (
			series string
			count  int
		)
		err := rows.Scan(&series, &count)
		if err != nil {
			revel.AppLog.Errorf("read last 50 hits' series: %#v", err)
			return c.RenderError(err)
		}
		seriesData = append(seriesData, [2]interface{}{series, count})
	}

	rows, err = app.DB.Query(app.Q_LAST_50_STATES)
	if err != nil {
		revel.AppLog.Errorf("query last 50 hits' states: %#v", err)
		return c.RenderError(err)
	}
	for rows.Next() {
		var (
			state string
			count int
		)
		err := rows.Scan(&state, &count)
		if err != nil {
			revel.AppLog.Errorf("read last 50 hits' states: %#v", err)
			return c.RenderError(err)
		}
		stateData = append(stateData, [2]interface{}{state, count})
	}

	rows, err = app.DB.Query(app.Q_LAST_50_COUNTIES)
	if err != nil {
		revel.AppLog.Errorf("query last 50 hits' counties: %#v", err)
		return c.RenderError(err)
	}
	for rows.Next() {
		var (
			state  string
			county string
			count  int
		)
		err := rows.Scan(&state, &county, &count)
		if err != nil {
			revel.AppLog.Errorf("read last 50 hits' counties: %#v", err)
			return c.RenderError(err)
		}
		countyData = append(countyData, [3]interface{}{state, county, count})
	}
	return c.Render(denomData, seriesData, stateData, countyData)
}

// vim:foldmethod=marker:
