package controllers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/routes"
	"github.com/jeff-blank/wg/app/util"
	"github.com/revel/revel"
)

type Bills struct {
	*revel.Controller
}

const SQL_BILL_BY_ID = `select id, serial, series, denomination, rptkey, residence from bills where id=$1`

func (c Bills) Edit() revel.Result {
	id := c.Params.Route.Get("id")
	bill, err := getBillById(id)
	if err != nil {
		revel.AppLog.Error("bills.Edit(): get bill with id '%s': %#v", id, err)
		return c.RenderError(err)
	}
	residences, err := util.GetResidences()
	if err != nil {
		return c.RenderError(err)
	}
	currentResidence, err := util.GetCurrentResidence()
	if err != nil {
		return c.RenderError(err)
	}
	hitList, err := util.GetHits("and h.bill_id=" + id + " order by entdate desc, b.id desc")
	return c.Render(bill, residences, currentResidence, hitList)
}

func (c Bills) Update() revel.Result {
	billId := c.Params.Route.Get("id")
	res, err := app.DB.Exec(
		`update bills set rptkey=$1, serial=$2, denomination=$3, series=$4, residence=$5 where id=$6`,
		c.Params.Get("key"),
		c.Params.Get("serial"),
		c.Params.Get("denom"),
		c.Params.Get("series"),
		c.Params.Get("residence"),
		billId,
	)
	if err != nil {
		revel.AppLog.Errorf("update bill with id %s: %#v", billId, err)
		return c.RenderError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return c.RenderError(err)
	}
	if n != 1 {
		err := errors.New(fmt.Sprintf("update affected %d rows (expected 1)", n))
		revel.AppLog.Error(err.Error())
		return c.RenderError(err)
	}

	updateFlash := app.HitInfo{GenericMessage: "Bill information successfully updated"}
	flashJson, err := json.Marshal(updateFlash)
	if err == nil {
		c.Flash.Out["info"] = string(flashJson)
	}

	return c.Redirect(routes.Hits.Index() + "?year=current")
}

func getBillById(id string) (app.Bill, error) {
	var bill app.Bill

	err := app.DB.QueryRow(SQL_BILL_BY_ID, id).Scan(
		&bill.Id,
		&bill.Serial,
		&bill.Series,
		&bill.Denomination,
		&bill.Rptkey,
		&bill.Residence,
	)
	if err != nil {
		if err.Error() == app.SQL_ERR_NO_ROWS {
			return app.Bill{}, nil
		}
		return app.Bill{}, err
	}
	return bill, nil
}
