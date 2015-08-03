package controllers

import (
	"github.com/revel/revel"
)

type Category struct {
	GormController
}

func (c Category) New() revel.Result {
	return c.Render()
}
