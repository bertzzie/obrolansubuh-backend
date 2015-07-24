package controllers

import (
	"github.com/revel/revel"
	//"obrolansubuh.com/models"
)

type Contributor struct {
	GormController
}

func (c Contributor) EditProfile() revel.Result {
	return c.Render()
}
