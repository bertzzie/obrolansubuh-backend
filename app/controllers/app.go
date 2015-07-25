package controllers

import (
	"fmt"
	"github.com/revel/revel"
	"obrolansubuh.com/backend/app/routes"
	"strconv"
	"time"
)

type App struct {
	GormController
}

type ToolbarItem struct {
	Id       string
	Text     string
	Icon     string
	Url      string
	UrlParam string
}

type Action struct {
	Rel string
	Uri string
}

type JsonResponse struct {
	Actions []string
}

// JSON request type for 400
type FailRequest struct {
	// we want to support sending many errors in 1 response
	// for i.e. form validation
	Messages []string `json:"messages"`
}

func (c App) Index() revel.Result {
	if checkUser(c.GormController.Controller) != nil {
		return c.Redirect(routes.App.Login())
	}

	return c.Render()
}

func (c App) Login() revel.Result {
	if checkUser(c.GormController.Controller) == nil {
		c.Flash.Error(c.Message("login.message.alreadyli"))
		return c.Redirect(routes.App.Index())
	}

	return c.Render()
}

func (c App) ProcessLogin(email, password string, remember bool) revel.Result {
	contributor, uerr := c.GetContributor(email)

	if uerr != nil {
		revel.INFO.Printf("[LGINFO] Login as %s failed. No email in database.", email)
		c.Flash.Error(c.Message("errors.login.email"))
	} else {
		if contributor.CheckPassword(password) {
			priv := contributor.Type.Type

			c.Session["userid"] = strconv.FormatInt(contributor.ID, 10)
			c.Session["user"] = contributor.Email
			c.Session["username"] = contributor.Name
			c.Session["userphoto"] = contributor.Photo
			c.Session["usertype"] = priv

			if remember {
				c.Session.SetDefaultExpiration()
			} else {
				c.Session.SetNoExpiration()
			}

			loginTime := time.Now().Local().Format(revel.Config.StringDefault("format.datetime", "02 Jan 2006 15:04"))
			revel.INFO.Printf("[LGINFO] Contributor %s (Privilege: %s) logged in at %s.", email, priv, loginTime)
			c.Flash.Success(fmt.Sprintf(c.Message("login.message.success"), contributor.Name))

			return c.Redirect(routes.App.Index())
		} else {
			revel.INFO.Printf("[LGINFO] Login as %s failed. Wrong password.", email)
			c.Flash.Error(c.Message("errors.login.passw"))
		}
	}

	c.Flash.Out["username"] = email
	return c.Redirect(routes.App.Login())
}

func (c App) Logout() revel.Result {
	email, _ := c.Session["user"]

	for k := range c.Session {
		delete(c.Session, k)
	}

	logoutTime := time.Now().Local().Format(revel.Config.StringDefault("format.datetime", "02 Jan 2006 15:04"))
	revel.INFO.Printf("[LGINFO] Contributor %s logged out at %s.", email, logoutTime)
	c.Flash.Success(c.Message("logout.message.success"))

	return c.Redirect(routes.App.Login())
}
