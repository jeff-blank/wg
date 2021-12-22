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
		data        []map[string]interface{}
		series      []chart.Series
		yMaxR       float64
		yMaxL       float64
		yTicksR     []chart.Tick
		yTickSpaceR int
		//yTicksL []chart.Tick
	)

	data = util.StatsData("table").([]map[string]interface{})
	nMonths := len(data)
	xValStepped := make([]time.Time, nMonths*2)
	xValLine := make([]time.Time, nMonths)
	fomDate, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-01 0:00:00", app.START_YEAR, app.START_MONTH), time.Local)

	if mapName == "score" {
		yValHits := make([]float64, nMonths*2)
		yValScore := make([]float64, nMonths)
		yValEnts := make([]float64, nMonths*2)
		m := 0
		for _, d := range data {
			xValStepped[m*2] = fomDate
			xValStepped[m*2+1] = fomDate.AddDate(0, 1, 0).Add(-1 * time.Second)
			fomDate = fomDate.AddDate(0, 0, 1).AddDate(0, 1, 0).AddDate(0, 0, -1)
			xValLine[m] = xValStepped[m*2+1]
			yValScore[m] = d["score"].(float64)

			yValHits[m*2] = float64(d["monthHits"].(int))
			yValEnts[m*2] = float64(d["monthBills"].(int))
			yValHits[m*2+1] = yValHits[m*2]
			yValEnts[m*2+1] = yValEnts[m*2]

			yMaxL = yValScore[m]
			yMaxR = math.Max(yMaxR, math.Max(yValHits[m*2], yValEnts[m*2]))

			m++
		}
		series = make([]chart.Series, 3)
		series = []chart.Series{
			chart.TimeSeries{
				XValues: xValStepped,
				YValues: yValEnts,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("f1c233"),
					FillColor:   drawing.ColorFromHex("f1c233").WithAlpha(128),
				},
			},
			chart.TimeSeries{
				XValues: xValStepped,
				YValues: yValHits,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("3366cc"),
					FillColor:   drawing.ColorFromHex("3366cc").WithAlpha(128),
				},
			},
			chart.TimeSeries{
				XValues: xValLine,
				YValues: yValScore,
				YAxis:   chart.YAxisSecondary,
				Style: chart.Style{
					StrokeWidth: 2,
				},
			},
		}
		yMaxR = float64(int(yMaxR/100)*100 + 100)
		yMaxL = float64(int(yMaxL/100)*100 + 100)

		yTicksR = make([]chart.Tick, int(yMaxR/25)+1)
		yTickSpaceR = 25

	} else if mapName == "hits" {
		yValHits := make([]float64, nMonths*2)
		yValAvgLine := make([]float64, nMonths)
		yValAvgMonth := make([]float64, nMonths)
		yValCumulative := make([]float64, nMonths)
		yValMonthly := make([]float64, nMonths)
		m := 0
		for _, d := range data {
			xValStepped[m*2] = fomDate
			xValStepped[m*2+1] = fomDate.AddDate(0, 1, 0).Add(-1 * time.Second)
			fomDate = fomDate.AddDate(0, 0, 1).AddDate(0, 1, 0).AddDate(0, 0, -1)
			xValLine[m] = xValStepped[m*2+1]
			yValAvgLine[m] = d["straightLineAvgHits"].(float64)
			yValAvgMonth[m] = d["oneYrAvgMonthlyHits"].(float64)
			yValCumulative[m] = float64(d["cumulativeHits"].(int))
			yValMonthly[m] = d["avgMonthlyHits"].(float64)

			yValHits[m*2] = float64(d["monthHits"].(int))
			yValHits[m*2+1] = yValHits[m*2]

			yMaxR = yValAvgLine[m]
			yMaxL = math.Max(yMaxL, yValHits[m*2])

			m++
		}

		series = make([]chart.Series, 5)
		series = []chart.Series{
			chart.TimeSeries{
				XValues: xValStepped,
				YValues: yValHits,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("3366cc"),
					FillColor:   drawing.ColorFromHex("3366cc").WithAlpha(128),
				},
				YAxis: chart.YAxisSecondary,
			},
			chart.TimeSeries{
				XValues: xValLine,
				YValues: yValCumulative,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("db4537"),
					StrokeWidth: 2,
				},
			},
			chart.TimeSeries{
				XValues: xValLine,
				YValues: yValAvgLine,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("f5b401"),
					StrokeWidth: 2,
				},
			},
			chart.TimeSeries{
				XValues: xValLine,
				YValues: yValMonthly,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("ff00ff"),
					StrokeWidth: 2,
				},
				YAxis: chart.YAxisSecondary,
			},
			chart.TimeSeries{
				XValues: xValLine,
				YValues: yValAvgMonth,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("000000"),
					StrokeWidth: 2,
				},
				YAxis: chart.YAxisSecondary,
			},
		}
		yMaxR = float64(int(yMaxR/200)*200 + 200)
		yMaxL = float64(int(yMaxL/10)*10 + 10)

		yTickSpaceR = 200
	} else if mapName == "bills" {
		yValBills := make([]float64, nMonths*2)
		yValAvgBills := make([]float64, nMonths)
		yValAvgHits := make([]float64, nMonths)
		m := 0
		for _, d := range data {
			xValStepped[m*2] = fomDate
			xValStepped[m*2+1] = fomDate.AddDate(0, 1, 0).Add(-1 * time.Second)
			fomDate = fomDate.AddDate(0, 0, 1).AddDate(0, 1, 0).AddDate(0, 0, -1)
			xValLine[m] = xValStepped[m*2+1]
			yValAvgBills[m] = d["oneYrAvgMonthlyBills"].(float64)
			yValAvgHits[m] = d["oneYrAvgMonthlyHits"].(float64)

			yValBills[m*2] = float64(d["monthBills"].(int))
			yValBills[m*2+1] = yValBills[m*2]

			m++
		}

		series = make([]chart.Series, 3)
		series = []chart.Series{
			chart.TimeSeries{
				XValues: xValStepped,
				YValues: yValBills,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("3366cc"),
					FillColor:   drawing.ColorFromHex("3366cc").WithAlpha(128),
				},
				YAxis: chart.YAxisSecondary,
			},
			chart.TimeSeries{
				XValues: xValLine,
				YValues: yValAvgBills,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("274e13"),
					StrokeWidth: 2,
				},
				YAxis: chart.YAxisSecondary,
			},
			chart.TimeSeries{
				XValues: xValLine,
				YValues: yValAvgHits,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("ff00ff"),
					StrokeWidth: 2,
				},
			},
		}
		yMaxR = 40
		yMaxL = 400
		yTickSpaceR = 7.0

	} else {
		return c.RenderText("unknown chart '" + mapName + "'")
	}
	yTicksR = make([]chart.Tick, int(yMaxR/float64(yTickSpaceR))+1)
	v := 0
	for t, _ := range yTicksR {
		yTicksR[t] = chart.Tick{Label: fmt.Sprintf("%2d", v), Value: float64(v)}
		v += yTickSpaceR
	}

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
			/*
				ValueFormatter: func(v interface{}) string {
					if vf, isFloat := v.(float64); isFloat {
						return fmt.Sprintf("%4d", int(vf))
					}
					return ""
				},
			*/
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
