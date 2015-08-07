package controllers

import (
	"github.com/revel/revel"
	"obrolansubuh.com/backend/app/routes"
)

func checkUser(c *revel.Controller) revel.Result {
	if _, ok := c.Session["user"]; ok {
		return nil
	}

	c.Flash.Error(c.Message("login.message.notloggedin"))
	return c.Redirect(routes.App.Login())
}

func adminOnly(c *revel.Controller) revel.Result {
	if c.Session["usertype"] == "ADMIN" {
		return nil
	}

	c.Flash.Error(c.Message("access.message.notallowed"))
	return c.Redirect(routes.App.Index())
}

func init() {
	revel.OnAppStart(InitDB)
	revel.InterceptFunc(checkUser, revel.BEFORE, &Post{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &Profile{})
	revel.InterceptFunc(checkUser, revel.BEFORE, &Asset{})

	revel.InterceptFunc(adminOnly, revel.BEFORE, &Contributor{})
	revel.InterceptFunc(adminOnly, revel.BEFORE, &SiteInfo{})
	//revel.InterceptFunc(adminOnly, revel.BEFORE, &Category{})
	revel.InterceptMethod((*GormController).Begin, revel.BEFORE)
	revel.InterceptMethod((*GormController).Commit, revel.AFTER)
	revel.InterceptMethod((*GormController).RollBack, revel.FINALLY)

	revel.TemplateFuncs["config"] = func(key string) string {
		return revel.Config.StringDefault(key, "")
	}
}
