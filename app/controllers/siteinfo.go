package controllers

import (
	"github.com/revel/revel"
	"obrolansubuh.com/models"
)

type SiteInfo struct {
	GormController
}

func (c SiteInfo) EditAboutUs() revel.Result {
	siteInfo := models.SiteInfo{}

	c.Trx.First(&siteInfo)

	return c.Render(siteInfo)
}
