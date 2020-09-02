package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"database/sql"
	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/routes"
	"github.com/jeff-blank/wg/app/util"
	"github.com/revel/revel"
)

const (
	// SQL queries {{{
	Q_HBS = `
		select
			country,
			state,
			count(state) as count
		from
			hits
		where
			country in ('US', 'Canada')
		group by
			country, state
		order by
			count desc,
			state
	`

	Q_HBC = `
		select
			country,
			count(country) as count
		from
			hits
		where
			country not in ('US', 'Canada')
		group by
			country
		order by
			count desc
	`

	Q_BILL = `
		select
			id,
			serial,
			denomination,
			series,
			rptkey
		from
			bills
		where
			rptkey = $1 and
			denomination = $2 and
			series = $3
	`

	Q_HIT_BY_ID = `
		select
			b.serial,
			b.denomination,
			b.series,
			b.rptkey,
			h.country,
			h.state,
			h.county,
			h.entdate
		from
			bills b,
			hits h
		where
			h.id = %d and
			h.bill_id = b.id
	`

	Q_BREAKDOWN_US_CA = `
		select distinct
			state,
			count(1) as count
		from
			hits
		where
			country = $1
		group by
			state
		order by
			count desc,
			state
	`

	Q_BREAKDOWN_OTHER = `
		select distinct
			country,
			count(1) as count
		from
			hits
		where
			country not in ('US', 'Canada')
		group by
			country
		order by
			count desc,
			country
	`

	Q_REGION_BREAKDOWN = `
		select
			county,
			count(1)
		from
			hits
		where
			country = $1 and
			state = $2
		group by
			county
		order by
			count desc,
			county
	`

	S_INSERT_BILL = `insert into bills (serial, series, denomination, rptkey)values($1, $2, $3, $4)`
	S_INSERT_HIT  = `insert into hits (bill_id, country, state, county, entdate) values ($1, $2, $3, $4, $5)`

	// }}}

	START_YEAR = 2003
)

type Hits struct {
	*revel.Controller
}

func (c Hits) Index() revel.Result {

	var newHitFlash app.NewHitInfo

	if c.Flash.Data["info"] != "" {
		revel.AppLog.Debugf("%#v", c.Flash.Data)
		json.Unmarshal([]byte(c.Flash.Data["info"]), &newHitFlash)
	}
	revel.AppLog.Debugf("%#v", newHitFlash)

	filterCountry := c.Params.Get("country")
	filterState := c.Params.Get("state")
	filterCounty := c.Params.Get("county")
	filterYear := c.Params.Get("year")
	filterSort := c.Params.Get("sort")

	where := ""

	if filterCountry != `` {
		where += ` and country = '` + app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(filterCountry, `\$1`), `''`) + `'`
	}
	if filterState != `` {
		where += ` and state = '` + app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(filterState, `\$1`), `''`) + `'`
	}
	if filterCounty != `` {
		where += ` and county = '` + app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(filterCounty, `\$1`), `''`) + `'`
	}
	if filterYear != `` {
		var year string
		if filterYear == "current" {
			// get current year
			year = time.Now().Format("2006")
		} else {
			year = filterYear
		}
		where += ` and substr(entdate::varchar, 1, 4) = '` + app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(year, `\$1`), `''`) + `'`
		filterYear = year
	}

	order := " "
	if filterSort == "asc" {
		order += "order by entdate, id"
	} else {
		filterSort = ""
		order += "order by entdate desc, id desc"
	}

	hits, err := util.GetHits(where + order)
	if err != nil {
		return c.RenderError(err)
	}
	rowCount := len(hits)
	if filterSort == "asc" {
		for i := 0; i < rowCount; i++ {
			hits[i].Index = i + 1
		}
	} else {
		for i := 0; i < rowCount; i++ {
			hits[rowCount-1-i].Index = i + 1
		}
	}

	// combine filters for ease of passing data to template
	filters := make(map[string]string)
	filters = map[string]string{
		"country": filterCountry,
		"state":   filterState,
		"county":  filterCounty,
		"year":    filterYear,
		"sort":    filterSort,
		"serial":  "",
	}
	links := make(map[string]string)
	links = map[string]string{
		"allHits":         routes.Hits.Index(),
		"currentYearHits": routes.Hits.Index() + "?year=current",
		"breakdown":       routes.Hits.Breakdown(),
		"new":             routes.Hits.New(),
	}
	return c.Render(hits, newHitFlash, filters, links)
}

func (c Hits) Breakdown() revel.Result {
	links := make(map[string]string)
	links = map[string]string{
		"allHits":         routes.Hits.Index(),
		"currentYearHits": routes.Hits.Index() + "?year=current",
		"breakdown":       routes.Hits.Breakdown(),
		"new":             routes.Hits.New(),
	}

	breakdowns := make(map[string][]app.HitsBrkEnt)
	breakdowns["US"] = make([]app.HitsBrkEnt, 0)
	breakdowns["Canada"] = make([]app.HitsBrkEnt, 0)
	breakdowns["Other"] = make([]app.HitsBrkEnt, 0)

	for _, hitSet := range [3]string{"US", "Canada", "Other"} {
		var rows *sql.Rows
		var err error
		if hitSet == "Other" {
			rows, err = app.DB.Query(Q_BREAKDOWN_OTHER)
		} else {
			rows, err = app.DB.Query(Q_BREAKDOWN_US_CA, hitSet)
		}
		if err != nil {
			revel.AppLog.Error(err.Error())
			return c.RenderError(err)
		}
		defer rows.Close()
		for rows.Next() {
			var (
				region string
				count  int
			)
			err := rows.Scan(&region, &count)
			if err != nil {
				revel.AppLog.Errorf("%v", err)
			} else {
				breakdowns[hitSet] = append(breakdowns[hitSet], app.HitsBrkEnt{Region: region, Count: count})
			}
		}
	}

	return c.Render(breakdowns, links)
}

func (c Hits) ShowBrk() revel.Result {
	var state string

	results := make([]app.HitsBrkEnt, 0)
	country := c.Params.Get("country")
	if country == "US" || country == "Canada" {
		state = c.Params.Get("region")
	} else {
		state = "--"
	}

	rows, err := app.DB.Query(Q_REGION_BREAKDOWN, country, state)
	if err != nil {
		msg := fmt.Sprintf("query breakdown: %#v", err)
		revel.AppLog.Error(msg)
		return c.RenderText(msg)
	}

	defer rows.Close()
	for rows.Next() {
		var (
			county string
			count  int
		)
		err = rows.Scan(&county, &count)
		if err != nil {
			msg := fmt.Sprintf("read breakdown: %#v", err)
			revel.AppLog.Error(msg)
			return c.RenderText(msg)
		} else {
			results = append(results, app.HitsBrkEnt{Region: county, Count: count})
		}
	}

	return c.RenderJSON(results)
}

func dbSanitize(input string) string {
	return app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(input, `\$1`), `''`)
}

func (c Hits) Create() revel.Result {
	var (
		bId        int
		bSerial    string
		bDenom     int
		bSeries    string
		bRptkey    string
		series     string
		infoFlash  app.NewHitInfo
		dateHits   int
		countyHits int
	)
	revel.AppLog.Debugf("%#v", c.Params.Form)

	rptkey := dbSanitize(app.RE_whitespace.ReplaceAllString(c.Params.Form["key"][0], ""))

	country := app.RE_trailingWhitespace.ReplaceAllString(app.RE_leadingWhitespace.ReplaceAllString(c.Params.Form["country"][0], ""), "")
	state := app.RE_trailingWhitespace.ReplaceAllString(app.RE_leadingWhitespace.ReplaceAllString(c.Params.Form["state"][0], ""), "")
	county := app.RE_trailingWhitespace.ReplaceAllString(app.RE_leadingWhitespace.ReplaceAllString(c.Params.Form["county"][0], ""), "")

	serial := dbSanitize(app.RE_whitespace.ReplaceAllString(c.Params.Form["serial"][0], ""))
	if !app.RE_serial.MatchString(serial) {
		return c.RenderText("invalid serial number")
	}
	if s, ok := c.Params.Form["series"]; ok {
		series = dbSanitize(app.RE_whitespace.ReplaceAllString(s[0], ""))
	} else if s, ok := app.SeriesByLetter[serial[:1]]; ok {
		series = s
	} else {
		return c.RenderText("missing bill series")
	}
	denom, _ := strconv.Atoi(dbSanitize(app.RE_whitespace.ReplaceAllString(c.Params.Form["denom"][0], "")))
	entdate := dbSanitize(c.Params.Form["year"][0]) + "-" + dbSanitize(c.Params.Form["month"][0]) + "-" + dbSanitize(c.Params.Form["day"][0])

	err := app.DB.QueryRow(`select count(1) from hits where substr(entdate::text, 6) = $1`, entdate[5:]).Scan(&dateHits)
	if err == nil && dateHits == 0 {
		infoFlash.FirstOnDate = entdate[5:]
	} else if err != nil {
		revel.AppLog.Errorf("is first date hit? err=%#v", err)
	}
	if country == "US" {
		err := app.DB.QueryRow(`select count(1) from hits where country = 'US' and state = $1 and county = $2`, state, county).Scan(&countyHits)
		if err == nil && countyHits == 0 {
			infoFlash.FirstInCounty = fmt.Sprintf("%s, %s", county, state)
			rows, err := app.DB.Query(app.Q_BINGOS, state, county)
			if err == nil {
				infoFlash.CountyBingoNames = make([]string, 0)
				defer rows.Close()
				for rows.Next() {
					var bingo string
					err := rows.Scan(&bingo)
					if err == nil {
						infoFlash.CountyBingoNames = append(infoFlash.CountyBingoNames, bingo)
					}
				}
			} else if err != nil {
				revel.AppLog.Errorf("county in bingos err=%#v", err)
			}
			borderCounties, err := util.GetAdjacentWithHits(state, county)
			if err != nil {
				revel.AppLog.Errorf("adjacent counties err=%#v", err)
			} else {
				infoFlash.AdjacentCounties = borderCounties
			}
		} else if err != nil {
			revel.AppLog.Errorf("is first county hit? err=%#v", err)
		}
	}
	flashJson, err := json.Marshal(infoFlash)
	if err == nil {
		c.Flash.Out["info"] = string(flashJson)
	}

	bId = -1
	err = app.DB.QueryRow(Q_BILL, rptkey, denom, series).Scan(&bId, &bSerial, &bDenom, &bSeries, &bRptkey)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			revel.AppLog.Errorf("search for existing bill: %#v", err)
			return c.RenderText(err.Error())
		}
	}

	if bId == -1 {
		res, err := app.DB.Exec(S_INSERT_BILL, serial, series, denom, rptkey)
		if err != nil {
			revel.AppLog.Errorf("insert new bill: %#v", err)
			return c.RenderText(err.Error())
		}
		n, err := res.RowsAffected()
		if err != nil {
			revel.AppLog.Errorf("get # rows affected by bill insert: %#v", err)
			return c.RenderText(err.Error())
		}
		if n != 1 {
			revel.AppLog.Errorf("insert bill failed: %d rows affected", n)
			return c.RenderText(fmt.Sprintf("insert bill failed: %d rows affected", n))
		}
		err = app.DB.QueryRow(Q_BILL, rptkey, denom, series).Scan(&bId, &bSerial, &bDenom, &bSeries, &bRptkey)
		if err != nil {
			revel.AppLog.Errorf("get bill after insert: %#v", err)
			return c.RenderText(err.Error())
		}
	}

	res, err := app.DB.Exec(S_INSERT_HIT, bId, country, state, county, entdate)
	if err != nil {
		revel.AppLog.Errorf("insert new hit: %#v", err)
		return c.RenderText(err.Error())
	}
	n, err := res.RowsAffected()
	if err != nil {
		revel.AppLog.Errorf("get # rows affected by hit insert: %#v", err)
		return c.RenderText(err.Error())
	}
	if n != 1 {
		revel.AppLog.Errorf("insert hit failed: %d rows affected", n)
		return c.RenderText(fmt.Sprintf("insert hit failed: %d rows affected", n))
	}

	return c.Redirect(routes.Hits.Index() + "?year=current")
}

func (c Hits) New() revel.Result {
	var months [12]string
	var days [31]string

	now := time.Now().Format("2006-01-02")
	year, _ := strconv.Atoi(now[0:4])
	month := now[5:7]
	day := now[8:]

	years := make([]int, (year+1)-(START_YEAR-1))
	for y, _ := range years {
		years[y] = START_YEAR + y
	}
	for m, _ := range months {
		months[m] = fmt.Sprintf("%02d", m+1)
	}
	for d, _ := range days {
		days[d] = fmt.Sprintf("%02d", d+1)
	}

	states, err := util.GetStates("US")
	if err != nil {
		return c.RenderError(err)
	}
	homeState := util.GetHomeRegion("state")
	return c.Render(states, homeState, years, year, months, month, days, day)
}

// vim:foldmethod=marker:
