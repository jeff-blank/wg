package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/jeff-blank/wg/app"
	"github.com/jeff-blank/wg/app/routes"
	"github.com/revel/revel"
)

const WG_LOGIN_URL = "https://www.wheresgeorge.com/logon.php"

//const WG_LOGIN_URL = "http://test2.mr-happy.com/tmp.php"

type Settings struct {
	*revel.Controller
}

func (c Settings) Index() revel.Result {
	var (
		userTZ    string
		tz        string
		tzDescr   string
		tzList    []app.TZRec
		prefs     app.UserPrefs
		wgCookies app.WGCreds
		userTZLoc *time.Location
	)

	rows, err := app.DB.Query(`select display_name, tz_name from tz order by display_name`)
	if err != nil {
		return c.RenderError(err)
	}
	for rows.Next() {
		err = rows.Scan(&tzDescr, &tz)
		if err != nil {
			return c.RenderError(err)
		}
		tzList = append(tzList, app.TZRec{TZDescr: tzDescr, TZString: tz})
	}
	prefs_a, err := getPrefs()
	if err != nil {
		return c.RenderError(err)
	}
	if prefs_a != nil {
		prefs = prefs_a[0]
	}

	if prefs.TZString != "" {
		userTZLoc, err = time.LoadLocation(prefs.TZString)
		if err != nil {
			return c.RenderError(err)
		}
	} else {
		userTZLoc = time.Local
	}

	wgCredsStatus := "Not logged in"
	if prefs.WGCreds != "" {
		err := json.Unmarshal([]byte(prefs.WGCreds), &wgCookies)
		if err != nil {
			return c.RenderError(err)
		}
		now := time.Now()
		wgCredsExpire := wgCookies.UserKey.Expires
		if now.After(wgCredsExpire) {
			wgCredsStatus = "Login session expired"
		} else {
			wgCredsStatus = "Logged in until " + wgCredsExpire.In(userTZLoc).Format(app.DATE_TIME_LAYOUT)
		}
	}
	wgLoginFormUrl := routes.Settings.WGLoginForm()
	userTZ = prefs.TZString
	//clearWgLoginUrl := routes.Settings.ClearWGLogin()
	clearWgLoginUrl := ""
	return c.Render(tzList, userTZ, wgCredsStatus, wgLoginFormUrl, clearWgLoginUrl)
}

func (c Settings) WGLoginForm() revel.Result {
	wgLoginUrl := routes.Settings.WGLogin()
	return c.Render(wgLoginUrl)
}

func (c Settings) WGLogin() revel.Result {
	var (
		execString string
		creds      app.WGCreds
	)

	email := c.Params.Get("email")
	pw := c.Params.Get("password")
	formData := url.Values{
		"email":    {email},
		"password": {pw},
		"duration": {"8760"},
		"go":       {"Log On"},
	}

	urlObj, err := url.Parse(WG_LOGIN_URL)
	if err != nil {
		return c.RenderError(err)
	}

	prefs, err := getPrefs()
	if err != nil {
		return c.RenderError(err)
	}

	jar, err := cookiejar.New(nil)

	browser := &http.Client{Jar: jar}
	req, err := http.NewRequest("GET", WG_LOGIN_URL, strings.NewReader(formData.Encode()))
	req.Header.Set("User-Agent", `Mozilla/5.0 (Windows NT 6.3; ) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36`)
	response, err := browser.Do(req)
	if err != nil {
		return c.RenderError(err)
	}
	foundMid := false
	for _, cookie := range jar.Cookies(urlObj) {
		if cookie.Name == "mid" {
			tmp := make([]*http.Cookie, 1)
			tmp[0] = cookie
			jar.SetCookies(urlObj, tmp)
			foundMid = true
			break
		}
	}
	if !foundMid {
		errMsg := "cookie 'mid' not found in response from " + WG_LOGIN_URL
		revel.AppLog.Error(errMsg)
		return c.RenderError(errors.New(errMsg))
	}
	//jar.SetCookies(urlObj, []*http.Cookie{{}}
	req, err = http.NewRequest("POST", WG_LOGIN_URL, strings.NewReader(formData.Encode()))
	if err != nil {
		return c.RenderError(err)
	}

	req.Header.Set("User-Agent", `Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Safari/537.36`)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", WG_LOGIN_URL)
	response, err = browser.Do(req)
	if err != nil {
		return c.RenderError(err)
	}
	cookies := response.Cookies()
	revel.AppLog.Debugf("%d cookies: %#v", len(cookies), cookies)
	revel.AppLog.Debugf("status: %s; headers: %#v", response.Status, response.Header)
	for _, cookie := range cookies {
		if cookie.Name == "mid" {
			creds.MID = *cookie
			revel.AppLog.Debugf("mid cookie: %#v", cookie)
		} else if cookie.Name == "userkey" {
			creds.UserKey = *cookie
			revel.AppLog.Debugf("userkey cookie: %#v", cookie)
		}
	}

	if creds.UserKey.Name == "" {
		errMsg := "no 'userkey' cookie returned"
		revel.AppLog.Error(errMsg)
		return c.RenderError(errors.New(errMsg))
	}

	if prefs != nil {
		revel.AppLog.Debugf("prefs found: (%d)%#v", len(prefs), prefs)
		execString = `update user_prefs set wg_site_creds=$1`
	} else {
		revel.AppLog.Debug("no prefs found")
		execString = `insert into user_prefs (wg_site_creds) values ($1)`
	}
	cookieStr, err := json.Marshal(creds)
	if err != nil {
		return c.RenderError(err)
	}
	res, err := app.DB.Exec(execString, cookieStr)
	if err != nil {
		return c.RenderError(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return c.RenderError(err)
	}
	if rows != 1 {
		errMsg := fmt.Sprintf("wg_site_creds: updated %d rows", rows)
		revel.AppLog.Error(errMsg)
		return c.RenderText("error:" + " " + errMsg)
	}
	return c.Redirect(routes.Settings.Index())
}

func getPrefs() ([]app.UserPrefs, error) {
	var tz, creds string

	rows, err := app.DB.Query(`select coalesce(tz, '') as tz, coalesce(wg_site_creds, '') from user_prefs`)
	if err != nil {
		if err.Error() != app.SQL_ERR_NO_ROWS {
			return nil, err
		} else {
			// no rows found
			revel.AppLog.Debug("getPrefs(): no rows selected")
			return nil, nil
		}
	}
	i := 0
	for rows.Next() {
		i++
		if i > 1 {
			errMsg := "getPrefs(): found at least 2 rows"
			revel.AppLog.Error(errMsg)
			return nil, errors.New(errMsg)
		}
		err = rows.Scan(&tz, &creds)
		if err != nil {
			return nil, err
		}
	}
	revel.AppLog.Debugf("getPrefs(): found %d rows", i)
	if i == 0 {
		revel.AppLog.Debug("getPrefs(): no rows found")
		return nil, nil
	}
	prefs := make([]app.UserPrefs, 1)
	prefs[0] = app.UserPrefs{TZString: tz, WGCreds: creds}
	return prefs, nil
}
