package app

import (
	"database/sql"
	re "regexp"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/revel/modules/server-engine/newrelic"
	"github.com/revel/revel"
)

type Hit struct {
	Id         int
	Index      int
	Denom      int
	Serial     string
	Series     string
	RptKey     string
	EntDate    string
	Country    string
	State      string
	CountyCity string
	Count      int
}

type HitsBrkEnt struct {
	Region string
	Count  int
}

type DayOfMonth struct {
	Label  string
	Months [12]int
	Total  string
}

type Region struct {
	Id     int
	Region string
}

type BingoSummary struct {
	Id    int
	Label string
	Count int
	Max   int
}

type BingoDetail struct {
	Id     int
	State  string
	County string
	Hits   bool
}

type NewHitInfo struct {
	FirstOnDate      string
	FirstInCounty    string
	CountyBingoNames []string
	AdjacentCounties []string
}

type Bill struct {
	Id           int
	Serial       string
	Series       string
	Denomination int
	Rptkey       string
	Residence    string
	Message      string
}

const (
	SQL_ERR_NO_ROWS = `sql: no rows in result set`
	DATE_LAYOUT     = `2006-01-02`
	START_YEAR      = 2003
	START_MONTH     = time.November
	Q_HITS          = `
		select
			h.id,
			b.denomination,
			b.serial,
			b.series,
			b.rptkey,
			h.entdate,
			h.country,
			h.state,
			h.county,
			(select count(*) from hits where bill_id = b.id)
		from
			bills b,
			hits h
		where
			h.bill_id = b.id
	`

	Q_BINGOS = `
		select
			b.name
		from
			bingos b,
			bingo_counties bc,
			counties_master cm
		where
			cm.state=$1 and
			cm.county=$2 and
			bc.county_id=cm.id and
			b.id=bc.bingo_id
		order by b.name
	`

	Q_ADJACENT_COUNTIES = `
		select
			cm.state,
			cm.county
		from
			counties_master cm_in,
			counties_master cm,
			counties_graph cg
		where
			cm_in.state = $1 and
			cm_in.county = $2 and
			((cg.a=cm_in.id and cg.b=cm.id)
			 or (cg.b=cm_in.id and cg.a=cm.id))
	`
)

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string

	// my vars
	SeriesByLetter = map[string]string{
		"A": "1996",
		"B": "1999",
		"C": "2001",
		"D": "2003",
		"E": "2004",
		"F": "2003A",
		"G": "2004A",
		"H": "2006",
		"I": "2006",
		"J": "2009",
		"K": "2006A",
		"L": "2009A",
		"M": "2013",
		"N": "2017",
		"P": "2017A",
	}
)

var DB *sql.DB
var (
	RE_dbUnsafe           *re.Regexp
	RE_singleQuote        *re.Regexp
	RE_whitespace         *re.Regexp
	RE_leadingWhitespace  *re.Regexp
	RE_trailingWhitespace *re.Regexp
	RE_serial             *re.Regexp
	RE_trailingCommas     *re.Regexp
)

func InitDB() {
	//connstring := fmt.Sprintf("user=%s password='%s' dbname=%s sslmode=disable", "user", "pass", "database")
	connstring := revel.Config.StringDefault("db.connect", "")

	var err error
	DB, err = sql.Open(revel.Config.StringDefault("db.driver", ""), connstring)
	if err != nil {
		revel.AppLog.Info("DB Error", err)
	}
	revel.AppLog.Info("DB Connected")
}

func InitRE() {
	RE_dbUnsafe = re.MustCompile(`(;)`)
	RE_singleQuote = re.MustCompile(`'`)
	RE_whitespace = re.MustCompile(`\s+`)
	RE_leadingWhitespace = re.MustCompile(`^\s+`)
	RE_trailingWhitespace = re.MustCompile(`\s+$`)
	RE_serial = re.MustCompile(`^[A-NP-Y]?[A-L][0-9\-]{8}[A-NP-Y\*]$`)
	RE_trailingCommas = re.MustCompile(`,*$`)
}

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.BeforeAfterFilter,       // Call the before and after filter functions
		revel.ActionInvoker,           // Invoke the action.
	}

	// Register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	// revel.OnAppStart(ExampleStartupScript)
	revel.OnAppStart(InitDB)
	revel.OnAppStart(InitRE)
	// revel.OnAppStart(FillCache)
}

// HeaderFilter adds common security headers
// There is a full implementation of a CSRF filter in
// https://github.com/revel/modules/tree/master/csrf
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")
	c.Response.Out.Header().Add("Referrer-Policy", "strict-origin-when-cross-origin")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

//func ExampleStartupScript() {
//	// revel.DevMod and revel.RunMode work here
//	// Use this script to check for dev mode and set dev/prod startup scripts here!
//	if revel.DevMode == true {
//		// Dev mode
//	}
//}
