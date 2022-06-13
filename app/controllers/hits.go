package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	s "strings"
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

	S_INSERT_BILL = `insert into bills (serial, series, denomination, rptkey, residence)values($1, $2, $3, $4, $5)`
	S_INSERT_HIT  = `insert into hits (bill_id, country, state, county, city, zip, entdate) values ($1, $2, $3, $4, $5, $6, $7)`
	S_UPDATE_HIT  = `update hits set country=$1, state=$2, county=$3, city=$4, zip=$5, entdate=$6 where id=$7`

	// }}}

	START_YEAR = 2003
)

type Hits struct {
	*revel.Controller
}

func (c Hits) Index() revel.Result {

	var flashData app.HitInfo

	if c.Flash.Data["info"] != "" {
		revel.AppLog.Debugf("hits.Index(): flashdatainfo: %#v", c.Flash.Data)
		json.Unmarshal([]byte(c.Flash.Data["info"]), &flashData)
	}
	revel.AppLog.Debugf("hits.Index(): flashData: %#v", flashData)

	filterSerial := app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(s.ToUpper(c.Params.Get("serial")), `\$1`), ``)
	filterDenom := app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(c.Params.Get("denom"), `\$1`), ``)
	filterCountry := app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(c.Params.Get("country"), `\$1`), `''`)
	filterState := app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(c.Params.Get("state"), `\$1`), `''`)
	filterCounty := app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(c.Params.Get("county"), `\$1`), `''`)
	filterCity := app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(c.Params.Get("city"), `\$1`), `''`)
	filterYear := app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(c.Params.Get("year"), `\$1`), `''`)
	filterSort := c.Params.Get("sort")

	where := ""

	if filterSerial != `` {
		where += ` and serial like '` + filterSerial + `'`
	}
	if filterDenom != `` {
		where += ` and denomination = ` + filterDenom
	}
	if filterCountry != `` {
		where += ` and country = '` + filterCountry + `'`
	}
	if filterState == `==` {
		filterState = ``
	}
	if (filterCountry == `US` || filterCountry == `Canada`) && filterState != `` {
		where += ` and state = '` + filterState + `'`
	}
	if filterCounty == `==` {
		filterCounty = ``
	}
	if filterCountry == `US` && filterCounty != `` {
		where += ` and county = '` + filterCounty + `'`
	}
	if filterCity != `` {
		where += ` and city = '` + filterCity + `'`
	}
	if filterYear != `` {
		var year string
		if filterYear == "current" {
			// get current year
			year = time.Now().Format("2006")
		} else {
			year = filterYear
		}
		where += ` and substr(entdate::varchar, 1, 4) = '` + year + `'`
		filterYear = year
	}

	order := " "
	if filterSort == "asc" {
		order += "order by entdate, h.id"
	} else {
		filterSort = ""
		order += "order by entdate desc, h.id desc"
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
		"denom":   filterDenom,
		"country": filterCountry,
		"state":   filterState,
		"county":  filterCounty,
		"year":    filterYear,
		"sort":    filterSort,
		"serial":  filterSerial,
	}
	links := make(map[string]string)
	links = map[string]string{
		"allHits":         routes.Hits.Index(),
		"currentYearHits": routes.Hits.Index() + "?year=current",
		"breakdown":       routes.Hits.Breakdown(),
		"new":             routes.Hits.New(),
	}
	return c.Render(hits, flashData, filters, links)
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
		infoFlash  app.HitInfo
		dateHits   int
		countyHits int
		zip        int
	)
	revel.AppLog.Debugf("%#v", c.Params.Form)

	rptkey := dbSanitize(app.RE_whitespace.ReplaceAllLiteralString(c.Params.Form["key"][0], ""))

	country := app.RE_trailingWhitespace.ReplaceAllLiteralString(app.RE_leadingWhitespace.ReplaceAllLiteralString(c.Params.Form["country"][0], ""), "")
	state := app.RE_trailingWhitespace.ReplaceAllLiteralString(app.RE_leadingWhitespace.ReplaceAllLiteralString(c.Params.Form["state"][0], ""), "")
	county := app.RE_trailingWhitespace.ReplaceAllLiteralString(app.RE_leadingWhitespace.ReplaceAllLiteralString(c.Params.Form["county"][0], ""), "")
	city := app.RE_whitespace.ReplaceAllLiteralString(app.RE_trailingWhitespace.ReplaceAllLiteralString(app.RE_leadingWhitespace.ReplaceAllLiteralString(c.Params.Form["city"][0], ""), ""), " ")
	zipStr := app.RE_nonNumeric.ReplaceAllLiteralString(c.Params.Form["zip"][0], "")
	if zipStr == "" {
		zip = -1
	} else {
		zip, _ = strconv.Atoi(zipStr)
	}

	serial := dbSanitize(app.RE_whitespace.ReplaceAllLiteralString(c.Params.Form["serial"][0], ""))
	if !app.RE_serial.MatchString(serial) {
		return c.RenderText("invalid serial number")
	}

	isSerial10 := app.RE_serial_10.MatchString(serial)
	isSerial11 := app.RE_serial_11.MatchString(serial)

	if str, ok := c.Params.Form["series"]; ok && str[0] != "" && isSerial10 {
		series = s.ToUpper(str[0])
		if app.RE_series.MatchString(series) {
			series = dbSanitize(app.RE_whitespace.ReplaceAllString(series, ""))
		} else {
			return c.RenderText("invalid bill series")
		}
	} else if str, ok := app.SeriesByLetter[serial[:1]]; ok && isSerial11 {
		series = str
	} else {
		return c.RenderText("missing bill series")
	}
	denom, _ := strconv.Atoi(dbSanitize(app.RE_whitespace.ReplaceAllString(c.Params.Form["denom"][0], "")))
	entdate := dbSanitize(c.Params.Form["year"][0]) + "-" + dbSanitize(c.Params.Form["month"][0]) + "-" + dbSanitize(c.Params.Form["day"][0])

	residence := ""
	if r, ok := c.Params.Form["residence"]; ok {
		residences, err := util.GetResidences()
		if err != nil {
			return c.RenderText("error retrieving list of residences")
		}
		for _, res := range residences {
			if res == r[0] {
				residence = res
				break
			}
		}
		if residence == "" {
			return c.RenderText(fmt.Sprintf("residence \"%s\" not found", r[0]))
		}
	}

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
	} else {
		state = "--"
		county = "--"
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
		res, err := app.DB.Exec(S_INSERT_BILL, serial, series, denom, rptkey, residence)
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

	res, err := app.DB.Exec(S_INSERT_HIT, bId, country, state, county, city, zip, entdate)
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
	dateSelData := dateSelPopulate(true)
	return c.Render(dateSelData)
}

func (c Hits) Edit() revel.Result {
	dateSelData := dateSelPopulate(false)

	id := c.Params.Route.Get("id")
	hits, err := util.GetHits(fmt.Sprintf("and h.id=%s", id))
	if err != nil {
		return c.RenderText(fmt.Sprintf("error retrieving hit with id %s: %s", id, err.Error()))
	}
	hit := hits[0]
	return c.Render(hit, dateSelData)
}

func (c Hits) Update() revel.Result {
	var (
		err         error
		updateFlash app.HitInfo
		zip         int
	)

	year := c.Params.Get("year")

	id := c.Params.Route.Get("id")
	delHit := c.Params.Get("delete")
	if delHit == "on" {
		err = del(id)
	} else {
		country := app.RE_trailingWhitespace.ReplaceAllLiteralString(app.RE_leadingWhitespace.ReplaceAllLiteralString(c.Params.Form["country"][0], ""), "")
		state := app.RE_trailingWhitespace.ReplaceAllLiteralString(app.RE_leadingWhitespace.ReplaceAllLiteralString(c.Params.Form["state"][0], ""), "")
		county := app.RE_trailingWhitespace.ReplaceAllLiteralString(app.RE_leadingWhitespace.ReplaceAllLiteralString(c.Params.Form["county"][0], ""), "")
		city := app.RE_whitespace.ReplaceAllLiteralString(app.RE_trailingWhitespace.ReplaceAllLiteralString(app.RE_leadingWhitespace.ReplaceAllLiteralString(c.Params.Form["city"][0], ""), ""), " ")
		zipStr := app.RE_nonNumeric.ReplaceAllLiteralString(c.Params.Form["zip"][0], "")
		if zipStr == "" {
			zip = -1
		} else {
			zip, _ = strconv.Atoi(zipStr)
		}

		date := fmt.Sprintf("%s-%s-%s", year, c.Params.Get("month"), c.Params.Get("day"))
		if !app.RE_date.MatchString(date) {
			return c.RenderText("error in date '" + date + "'")
		}
		err = update(id, country, state, county, city, zip, date)
	}
	if err != nil {
		return c.RenderText(fmt.Sprintf("edit/delete of hit '%s' failed: %s", id, err.Error()))
	}
	updateFlash.GenericMessage = "Hit successfully "
	if delHit == "on" {
		updateFlash.GenericMessage += "deleted"
	} else {
		updateFlash.GenericMessage += "updated"
	}
	flashJson, err := json.Marshal(updateFlash)
	if err == nil {
		c.Flash.Out["info"] = string(flashJson)
	}
	return c.Redirect(routes.Hits.Index() + "?year=" + year)
}

func dateSelPopulate(populateDays bool) app.DateSelData {
	var (
		thisMonthIndex int
		dateSelData    app.DateSelData
	)

	now := time.Now().Format("2006-01-02")
	dateSelData.Year, _ = strconv.Atoi(now[0:4])
	dateSelData.Month = now[5:7]
	dateSelData.Day = now[8:]

	dateSelData.Years = make([]int, (dateSelData.Year+1)-(START_YEAR-1))
	for y := range dateSelData.Years {
		dateSelData.Years[y] = START_YEAR + y
	}
	for m := range dateSelData.Months {
		dateSelData.Months[m] = fmt.Sprintf("%02d", m+1)
		if dateSelData.Months[m] == dateSelData.Month {
			thisMonthIndex = m
		}
	}
	if populateDays {
		daysInMonth := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
		dateSelData.Days = make([]string, daysInMonth[thisMonthIndex])
		for d := range dateSelData.Days {
			dateSelData.Days[d] = fmt.Sprintf("%02d", d+1)
		}
	}
	return dateSelData
}

func del(id string) error {
	revel.AppLog.Debugf("deleting %s", id)
	// figure out if this is the only hit on the bill; delete bill if so
	hits, err := util.GetHits(fmt.Sprintf("and h.id=%s", id))
	if err != nil {
		revel.AppLog.Errorf("hits.del(): get record for hit with id %s: %#v", id, err)
		return err
	}
	if len(hits) != 1 {
		err := errors.New(fmt.Sprintf("hits.del(): found %d records for hit with id %s not found", len(hits), id))
		revel.AppLog.Errorf("%#v", err)
		return err
	}
	delHit := hits[0]
	allHitsOnBill, err := util.GetHits(fmt.Sprintf("and b.rptkey='%s'", delHit.RptKey))
	if err != nil {
		revel.AppLog.Errorf("get hit count for bill with key %s (via hit id %s): %#v", delHit.RptKey, id, err)
		return err
	}
	if len(allHitsOnBill) < 1 {
		err := errors.New(fmt.Sprintf("found %d hits on bill with key %s (via hit id %s); cannot delete", len(allHitsOnBill), delHit.RptKey, id))
		revel.AppLog.Errorf("%#v", err)
		return err
	}
	res, err := app.DB.Exec("delete from hits where id=$1", id)
	if err != nil {
		revel.AppLog.Errorf("failed to delete hit with id %s: %#v", id, err)
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		revel.AppLog.Errorf("delete hit with id %s: RowsAffected(): %#v", id, err)
		return err
	}
	if n != 1 {
		err := errors.New(fmt.Sprintf("delete hit with id %s: deleted %d rows", id, n))
		revel.AppLog.Errorf("%#v", err)
		return err
	}
	if len(allHitsOnBill) == 1 {
		res, err := app.DB.Exec("delete from bills where rptkey=$1", delHit.RptKey)
		if err != nil {
			revel.AppLog.Errorf("failed to delete bill with key %s: %#v", delHit.RptKey, err)
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			revel.AppLog.Errorf("delete bill with key %s: RowsAffected(): %#v", delHit.RptKey, err)
			return err
		}
		if n != 1 {
			err := errors.New(fmt.Sprintf("delete bill with key %s: deleted %d bills instead of 1", delHit.RptKey, n))
			revel.AppLog.Errorf("%#v", err)
			return err
		}
	}
	return nil
}

func update(id, country, state, county, city string, zip int, date string) error {
	revel.AppLog.Debugf("updating hit id '%s'", id)
	if country != "US" {
		state = "--"
		county = "--"
	}
	res, err := app.DB.Exec(S_UPDATE_HIT, country, state, county, city, zip, date, id)
	if err != nil {
		revel.AppLog.Errorf("hits.update(): update hit %s: %#v", id, err)
		revel.AppLog.Errorf("hits.update(): query:\n\t%s\n\t'%s', '%s', '%s', '%s', '%s', '%s'",
			S_UPDATE_HIT, country, state, county, city, date, id)
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		revel.AppLog.Errorf("hits.update(): hit %s: check rows affected: %#v", id, err)
	}
	if n != 1 {
		errmsg := fmt.Sprintf("%d rows affected (should be 1)", n)
		revel.AppLog.Errorf("hits.update(): hit %s: %s", id, errmsg)
		err := errors.New(errmsg)
		return err
	}
	return nil
}

// vim:foldmethod=marker:
