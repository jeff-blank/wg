package util

import (
	"errors"
	"fmt"
	"math"
	s "strings"
	"time"

	"database/sql"
	"github.com/jeff-blank/wg/app"
	"github.com/revel/revel"
)

const (
	Q_REGIONS = `
		select distinct
			%s %s
		from
			%s
		order by
			%s
	`

	Q_HOME_REGION = `
		select
			%s
		from
			counties_master cm,
			residences r
		where
			r.label = '_all' and
			r.home = cm.id
	`

	Q_RESLIST = `
		select
			label
		from
			residences
		where
			label <> '_all'
		order by
			label
	`

	Q_CURRENT_RESIDENCE = `
		select
			cr.label
		from
			residences cr,
			residences ar
		where
			ar.label = '_all' and
			cr.label <> '_all' and
			ar.home = cr.home
	`

	DATE_LIST_LAYOUT  = `2006-01-02`
	STATS_START_YEAR  = 2003
	STATS_START_MONTH = time.November
)

func GetStates(country string) ([]string, error) {
	var stateColumn, table string

	if country == "US" {
		stateColumn = "state"
		table = "counties_master"
	} else if country == "Canada" {
		stateColumn = "abbr"
		table = "provinces"
	} else {
		return nil, nil
	}

	query := fmt.Sprintf(Q_REGIONS, stateColumn, "", table, stateColumn)

	states := make([]string, 0)
	rows, err := app.DB.Query(query)
	if err != nil {
		revel.AppLog.Errorf("%v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var state string
		err := rows.Scan(&state)
		if err != nil {
			revel.AppLog.Errorf("%v", err)
			return nil, err
		} else {
			states = append(states, state)
		}
	}
	return states, nil
}

func GetCounties(state string) ([]app.Region, error) {
	query := fmt.Sprintf(Q_REGIONS, "id,", "county", "counties_master where state='"+state+"'", "county")
	counties := make([]app.Region, 0)
	rows, err := app.DB.Query(query)
	if err != nil {
		revel.AppLog.Errorf("%v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id     int
			county string
		)
		err := rows.Scan(&id, &county)
		if err != nil {
			revel.AppLog.Errorf("%v", err)
			return nil, err
		} else {
			counties = append(counties, app.Region{Id: id, Region: county})
		}
	}
	return counties, nil
}

func GetCountyById(id int) (app.USCounty, error) {
	var (
		id_db  int
		state  string
		county string
	)
	err := app.DB.QueryRow(`select id, state, county from counties_master where id=$1`, id).Scan(&id_db, &state, &county)
	if err != nil {
		revel.AppLog.Error(err.Error())
		return app.USCounty{}, err
	}
	return app.USCounty{Id: id_db, State: state, County: county}, nil
}

func GetHomeRegion(regionColumn string) string {
	var homeRegion string

	query := fmt.Sprintf(Q_HOME_REGION, regionColumn)
	err := app.DB.QueryRow(query).Scan(&homeRegion)
	if err != nil {
		// TODO: render an error
		revel.AppLog.Error(err.Error())
		return ""
	}
	return homeRegion
}

func PrepMonthEnts() (*sql.Stmt, error) {
	return app.DB.Prepare(`select bills from entries where month = $1`)
}

func StatsData(returnType string) interface{} {
	months := getMonths()
	oneYrHits := make([]int, 0)
	oneYrBills := make([]int, 0)

	allData := make([]map[string]interface{}, len(months))

	q_hitCount, err := app.DB.Prepare(`select count(1) from hits where substr(entdate::text, 1, 7) = $1`)
	if err != nil {
		// TODO: return an error
		revel.AppLog.Errorf("prepare q_hitCount: %v", err)
		return nil
	}
	q_entCount, err := PrepMonthEnts()
	if err != nil {
		// TODO: return an error
		revel.AppLog.Errorf("prepare q_entCount: %v", err)
		return nil
	}

	totalHits := 0
	totalBills := 0
	for m, monthStr := range months {
		var billCount, hitCount int
		err := q_hitCount.QueryRow(monthStr[:7]).Scan(&hitCount)
		//revel.AppLog.Debugf("%s: %d hits", monthStr[:7], hitCount)
		if err != nil {
			// TODO: return an error
			revel.AppLog.Errorf("query q_hitCount: %v", err)
			return nil
		}

		allData[m] = make(map[string]interface{})
		allData[m]["month"] = monthStr
		allData[m]["monthHits"] = hitCount
		oneYrHits = append(oneYrHits, hitCount)
		if len(oneYrHits) > 12 {
			oneYrHits = oneYrHits[1:]
		}
		totalHits += hitCount
		allData[m]["cumulativeHits"] = totalHits
		allData[m]["avgMonthlyHits"] = float64(totalHits) / float64(m+1)
		allData[m]["oneYrAvgMonthlyHits"] = avgVal(oneYrHits)

		err = q_entCount.QueryRow(monthStr).Scan(&billCount)
		if err != nil && err.Error() != app.SQL_ERR_NO_ROWS {
			// TODO: return an error
			revel.AppLog.Errorf("query q_entCount: %v", err)
			return nil
		}
		allData[m]["monthBills"] = billCount
		oneYrBills = append(oneYrBills, billCount)
		if len(oneYrBills) > 12 {
			oneYrBills = oneYrBills[1:]
		}
		totalBills += billCount
		allData[m]["cumulativeBills"] = totalBills
		allData[m]["oneYrAvgMonthlyBills"] = avgVal(oneYrBills)

		allData[m]["score"] = wgScore(totalBills, totalHits)
	}

	avgHitsPerMonth := float64(totalHits) / float64(len(months))
	prevYear := ""
	monthsInYear := 0
	yearHits := 0
	for m := range months {
		var statsMonthInd int

		year := allData[m]["month"].(string)[:4]
		if prevYear != year && prevYear != "" {
			// new year
			// store year-total stats (entry and hit counters) in final month of previous year
			statsMonthInd = m - 1
			allData[statsMonthInd]["monthsInYear"] = monthsInYear
			allData[statsMonthInd]["yearHits"] = yearHits
			monthsInYear = 0
			yearHits = 0
		}
		monthsInYear++
		yearHits += allData[m]["monthHits"].(int)
		allData[m]["straightLineAvgHits"] = float64(m+1) * avgHitsPerMonth
		if m == len(months)-1 {
			allData[m]["monthsInYear"] = monthsInYear
			allData[m]["yearHits"] = yearHits
		}
		prevYear = allData[m]["month"].(string)[:4]
	}

	if returnType == "table" {
		return allData
	} else {
		return nil
	}
}

func getMonths() []string {
	now := time.Now()
	months := make([]string, 0)
	monthEndDate := time.Date(STATS_START_YEAR, STATS_START_MONTH+1, 1, 0, -1, 0, 0, time.UTC)
	for {
		months = append(months, monthEndDate.Format(DATE_LIST_LAYOUT))
		monthEndDate = monthEndDate.Add(time.Minute).AddDate(0, 1, 0).Add(-1 * time.Minute)
		if monthEndDate.Year() > now.Year() || (monthEndDate.Year() == now.Year() && monthEndDate.Month() > now.Month()) {
			break
		}
	}
	return months
}

func avgVal(values []int) float64 {
	sum := 0
	for _, val := range values {
		sum += val
	}
	return float64(sum) / float64(len(values))
}

func wgScore(ents, hits int) float64 {
	return 100 * (math.Sqrt(math.Log(float64(ents))) + math.Log(float64(hits+1)))
}

func GetFirstHits(regionType string) ([]app.Hit, error) {
	if regionType == "county" {
		return GetHits(`and h.id IN (SELECT h2.id
          FROM hits h2, counties_master cm2
          WHERE h2.country='US' AND h.county_id=h2.county_id AND h2.county_id=cm2.id AND cm2.state not in ('DC', 'AE', 'GU', 'PR')
          ORDER BY h2.entdate, h2.id LIMIT 1) ORDER BY h.entdate desc, h.id desc`)
	} else if regionType == "state" {
		return GetHits(`and h.id IN (SELECT h2.id
          FROM hits h2, counties_master cm2
          WHERE h2.country='US' AND h2.county_id=cm2.id AND cm.state=cm2.state
          ORDER BY h2.entdate LIMIT 1) ORDER BY h.entdate desc`)
	} else {
		return nil, nil
	}
}

func GetHits(whereGroupOrder string) ([]app.Hit, error) {
	var newHit app.Hit
	hits := make([]app.Hit, 0)

	revel.AppLog.Debugf(app.Q_HITS + whereGroupOrder)
	rows, err := app.DB.Query(app.Q_HITS + whereGroupOrder)
	if err != nil {
		revel.AppLog.Errorf("query:\n%s\nerror: %#v", app.Q_HITS+whereGroupOrder, err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(
			&newHit.Id,
			&newHit.BillId,
			&newHit.Denom,
			&newHit.Serial,
			&newHit.Series,
			&newHit.RptKey,
			&newHit.Residence,
			&newHit.EntDate,
			&newHit.Country,
			&newHit.State,
			&newHit.County,
			&newHit.City,
			&newHit.ZIP,
			&newHit.Count,
			&newHit.WGHitNum,
		)
		if err != nil {
			revel.AppLog.Errorf("%v", err)
			return nil, err
		} else {
			newHit.EntDate = newHit.EntDate[0:10]
			hits = append(hits, newHit)
		}
	}
	return hits, nil
}

func GetAdjacentWithHits(state, county string) ([]string, error) {
	var (
		hitState  string
		hitCounty string
	)
	counties := make([]string, 0)

	rows, err := app.DB.Query(app.Q_ADJACENT_COUNTIES, state, county)
	if err != nil {
		if err.Error() == app.SQL_ERR_NO_ROWS {
			return counties, nil
		}
		return nil, err
	}
	for rows.Next() {
		err := rows.Scan(&hitState, &hitCounty)
		if err != nil {
			return nil, err
		}
		revel.AppLog.Debugf("checking %s, %s", county, state)
		hits, _ := GetHits("and cm.state='" + hitState + "' and cm.county='" + hitCounty + "'")
		if len(hits) > 0 {
			counties = append(counties, fmt.Sprintf("%s, %s", hitCounty, hitState))
		}
	}

	return counties, nil
}

func GetCurrentResidence() (string, error) {
	var residence string

	err := app.DB.QueryRow(Q_CURRENT_RESIDENCE).Scan(&residence)
	if err != nil {
		if err.Error() != app.SQL_ERR_NO_ROWS {
			revel.AppLog.Errorf("GetCurrentResidence(): query current residence: %#v", err)
		}
		return "", err
	} else {
		return residence, nil
	}
}

func GetResidences() ([]string, error) {
	rows, err := app.DB.Query(Q_RESLIST)
	if err != nil {
		if err.Error() != app.SQL_ERR_NO_ROWS {
			revel.AppLog.Errorf("GetResidences(): query residence list: %#v", err)
		}
		return nil, err
	} else {
		resList := make([]string, 0)
		for rows.Next() {
			var residence string
			err := rows.Scan(&residence)
			if err != nil {
				revel.AppLog.Errorf("GetResidences(): scan residence: %#v", err)
				resList = nil
				break
			}
			resList = append(resList, residence)
		}
		return resList, nil
	}
}

func GetTZList() ([]app.TZRec, error) {
	var (
		tz      string
		tzDescr string
	)

	tzList := make([]app.TZRec, 0)
	rows, err := app.DB.Query(`select display_name, tz_name from tz order by display_name`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		err = rows.Scan(&tzDescr, &tz)
		if err != nil {
			return nil, err
		}
		tzList = append(tzList, app.TZRec{TZDescr: tzDescr, TZString: tz})
	}
	return tzList, nil
}

func SetTimeZone(tz_in string) error {
	tzList, err := GetTZList()
	if err != nil {
		return err
	}
	for _, tzRec := range tzList {
		if tzRec.TZString == tz_in {
			res, err := app.DB.Exec(`update user_prefs set tz=$1`, tzRec.TZString)
			if err != nil {
				revel.AppLog.Error(err.Error())
				return err
			}
			rows, err := res.RowsAffected()
			if err != nil {
				revel.AppLog.Error(err.Error())
				return err
			}
			if rows < 1 {
				return errors.New("Update failed")
			}
			return nil
		}
	}
	return errors.New("Time zone '" + tz_in + "' not found")
}

func GetPrefs() ([]app.UserPrefs, error) {
	var tz string

	rows, err := app.DB.Query(`select coalesce(tz, '') as tz from user_prefs`)
	if err != nil {
		if err.Error() != app.SQL_ERR_NO_ROWS {
			return nil, err
		} else {
			// no rows found
			revel.AppLog.Debug("GetPrefs(): no rows selected")
			return nil, nil
		}
	}
	i := 0
	for rows.Next() {
		i++
		if i > 1 {
			errMsg := "GetPrefs(): found at least 2 rows"
			revel.AppLog.Error(errMsg)
			return nil, errors.New(errMsg)
		}
		err = rows.Scan(&tz)
		if err != nil {
			return nil, err
		}
	}
	revel.AppLog.Debugf("GetPrefs(): found %d rows", i)
	if i == 0 {
		revel.AppLog.Debug("GetPrefs(): no rows found")
		return nil, nil
	}
	prefs := make([]app.UserPrefs, 1)
	prefs[0] = app.UserPrefs{TZString: tz}
	return prefs, nil
}

func GetFRBFromSerial(serial string) string {
	if app.RE_serial_11.MatchString(serial) {
		return serial[1:2]
	} else if app.RE_serial_10.MatchString(serial) {
		return serial[:1]
	}
	return ""
}

func GetStateCountyCityFromZIP(zip string) (string, string, string, error) {
	var state, county, city string
	err := app.DB.QueryRow(`select cm.state, cm.county, h.city from hits h, counties_master cm where h.county_id=cm.id and h.zip=$1`, zip).Scan(&state, &county, &city)
	if err != nil && err.Error() != app.SQL_ERR_NO_ROWS {
		return "", "", "", err
	}
	return state, county, city, nil
}

func CountyHasHits(id int) (bool, error) {
	var count int

	err := app.DB.QueryRow(`select count(1) from hits where county_id = $1`, id).Scan(&count)
	if err != nil && err.Error() != app.SQL_ERR_NO_ROWS {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func GetSerialSeries(serial_in, series_in string, denom int) (string, string, error) {
	revel.AppLog.Debugf("serial_in='%s'", serial_in)

	serial := s.ToUpper(DbSanitize(app.RE_whitespace.ReplaceAllLiteralString(serial_in, "")))
	if !app.RE_serial.MatchString(serial) {
		return "", "", errors.New("util.GetSerialSeries(): invalid serial number '" + serial + "'")
	}

	revel.AppLog.Debugf("serial_in='%s' -> serial='%s'", serial_in, serial)

	series := s.ToUpper(series_in)

	isSerial10 := app.RE_serial_10.MatchString(serial)
	isSerial11 := app.RE_serial_11.MatchString(serial)

	revel.AppLog.Debugf("series='%s'", series)
	if series != "" && isSerial10 {
		if app.RE_series.MatchString(series) {
			series = DbSanitize(app.RE_whitespace.ReplaceAllString(series, ""))
		} else {
			return "", "", errors.New("util.GetSerialSeries(): invalid bill series '" + series + "'")
		}
	} else if str, ok := app.SeriesByLetter[serial[:1]]; ok && isSerial11 && denom >= 5 {
		series = str
	} else if isSerial11 && denom < 5 {
		return "", "", errors.New(fmt.Sprintf("invalid bill serial/denomination (serial='%s'; denom='%d')", serial, denom))
	} else {
		return "", "", errors.New(fmt.Sprintf("missing or invalid bill series (series='%s')", series))
	}
	return serial, series, nil
}

func DbSanitize(input string) string {
	return app.RE_singleQuote.ReplaceAllString(app.RE_dbUnsafe.ReplaceAllString(input, `\$1`), `''`)
}

// vim:foldmethod=marker:
