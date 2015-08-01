package controllers

import (
	"github.com/revel/revel"
	"net/http"
	"obrolansubuh.com/models"
)

type SiteInfo struct {
	GormController
}

type AboutUsUpdated struct {
	Message string `json:"message"`
}

func (c SiteInfo) EditAboutUs() revel.Result {
	ToolbarItems := []ToolbarItem{
		ToolbarItem{Id: "update", Text: "Update", Icon: "editor:publish", Url: "SiteInfo.UpdateAboutUs"},
	}
	siteInfo := models.SiteInfo{}

	c.Trx.First(&siteInfo)

	return c.Render(siteInfo, ToolbarItems)
}

func (c SiteInfo) UpdateAboutUs(title, content string) revel.Result {
	c.Validation.Required(title).Message("JUNG JUNG PELET")
	c.Validation.Required(content).Message("TELEP GNUJ GNUJ")
	c.Validation.MaxSize(title, 1024).Message("MAMAKMU")

	if c.Validation.HasErrors() {
		messages := make([]string, 0, len(c.Validation.Errors))
		for _, v := range c.Validation.ErrorMap() {
			messages = append(messages, v.String())
		}

		FR := FailRequest{Messages: messages}

		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(FR)
	}

	var si models.SiteInfo
	c.Trx.First(&si)
	si.AboutUsTitle = title
	si.AboutUsContent = content
	c.Trx.Save(&si)

	if err := c.Trx.Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to save about us in database.")

		FR := FailRequest{Messages: []string{"AAA"}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	message := "WOOOO"
	AUU := AboutUsUpdated{
		Message: message,
	}

	return c.RenderJson(AUU)
}
