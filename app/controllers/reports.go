package controllers

import (
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
		// TODO: render an error
		revel.AppLog.Errorf("%v", err)
		return c.Render()
	}

	for rows.Next() {
		var date string
		var count int

		err = rows.Scan(&date, &count)
		if err != nil {
			// TODO: render an error
			revel.AppLog.Errorf("%v", err)
			return c.Render()
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
		revel.AppLog.Debugf("%#v", m)
		tableOut = append(tableOut, tableIn[m])
	}
	revel.AppLog.Debugf("%#v", tableOut)

	entsLink := routes.Entries.Edit()
	return c.Render(tableOut, entsLink)
}

// vim:foldmethod=marker:
