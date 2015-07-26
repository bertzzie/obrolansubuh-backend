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

type ContributorList struct {
	ID       int64
	Name     string
	Email    string
	Photo    string
	EditLink string
}

func (c Contributor) JsonList() revel.Result {
	var contributors []models.Contributor
	if err := c.Trx.Find(&contributors).Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to get user list from database. Error: %s", err)

		FR := FailRequest{Messages: []string{c.Message("errors.db.generic")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	contributorList := make([]ContributorList, 0, len(contributors))
	for _, contributor := range contributors {
		tmp := ContributorList{
			ID:       contributor.ID,
			Name:     contributor.Name,
			Email:    contributor.Email,
			Photo:    contributor.Photo,
			EditLink: routes.Contributor.Edit(contributor.ID),
		}
		contributorList = append(contributorList, tmp)
	}

	return c.RenderJson(contributorList)
}

func (c Contributor) List() revel.Result {
	return c.Render()
}

func (c Contributor) New() revel.Result {
	var cTypes []models.ContributorType
	c.Trx.Find(&cTypes)

	return c.Render(cTypes)
}

func (c Contributor) Save(email, name, password, privilege string) revel.Result {
	c.Validation.Required(name).Message(c.Message("contributor.validation.name"))
	c.Validation.Required(email).Message("contributor.validation.email.required")
	c.Validation.Email(email).Message("contributor.validation.email.invalid")
	c.Validation.Required(password).Message("contributor.validation.password")
	c.Validation.Required(privilege).Message("contributor.validation.privilege")

	cType, err := strconv.ParseInt(privilege, 10, 64)

	if err != nil {
		revel.ERROR.Printf("[LGWRD] Privilege parse error. Value: %s, should be valid int64 value. Parser Error: %s",
			privilege,
			err,
		)

		c.Flash.Error(c.Message("contributor.privilege.parseerror"))
		return c.Redirect(routes.Contributor.New())
	}

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.Contributor.New())
	}

	// existing contributor check
	_, dupe := c.GetContributor(email)
	if dupe == nil {
		c.Flash.Error(c.Message("contributor.validation.email.duplicate"))
		return c.Redirect(routes.Contributor.New())
	}

	contributor := models.Contributor{
		Name:   name,
		Email:  email,
		TypeID: cType,
		Photo:  "/public/img/default-user.png",
		About:  "A contributor of ObrolanSubuh.com",
	}
	contributor.SetPassword(password)

	if err = c.Trx.Create(&contributor).Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Error in creating new contributor %s. Error: %s",
			email,
			err,
		)

		// Gorm appears to have no concept of error code.
		// We'll have to transfer the direct SQL Error so user
		// can have more clue of what's happening in case of
		// duplicate email error leaking here.
		c.Flash.Error(err.Error())
		return c.Redirect(routes.Contributor.New())
	}

	return c.Redirect(routes.Contributor.List())
}

func (c Contributor) Edit(id int64) revel.Result {
	var contributor models.Contributor
	var cTypes []models.ContributorType
	c.Trx.Find(&cTypes)

	if err := c.Trx.Preload("Type").Where("id = ?", id).First(&contributor).Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to get Contributor ID %d data. Error: %s",
			id,
			err,
		)

		c.Flash.Error(c.Message("errors.db.generic"))
		return c.Redirect(routes.App.Index())
	}

	return c.Render(contributor, cTypes)
}

func (c Contributor) Update(id, name, email, password, privilege string) revel.Result {
	c.Validation.Required(name).Message(c.Message("contributor.validation.name"))
	c.Validation.Required(email).Message("contributor.validation.email.required")
	c.Validation.Email(email).Message("contributor.validation.email.invalid")
	c.Validation.Required(password).Message("contributor.validation.password")
	c.Validation.Required(privilege).Message("contributor.validation.privilege")

	cType, err := strconv.ParseInt(privilege, 10, 64)
	realID, iderr := strconv.ParseInt(id, 10, 64)

	if err != nil || iderr != nil {
		revel.ERROR.Printf("[LGWRD] Privilege parse error. Value: %s, should be valid int64 value. Parser Error: %s",
			privilege,
			err,
		)

		c.Flash.Error(c.Message("contributor.privilege.parseerror"))
		return c.Redirect(routes.Contributor.Edit(realID))
	}

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.Contributor.Edit(realID))
	}

	var contributor models.Contributor
	if err = c.Trx.Preload("Type").Where("id = ?", id).First(&contributor).Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to get Contributor ID %d data. Error: %s",
			id,
			err,
		)

		c.Flash.Error(c.Message("errors.db.generic"))
		return c.Redirect(routes.Contributor.Edit(realID))
	}

	contributor.Name = name
	contributor.Email = email
	contributor.SetPassword(password)
	contributor.TypeID = cType

	if err := c.Trx.Save(&contributor).Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Error in updating new contributor %s. Error: %s",
			email,
			err,
		)

		// Gorm appears to have no concept of error code.
		// We'll have to transfer the direct SQL Error so user
		// can have more clue of what's happening in case of
		// duplicate email error leaking here.
		c.Flash.Error(err.Error())
		return c.Redirect(routes.Contributor.Edit(realID))
	}

	return c.Redirect(routes.Contributor.List())
}
