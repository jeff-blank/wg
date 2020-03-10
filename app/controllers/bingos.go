package controllers

import (
	"fmt"
	"strconv"
	s "strings"

	"database/sql"
	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/routes"
	"github.com/revel/revel"
)

type Bingos struct {
	*revel.Controller
}

const (
	// {{{
	Q_BINGO_LIST = `
		select
			b.id,
			b.name,
			count(1)
		from
			bingos b,
			bingo_counties bc
		where
			b.id = bc.bingo_id
		group by
			b.id,
			b.name
		order by
			b.name
`

	Q_BINGO_COUNTIES = `
		select
			bc.county_id,
			cm.state,
			cm.county
		from
			bingo_counties bc,
			counties_master cm
		where
			bc.bingo_id = $1 and
			bc.county_id = cm.id
		order by
			cm.state,
			cm.county
`

	Q_COUNTY_HAS_HITS = `
		select
			count(1)
		from
			hits h,
			counties_master cm
		where
			cm.id = $1 and
			cm.state = h.state and
			cm.county = h.county
`
	// }}}
)

func (c Bingos) Index() revel.Result {
	bingoList := make([]app.BingoSummary, 0)
	foundBingos := true
	newLink := routes.Bingos.New()

	q_counties, err := app.DB.Prepare(Q_BINGO_COUNTIES)
	if err != nil && err.Error() != app.SQL_ERR_NO_ROWS {
		msg := `prepare counties query: `
		revel.AppLog.Errorf("%s%#v", msg, err)
		return c.RenderText(msg + err.Error())
	}
	q_countyHasHits, err := app.DB.Prepare(Q_COUNTY_HAS_HITS)
	if err != nil && err.Error() != app.SQL_ERR_NO_ROWS {
		msg := `prepare county-has-hits query: `
		revel.AppLog.Errorf("%s%#v", msg, err)
		return c.RenderText(msg + err.Error())
	}

	rows, err := app.DB.Query(Q_BINGO_LIST)
	if err != nil && err.Error() != app.SQL_ERR_NO_ROWS {
		msg := `get list of bingos: `
		revel.AppLog.Errorf("%s%#v", msg, err)
		return c.RenderText(msg + err.Error())
	} else if err != nil {
		// implies that no rows were found
		foundBingos = false
		return c.Render(newLink, foundBingos)
	}
	for rows.Next() {
		var (
			bName  string
			bId    int
			bCount int
		)
		err := rows.Scan(&bId, &bName, &bCount)
		if err != nil {
			msg := `read list of bingos: `
			revel.AppLog.Errorf("%s%#v", msg, err)
			return c.RenderText(msg + err.Error())
		}
		cRows, err := q_counties.Query(bId)
		if err != nil {
			msg := `query counties for bingo ` + bName + `: `
			revel.AppLog.Errorf("%s%#v", msg, err)
			return c.RenderText(msg + err.Error())
		}
		nCounties := 0
		nhCounties := 0
		for cRows.Next() {
			var (
				cId     int
				count   int
				discard string
			)
			err := cRows.Scan(&cId, &discard, &discard)
			if err != nil {
				msg := `read list of counties for bingo ` + bName + `: `
				revel.AppLog.Errorf("%s%#v", msg, err)
				return c.RenderText(msg + err.Error())
			}
			nCounties++
			err = q_countyHasHits.QueryRow(cId).Scan(&count)
			if err != nil {
				msg := fmt.Sprintf("query hits for county %d (bingo %s): ", cId, bName)
				revel.AppLog.Errorf("%s%#v", msg, err)
				return c.RenderText(msg + err.Error())
			}
			if count > 0 {
				nhCounties++
			}
		}
		bingoList = append(bingoList, app.BingoSummary{Id: bId, Label: bName, Count: nhCounties, Max: nCounties})
	}

	return c.Render(newLink, foundBingos, bingoList)
}

func addCounties(counties []string, bId int, tx *sql.Tx) string {
	q, err := tx.Prepare(`insert into bingo_counties (bingo_id, county_id) values ($1, $2)`)
	if err != nil {
		return fmt.Sprintf("prepare county-id insert: %#v", err)
	}

	for _, county := range counties {
		cId, err := strconv.Atoi(county)
		if err != nil {
			return fmt.Sprintf("convert county id '%s' to int: %#v", err)
		}
		res, err := q.Exec(bId, cId)
		if err != nil {
			return fmt.Sprintf("insert county %d into bingo %d: %#v", cId, bId, err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Sprintf("get nrows after insert county %d into bingo %d", cId, bId, err)
		}
		if n != 1 {
			return fmt.Sprintf("rows inserted (%d -> %d): %d", n)
		}
	}
	return ""
}

func (c Bingos) Create() revel.Result {
	var bId int

	//revel.AppLog.Debugf("%#v", c.Params.Form)
	title := c.Params.Form["title"][0]
	if len(title) < 1 {
		return c.RenderText("no title specified")
	}
	counties_in := app.RE_trailingCommas.ReplaceAllString(c.Params.Form["counties"][0], "")
	counties := s.Split(counties_in, ",")
	if len(counties) < 1 {
		return c.RenderText("no counties specified")
	}

	tx, err := app.DB.Begin()
	if err != nil {
		revel.AppLog.Errorf("begin transaction: %#v", err)
		tx.Rollback()
		return c.RenderText("begin transaction: ", err.Error())
	}

	res, err := tx.Exec(`insert into bingos (name) values ($1)`, title)
	if err != nil {
		revel.AppLog.Errorf("insert bingo name: %#v", err)
		tx.Rollback()
		return c.RenderText("unable to insert bingo name: ", err.Error())
	}
	n, err := res.RowsAffected()
	if err != nil {
		revel.AppLog.Errorf("get nrows after insert bingo name: %#v", err)
		tx.Rollback()
		return c.RenderText("get nrows after insert bingo name: ", err.Error())
	}
	if n != 1 {
		revel.AppLog.Errorf("rows inserted (bingo name): %d", n)
		tx.Rollback()
		return c.RenderText(fmt.Sprintf("rows inserted (bingo name): %d", n))
	}

	err = tx.QueryRow(`select id from bingos where name=$1`, title).Scan(&bId)
	if err != nil {
		revel.AppLog.Errorf("get new bingo id: %#v", err)
		tx.Rollback()
		return c.RenderText("get new bingo id: " + err.Error())
	}

	errStr := addCounties(counties, bId, tx)
	if errStr != "" {
		tx.Rollback()
		revel.AppLog.Error(errStr)
		return c.RenderText(errStr)
	}

	tx.Commit()

	return c.Redirect(routes.Bingos.Index())
}

func (c Bingos) New() revel.Result {
	postDest := routes.Bingos.Create()
	return c.Render(postDest)
}

func getBingoDetail(id int) ([]app.BingoDetail, string) {
	results := make([]app.BingoDetail, 0)
	q_hasHits, err := app.DB.Prepare(Q_COUNTY_HAS_HITS)
	if err != nil && err.Error() != app.SQL_ERR_NO_ROWS {
		return nil, fmt.Sprintf("prepare county-has-hits query: %#v", err)
	}

	rows, err := app.DB.Query(Q_BINGO_COUNTIES, id)
	if err != nil && err.Error() != app.SQL_ERR_NO_ROWS {
		return nil, fmt.Sprintf("query counties in bingo: %#v", err)
	}

	for rows.Next() {
		var (
			cId, nHits    int
			hitsFound     bool
			state, county string
		)
		err := rows.Scan(&cId, &state, &county)
		if err != nil {
			return nil, fmt.Sprintf("read county id: %#v", err)
		}
		err = q_hasHits.QueryRow(cId).Scan(&nHits)
		if err != nil {
			return nil, fmt.Sprintf("get hits for county %d: %#v", cId, err)
		}
		if nHits > 0 {
			hitsFound = true
		}
		results = append(results, app.BingoDetail{Id: cId, State: state, County: county, Hits: hitsFound})
	}
	return results, ""
}

func (c Bingos) Show() revel.Result {
	id, _ := strconv.Atoi(c.Params.Route.Get("id"))
	results, errStr := getBingoDetail(id)
	if results == nil {
		revel.AppLog.Error(errStr)
		results = make([]app.BingoDetail, 0)
	}
	return c.RenderJSON(results)
}

func (c Bingos) Update(id int) revel.Result {
	title := c.Params.Form["title"][0]
	counties_in := app.RE_trailingCommas.ReplaceAllString(c.Params.Form["counties"][0], "")
	revel.AppLog.Debugf("%#v", counties_in)
	counties := s.Split(counties_in, ",")
	if len(counties) < 1 {
		return c.RenderText("no counties specified")
	}
	tx, err := app.DB.Begin()
	if err != nil {
		msg := fmt.Sprintf("begin transaction: %#v", err)
		revel.AppLog.Error(msg)
		tx.Rollback()
		return c.RenderText(msg)
	}

	res, err := tx.Exec(`update bingos set name=$1 where id=$2`, title, id)
	if err != nil {
		msg := fmt.Sprintf("update bingo name: %#v", err)
		revel.AppLog.Error(msg)
		tx.Rollback()
		return c.RenderText(msg)
	}
	n, err := res.RowsAffected()
	if err != nil {
		msg := fmt.Sprintf("get nrows after update bingo name: %#v", err)
		revel.AppLog.Error(msg)
		tx.Rollback()
		return c.RenderText(msg)
	}
	if n != 1 {
		msg := fmt.Sprintf("rows update (bingo name): %d", n)
		revel.AppLog.Error(msg)
		tx.Rollback()
		return c.RenderText(msg)
	}

	_, err = tx.Exec(`delete from bingo_counties where bingo_id=$1`, id)
	if err != nil {
		msg := fmt.Sprintf("delete counties from bingo %d: %#v", id, err)
		revel.AppLog.Error(msg)
		tx.Rollback()
		return c.RenderText(msg)
	}
	errStr := addCounties(counties, id, tx)
	if errStr != "" {
		tx.Rollback()
		revel.AppLog.Error(errStr)
		return c.RenderText(errStr)
	}
	tx.Commit()
	return c.Redirect(routes.Bingos.Index())
}

func (c Bingos) Edit() revel.Result {
	var bTitle string

	id, _ := strconv.Atoi(c.Params.Route.Get("id"))
	postDest := routes.Bingos.Update(id)

	err := app.DB.QueryRow(`select name from bingos where id=$1`, id).Scan(&bTitle)
	if err != nil {
		msg := fmt.Sprintf("query/read bingo title: %#v", err)
		revel.AppLog.Error(msg)
		return c.RenderText(msg)
	}

	selectedCounties, errStr := getBingoDetail(id)
	if selectedCounties == nil {
		revel.AppLog.Error(errStr)
		return c.RenderText(errStr)
	}
	return c.Render(postDest, bTitle, selectedCounties)
}

// vim:foldmethod=marker:
