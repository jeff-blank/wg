package controllers

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/util"
	"github.com/revel/revel"
	"github.com/wcharczuk/go-chart"
)

type Charts struct {
	*revel.Controller
}

func (c Charts) Grapher(mapName string) revel.Result {
	var data []map[string]interface{}
	var series []chart.Series

	data = util.StatsData("table").([]map[string]interface{})
	nMonths := len(data)

	if mapName == "score" {
		eomDate, _ := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-01 12:00:00", app.START_YEAR, app.START_MONTH))
		eomDate = eomDate.AddDate(0, 1, 0).AddDate(0, 0, -1)
		xVal := make([]time.Time, nMonths+1)
		xVal[0] = eomDate.Add(-1 * time.Second)
		yValHits := make([]float64, nMonths+1)
		yValScore := make([]float64, nMonths+1)
		yValEnts := make([]float64, nMonths+1)
		for m, d := range data {
			xVal[m+1] = eomDate
			eomDate = eomDate.AddDate(0, 0, 1).AddDate(0, 1, 0).AddDate(0, 0, -1)
			yValHits[m+1] = float64(d["monthHits"].(int))
			yValScore[m+1] = d["score"].(float64)
			yValEnts[m+1] = float64(d["monthBills"].(int))
		}
		series = make([]chart.Series, 3)
		series = []chart.Series{
			chart.TimeSeries{
				XValues: xVal,
				YValues: yValHits,
			},
			chart.TimeSeries{
				XValues: xVal,
				YValues: yValEnts,
			},
			chart.TimeSeries{
				XValues: xVal,
				YValues: yValScore,
				YAxis:   chart.YAxisSecondary,
			},
		}
	} else {
		return c.RenderText("unknown chart '" + mapName + "'")
	}

	graph := chart.Chart{
		YAxis: chart.YAxis{
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%4d", int(vf))
				}
				return ""
			},
		},
		YAxisSecondary: chart.YAxis{
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%4d", int(vf))
				}
				return ""
			},
		},
		Series: series,
	}
	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		return c.RenderError(err)
	}
	return c.RenderBinary(buffer, "chart.png", "", time.Now())
}

// vim:foldmethod=marker:
