package controllers

import (
	"fmt"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"obrolansubuh.com/backend/app/routes"
	"time"
)

type App struct {
	DBRController
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Login() revel.Result {
	return c.Render()
}

func (c App) ProcessLogin(email, password string, remember bool) revel.Result {
	contributor, uerr := c.GetContributor(email)

	if uerr != nil {
		revel.INFO.Printf("[LGINFO] Login as %s failed. No email in database.", email)
		c.Flash.Error(c.Message("errors.login.email"))
	} else {
		err := bcrypt.CompareHashAndPassword([]byte(contributor.Password), []byte(password))
		if err == nil {
			c.Session["user"] = contributor.Email

			if remember {
				c.Session.SetDefaultExpiration()
			} else {
				c.Session.SetNoExpiration()
			}

			loginTime := time.Now().Local().Format(revel.Config.StringDefault("format.datetime", "02 Jan 2006 15:04"))
			revel.INFO.Printf("[LGINFO] Contributor %s logged in at %s.", email, loginTime)
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
