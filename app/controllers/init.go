package controllers

import (
	"net/url"

	"github.com/jeff-blank/wg/app/routes"
	"github.com/revel/revel"
)

func Authenticate(c *revel.Controller) revel.Result {
	var passedQueryString string

	username, err := c.Session.Get("username")
	if err != nil {
		if err.Error() == "Session value not found" {
			revel.AppLog.Debug("no username")
			c.Flash.Error("You must be logged in to continue")
		} else {
			revel.AppLog.Debugf("username error: %#v", err)
			c.Flash.Error(err.Error())
		}
		if len(c.Params.Values) > 0 {
			for param := range c.Params.Values {
				passedQueryString += "&" + param + "=" + c.Params.Get(param)
			}
			passedQueryString = url.QueryEscape("?" + passedQueryString[1:])
		}
		return c.Redirect(routes.Login.Index() + "?back=" + c.Request.URL.Path + passedQueryString)
	} else {
		revel.AppLog.Debugf("username: %#v", username)
	}
	return nil
}

func init() {
	revel.InterceptFunc(Authenticate, revel.BEFORE, &Bingos{})
	revel.InterceptFunc(Authenticate, revel.BEFORE, &Entries{})
	revel.InterceptFunc(Authenticate, revel.BEFORE, &Hits{})
	revel.InterceptFunc(Authenticate, revel.BEFORE, &Reports{})
	revel.InterceptFunc(Authenticate, revel.BEFORE, &Charts{})
	revel.InterceptFunc(Authenticate, revel.BEFORE, &Util{})
	revel.InterceptFunc(Authenticate, revel.BEFORE, &App{})
}
