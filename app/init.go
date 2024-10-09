package app

import (
	"database/sql"
	"io/ioutil"
	re "regexp"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/revel/modules/server-engine/newrelic"
	"github.com/revel/revel"
	"gopkg.in/yaml.v2"
)

type Hit struct {
	Id        int
	BillId    int
	Index     int
	Denom     int
	Serial    string
	Series    string
	RptKey    string
	EntDate   string
	Residence string
	Country   string
	State     string
	County    string
	City      string
	ZIP       string
	Count     int
	WGHitNum  int
}

type HitsBrkEnt struct {
	Region string
	Count  int
}

type DualRegionBrkEnt struct {
	Rank    int
	RankStr string
	State   string
	County  string
	Count   int
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

type USCounty struct {
	Id     int
	State  string
	County string
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

type HARFillerSet struct {
	Series            string
	Denom             int
	FRB               string
	BlockLetter       string
	AnyFirst          bool
	SeriesDenom       bool
	DenomFRB          bool
	SeriesFRB         bool
	FRBBlockLetter    bool
	SeriesBlockLetter bool
}

type HitInfo struct {
	FirstOnDate      string
	FirstInCounty    string
	CountyBingoNames []string
	AdjacentCounties []string
	HARFillers       HARFillerSet
	GenericMessage   string
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

type DateSelData struct {
	Year   int
	Month  string
	Day    string
	Years  []int
	Months [12]string
	Days   []string
}

type TZRec struct {
	TZDescr  string
	TZString string
}

type UserPrefs struct {
	TZString     string
	WGProfileKey string
}

type SimpleCount struct {
	Label string
	Count int
}

const (
	SQL_ERR_NO_ROWS  = `sql: no rows in result set`
	DATE_LAYOUT      = `2006-01-02`
	DATE_TIME_LAYOUT = `2006-01-02 15:04:05 MST`
	START_YEAR       = 2003
	START_MONTH      = time.November

	// SQL queries {{{
	Q_HITS = `
		select
			h.id,
			b.id,
			b.denomination,
			b.serial,
			b.series,
			b.rptkey,
			b.residence,
			h.entdate,
			h.country,
			coalesce(cm.state, '--') as state,
			coalesce(cm.county, '--') as county,
			coalesce(h.city, '') as city,
			coalesce(h.zip, '') as zip,
			(select count(*) from hits where bill_id = b.id),
			coalesce(h.wg_hit_number, -1)
		from
			bills b,
			hits h
		left outer join
			counties_master cm
		on
			h.county_id = cm.id
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

	Q_LAST_50_HIT_IDS = `select id from hits order by entdate desc, id desc limit 50`

	Q_LAST_50_DENOM_SERIES = `
		select
			b.%s,
			count(1)
		from
			bills b,
			hits h
		where
			b.id=h.bill_id and h.id in
				(` + Q_LAST_50_HIT_IDS + `)
		group by
			b.%s
		order by
			b.%s
	`

	Q_LAST_50_STATES = `
		select cm.state, count(1)
		from hits h, counties_master cm
		where h.id in (` + Q_LAST_50_HIT_IDS + `) and cm.id = h.county_id and country = 'US'
		group by cm.state
		union
		  select state, count(1)
		  from hits
		  where id in (` + Q_LAST_50_HIT_IDS + `) and country = 'Canada'
		  group by state
		  union
		    select country, count(1)
		    from hits
		    where id in (` + Q_LAST_50_HIT_IDS + `) and country not in ('US', 'Canada')
		    group by country
		order by count desc, state asc
	`

	Q_LAST_50_COUNTIES = `
		select cm.state, cm.county, count(1)
		from hits h, counties_master cm
		where h.id in (` + Q_LAST_50_HIT_IDS + `) and cm.id = h.county_id and country = 'US'
		group by cm.state, cm.county
		union
		  select state, city, count(1)
		  from hits
		  where id in (` + Q_LAST_50_HIT_IDS + `) and country = 'Canada'
		  group by state, city
		  union
		    select country, city, count(1)
		    from hits
		    where id in (` + Q_LAST_50_HIT_IDS + `) and country not in ('US', 'Canada')
		    group by country, city
		order by count desc, state asc, county asc
	`

	Q_HITS_CALENDAR = `
		select
			substr(entdate::text, 6) as date,
			count(1)
		from
			hits
		group by
			date
	`

	// }}}

	SERIAL_RE_BASE = `[A-L][0-9\-]{8}[A-NP-Y\*]$`
)

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string

	// my vars
	SeriesByLetter map[string]string
)

var DB *sql.DB
var Environment string
var (
	RE_dbUnsafe           *re.Regexp
	RE_singleQuote        *re.Regexp
	RE_whitespace         *re.Regexp
	RE_leadingWhitespace  *re.Regexp
	RE_trailingWhitespace *re.Regexp
	RE_serial             *re.Regexp
	RE_serial_10          *re.Regexp
	RE_serial_11          *re.Regexp
	RE_series             *re.Regexp
	RE_trailingCommas     *re.Regexp
	RE_date               *re.Regexp
	RE_nonNumeric         *re.Regexp
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
	RE_serial = re.MustCompile(`^[A-NP-Y]?` + SERIAL_RE_BASE)
	RE_serial_10 = re.MustCompile(`^` + SERIAL_RE_BASE)
	RE_serial_11 = re.MustCompile(`^[A-NP-Y]` + SERIAL_RE_BASE)
	RE_series = re.MustCompile(`^(19|2[0-3])[0-9]{2}[A-NP-Y]?$`)
	RE_trailingCommas = re.MustCompile(`,*$`)
	RE_date = re.MustCompile(`^\d{4}-(0\d|1[012])-([012]\d|3[01])$`)
	RE_nonNumeric = re.MustCompile(`\D+`)
}

func InitSeries() {
	yamlcfg, err := ioutil.ReadFile("data/seriesMap.yaml")
	if err != nil {
		revel.AppLog.Fatalf("read config file 'data/seriesMap.yaml': %v", err)
	}

	if err := yaml.Unmarshal(yamlcfg, &SeriesByLetter); err != nil {
		revel.AppLog.Fatal("yaml.Unmarshal(): ", err)
	}
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
	revel.OnAppStart(InitSeries)
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

// vim:foldmethod=marker:
