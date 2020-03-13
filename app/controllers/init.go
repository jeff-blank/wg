package controllers

import (
	"github.com/jeff-blank/wg/app/routes"
	"github.com/revel/revel"
)

func Authenticate(c *revel.Controller) revel.Result {
	username, err := c.Session.Get("username")
	if err != nil {
		if err.Error() == "Session value not found" {
			revel.AppLog.Debugf("no username")
			c.Flash.Error("You are not logged in")
		} else {
			revel.AppLog.Debugf("username error: %#v", err)
			c.Flash.Error(err.Error())
		}
		return c.Redirect(routes.Login.Index() + "?back=" + c.Request.URL.Path)
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
	//revel.InterceptFunc(Authenticate, revel.BEFORE, &Charts{})
	//revel.InterceptFunc(Authenticate, revel.BEFORE, &Util{})
	revel.InterceptFunc(Authenticate, revel.BEFORE, &App{})
}
