package controllers

import (
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
		if err.Error() == "sql: no rows in result set" {
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

// vim:foldmethod=marker:
