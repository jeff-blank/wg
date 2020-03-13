package util

import (
	"fmt"
	"math"
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

	DATE_LIST_LAYOUT  = `2006-01-02`
	STATS_START_YEAR  = 2003
	STATS_START_MONTH = time.November
)

func GetStates(country string) []string {
	var stateColumn, table string

	if country == "US" {
		stateColumn = "state"
		table = "counties_master"
	} else if country == "Canada" {
		stateColumn = "abbr"
		table = "provinces"
	} else {
		return nil
	}

	query := fmt.Sprintf(Q_REGIONS, stateColumn, "", table, stateColumn)

	states := make([]string, 0)
	rows, err := app.DB.Query(query)
	if err != nil {
		// TODO: render an error
		revel.AppLog.Error(err.Error())
		return nil
	}
	for rows.Next() {
		var state string
		err := rows.Scan(&state)
		if err != nil {
			revel.AppLog.Errorf("%v", err)
			return nil
		} else {
			states = append(states, state)
		}
	}
	return states
}

func GetCounties(state string) []app.Region {
	query := fmt.Sprintf(Q_REGIONS, "id,", "county", "counties_master where state='"+state+"'", "county")
	counties := make([]app.Region, 0)
	rows, err := app.DB.Query(query)
	if err != nil {
		// TODO: render an error
		revel.AppLog.Error(err.Error())
		return nil
	}
	for rows.Next() {
		var (
			id     int
			county string
		)
		err := rows.Scan(&id, &county)
		if err != nil {
			revel.AppLog.Errorf("%v", err)
			return nil
		} else {
			counties = append(counties, app.Region{Id: id, Region: county})
		}
	}
	return counties
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
		// TODO: render an error
		revel.AppLog.Errorf("prepare q_hitCount: %v", err)
		return nil
	}
	q_entCount, err := PrepMonthEnts()
	if err != nil {
		// TODO: render an error
		revel.AppLog.Errorf("%v", err)
		revel.AppLog.Errorf("prepare q_entCount: %v", err)
		return nil
	}

	//m := len(months) - 1
	totalHits := 0
	totalBills := 0
	for m, monthStr := range months {
		var billCount, hitCount int
		err := q_hitCount.QueryRow(monthStr[:7]).Scan(&hitCount)
		//revel.AppLog.Debugf("%s: %d hits", monthStr[:7], hitCount)
		if err != nil {
			// TODO: render an error
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
		if err != nil && err.Error() != "sql: no rows in result set" {
			// TODO: render an error
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

		//m--
	}

	avgHitsPerMonth := float64(totalHits) / float64(len(months))
	prevYear := ""
	monthsInYear := 0
	yearHits := 0
	for m, _ := range months {
		year := allData[m]["month"].(string)[:4]
		if (prevYear != year && prevYear != "") || m == len(months)-1 {
			// new year
			firstMonthInd := m - 1 // - monthsInYear
			if m == len(months)-1 {
				firstMonthInd++
				monthsInYear++
				yearHits += allData[m]["monthHits"].(int)
			}
			allData[firstMonthInd]["monthsInYear"] = monthsInYear
			allData[firstMonthInd]["yearHits"] = yearHits
			monthsInYear = 0
			yearHits = 0
		}
		if m != len(months)-1 {
			monthsInYear++
			yearHits += allData[m]["monthHits"].(int)
		}
		allData[m]["straightLineAvgHits"] = float64(m+1) * avgHitsPerMonth
		prevYear = allData[m]["month"].(string)[:4]
	}

	//revel.AppLog.Debugf("%#v", allData[0])

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

// vim:foldmethod=marker:
