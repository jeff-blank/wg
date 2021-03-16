package controllers

import (
	"crypto/sha256"
	"fmt"

	"github.com/jeff-blank/wg/app"
	"github.com/revel/revel"
)

const (
	ERR_NO_USER       = "No username specified"
	ERR_NO_PASS       = "No password specified"
	ERR_BAD_USER_PASS = "You supplied an incorrect username and/or password."
	ERR_SYSTEM_ERROR  = "A system error occurred when trying to validate your credentials."
)

type Login struct {
	*revel.Controller
}

func (c Login) Index() revel.Result {
	backURL := c.Params.Get("back")
	return c.Render(backURL)
}

func (c Login) Login() revel.Result {
	var (
		errStr string
		user   string
		pass   string
		dbPass string
	)

	// check that form vars are present
	if _, ok := c.Params.Form["user"]; !ok {
		errStr = ERR_NO_USER
	} else if _, ok := c.Params.Form["passwd"]; !ok {
		errStr = ERR_NO_PASS
	}
	if errStr != "" {
		c.Flash.Error(errStr)
		return c.Redirect(App.Index)
	}

	// check that form vars are populated
	user = c.Params.Form["user"][0]
	pass = c.Params.Form["passwd"][0]
	if user == "" {
		errStr = ERR_NO_USER
	} else if pass == "" {
		errStr = ERR_NO_PASS
	}
	if errStr != "" {
		c.Flash.Error(errStr)
		return c.Redirect(App.Index)
	}

	pwHash := fmt.Sprintf("%064x", sha256.Sum256([]byte(pass)))

	err := app.DB.QueryRow(`select password from auth where username=$1`, user).Scan(&dbPass)
	if err != nil {
		if err.Error() == app.SQL_ERR_NO_ROWS {
			errStr = ERR_BAD_USER_PASS
		} else {
			revel.AppLog.Errorf("look up user '%s' in database: %#v", user, err)
			errStr = ERR_SYSTEM_ERROR
		}
	} else if pwHash != dbPass {
		errStr = ERR_BAD_USER_PASS
	}
	if errStr != "" {
		c.Flash.Error(errStr)
		return c.Redirect(App.Index)
	}

	c.Session.Set("username", user)

	backURL := c.Params.Get("back")
	if backURL == "" {
		return c.Redirect(App.Index)
	}
	revel.AppLog.Debugf("successful login for %s", user)
	return c.Redirect(backURL)
}

func (c Login) Logout() revel.Result {
	c.Session.Del("username")
	c.Flash.Success("You have been logged out.")
	return c.Redirect(Login.Index)
}

// vim:foldmethod=marker:
