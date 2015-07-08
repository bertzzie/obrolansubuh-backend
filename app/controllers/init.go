package controllers

import "github.com/revel/revel"

func init() {
	revel.OnAppStart(InitDB)
	revel.InterceptMethod((*GormController).Begin, revel.BEFORE)
	revel.InterceptMethod((*GormController).Commit, revel.AFTER)
	revel.InterceptMethod((*GormController).RollBack, revel.FINALLY)

	revel.TemplateFuncs["config"] = func(key string) string {
		return revel.Config.StringDefault(key, "")
	}
}
