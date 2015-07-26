package controllers

import (
	"github.com/revel/revel"
	"net/http"
	"obrolansubuh.com/backend/app/routes"
	"obrolansubuh.com/models"
	"strconv"
)

type Contributor struct {
	GormController
}

func (c Contributor) New() revel.Result {
	var cTypes []models.ContributorType
	c.Trx.Find(&cTypes)

	return c.Render(cTypes)
}

func (c Contributor) Save(email, name, password, privilege string) revel.Result {
	c.Validation.Required(name).Message("Nama")
	c.Validation.Required(email).Message("Email kosong")
	c.Validation.Email(email).Message("Email tak valid")
	c.Validation.Required(password).Message("Password")
	c.Validation.Required(privilege).Message("Privilege")

	cType, err := strconv.ParseInt(privilege, 10, 64)

	if err != nil {
		return c.RenderText("ERROR")
	}

	if c.Validation.HasErrors() {
		messages := make([]string, 0, len(c.Validation.Errors))
		for _, v := range c.Validation.ErrorMap() {
			messages = append(messages, v.String())
		}

		FR := FailRequest{Messages: messages}

		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(FR)
	}

	contributor := models.Contributor{
		Name:     name,
		Email:    email,
		Password: password,
		TypeID:   cType,
	}

	c.Trx.Create(&contributor)

	if err := c.Trx.Error; err != nil {
		revel.ERROR.Printf("%s", err)

		return c.Redirect(routes.Contributor.New())
	}

	return c.Redirect(routes.Contributor.List())
}

func (c Contributor) List() revel.Result {
	return c.RenderText("Contributor List")
}
