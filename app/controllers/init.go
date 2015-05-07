package controllers

import "github.com/revel/revel"

func init() {
	revel.OnAppStart(InitDB)
	revel.InterceptMethod((*DBRController).Begin, revel.BEFORE)
	revel.InterceptMethod((*DBRController).Commit, revel.AFTER)
	revel.InterceptMethod((*DBRController).RollBack, revel.FINALLY)
}
