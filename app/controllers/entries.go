package controllers

import (
	"fmt"
	"strconv"

	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/routes"
	"github.com/jeff-blank/wg/app/util"
	"github.com/revel/revel"
)

type Entries struct {
	*revel.Controller
}

func (c Entries) Update() revel.Result {
	entsInt, _ := strconv.Atoi(c.Params.Form["ents"][0])
	res, err := app.DB.Exec(`update entries set bills = $1 where month = $2`, entsInt, c.Params.Form["month"][0])
	if err != nil {
		revel.AppLog.Errorf("update entries: %#v", err)
		return c.RenderText("error updating entries")
	}
	n, err := res.RowsAffected()
	if err != nil {
		revel.AppLog.Errorf("get rows affected: %#v", err)
		return c.RenderText("error verifying update")
	}
	if n != 1 {
		revel.AppLog.Errorf("update entries: %d rows", n)
		return c.RenderText("wrong number of rows updated")
	}
	return c.Redirect(routes.Reports.MasterStats())
}

func (c Entries) Edit() revel.Result {
	var entries, totalEnts int

	rows, err := app.DB.Query(`select bills from entries`)
	if err != nil {
		revel.AppLog.Errorf("get total entries: %#v", err)
		return c.RenderText("error getting total bill entries")
	}
	defer rows.Close()
	for rows.Next() {
		var ents int
		err := rows.Scan(&ents)
		if err != nil {
			revel.AppLog.Errorf("get total entries: read month: %#v", err)
			return c.RenderText("error getting total bill entries by month")
		}
		totalEnts += ents
	}
	month := c.Params.Get("month")
	q, err := util.PrepMonthEnts()
	if err != nil {
		revel.AppLog.Errorf("prepare: %#v", err)
		return c.RenderText("error preparing month query")
	}
	err = q.QueryRow(month).Scan(&entries)
	if err != nil {
		if err.Error() == app.SQL_ERR_NO_ROWS {
			res, err := app.DB.Exec(`insert into entries(month, bills)values($1, $2)`, month, 0)
			if err != nil {
				revel.AppLog.Errorf("create month in 'entries': %#v", err)
				return c.RenderText("error creating month")
			}
			n, err := res.RowsAffected()
			if err != nil {
				revel.AppLog.Errorf("verify creaton of month: %#v", err)
				return c.RenderText("error verifying creation of month")
			}
			if n != 1 {
				revel.AppLog.Errorf("create month: %d rows", n)
				return c.RenderText("create month: " + strconv.Itoa(int(n)) + " rows")
			}
		} else {
			revel.AppLog.Errorf("query: %#v", err)
			return c.RenderText("error executing month query")
		}
	}
	updateLink := routes.Entries.Update()
	return c.Render(month, entries, totalEnts, updateLink)
}

func (c Entries) GetEntryByKey(key string) revel.Result {
	var (
		bill      app.Bill
		id        int
		serial    string
		denom     int
		series    string
		residence string
	)

	err := app.DB.QueryRow(`select id, serial, denomination, series, residence from bills where rptkey=$1`, key).Scan(&id, &serial, &denom, &series, &residence)
	if err == nil {
		bill.Id = id
		bill.Serial = serial
		bill.Denomination = denom
		bill.Series = series
		bill.Residence = residence
		bill.Rptkey = key
	} else if err.Error() != app.SQL_ERR_NO_ROWS {
		msg := fmt.Sprintf("look up bill with key '%s': %#v", key, err)
		revel.AppLog.Error(msg)
		bill.Message = msg
	}
	return c.RenderJSON(bill)
}

// vim:foldmethod=marker:
