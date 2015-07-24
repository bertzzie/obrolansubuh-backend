package controllers

import (
	"fmt"
	"github.com/revel/revel"
	"obrolansubuh.com/models"
)

type Contributor struct {
	GormController
}

func (c Contributor) EditProfile(id int64) revel.Result {
	contributor := &models.Contributor{}
	c.Trx.Where("id = ?", id).First(&contributor)

	if err := c.Trx.Error; err != nil {
		fmt.Println("ERROR")
		return c.RenderText("ERROR")
	}

	return c.Render(contributor)
}
