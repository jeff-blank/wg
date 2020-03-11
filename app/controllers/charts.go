package controllers

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/util"
	"github.com/revel/revel"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

type Charts struct {
	*revel.Controller
}

func (c Charts) Grapher(mapName string) revel.Result {
	var (
		data    []map[string]interface{}
		series  []chart.Series
		yMaxR   float64
		yMaxL   float64
		yTicksR []chart.Tick
		//yTicksL []chart.Tick
	)

	data = util.StatsData("table").([]map[string]interface{})
	nMonths := len(data)

	if mapName == "score" {
		fomDate, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-01 0:00:00", app.START_YEAR, app.START_MONTH), time.Local)
		xVal := make([]time.Time, nMonths*2)
		xValScore := make([]time.Time, nMonths)
		yValHits := make([]float64, nMonths*2)
		yValScore := make([]float64, nMonths)
		yValEnts := make([]float64, nMonths*2)
		m := 0
		for _, d := range data {
			xVal[m*2] = fomDate
			xVal[m*2+1] = fomDate.AddDate(0, 1, 0).Add(-1 * time.Second)
			fomDate = fomDate.AddDate(0, 0, 1).AddDate(0, 1, 0).AddDate(0, 0, -1)
			xValScore[m] = xVal[m*2+1]
			yValScore[m] = d["score"].(float64)

			yValHits[m*2] = float64(d["monthHits"].(int))
			yValEnts[m*2] = float64(d["monthBills"].(int))
			yValHits[m*2+1] = float64(d["monthHits"].(int))
			yValEnts[m*2+1] = float64(d["monthBills"].(int))

			yMaxL = d["score"].(float64)
			yMaxR = math.Max(yMaxR, math.Max(yValHits[m*2], yValEnts[m*2]))

			m++
		}
		series = make([]chart.Series, 3)
		series = []chart.Series{
			chart.TimeSeries{
				XValues: xVal,
				YValues: yValEnts,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("f1c233"),
					FillColor:   drawing.ColorFromHex("f1c233").WithAlpha(128),
				},
			},
			chart.TimeSeries{
				XValues: xVal,
				YValues: yValHits,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("3366cc"),
					FillColor:   drawing.ColorFromHex("3366cc").WithAlpha(128),
				},
			},
			chart.TimeSeries{
				XValues: xValScore,
				YValues: yValScore,
				YAxis:   chart.YAxisSecondary,
			},
		}
		yMaxR = float64(int(yMaxR/100)*100 + 100)
		yMaxL = float64(int(yMaxL/100)*100 + 100)

		yTicksR = make([]chart.Tick, int(yMaxR/25)+1)
		v := 0
		for t, _ := range yTicksR {
			yTicksR[t] = chart.Tick{Label: fmt.Sprintf("%3d", v), Value: float64(v)}
			v += 25
		}
		revel.AppLog.Debugf("%#v", yTicksR)

		/*
			go-chart bug
			yTicksL = make([]chart.Tick, int(yMaxL/100)+1)
			v = 0
			for t, _ := range yTicksL {
				yTicksL[t] = chart.Tick{Label: fmt.Sprintf("%3d", v), Value: float64(v)}
				v += 100
			}
			revel.AppLog.Debugf("%#v", yTicksL)
		*/
	} else {
		return c.RenderText("unknown chart '" + mapName + "'")
	}

	graph := chart.Chart{
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: float64(int(yMaxR)),
			},
			Ticks: yTicksR,
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%4d", int(vf))
				}
				return ""
			},
		},
		YAxisSecondary: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: float64(int(yMaxL)),
			},
			//Ticks: yTicksL,
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
	err := graph.Render(chart.SVG, buffer)
	if err != nil {
		return c.RenderError(err)
	}
	return c.RenderBinary(buffer, "chart.svg", "", time.Now())
}

// vim:foldmethod=marker:
