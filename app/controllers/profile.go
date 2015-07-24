package controllers

import (
	"github.com/revel/revel"
	"obrolansubuh.com/backend/app/routes"
	"obrolansubuh.com/models"
)

type Profile struct {
	GormController
}

func (c Profile) Edit() revel.Result {
	contributor := &models.Contributor{}
	c.Trx.Where("id = ?", c.Session["userid"]).First(&contributor)

	if err := c.Trx.Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to get user in edit profile. Error: %s", err)

		c.Flash.Error(c.Message("errors.db.generic"))
		return c.Redirect(routes.App.Index())
	}

	return c.Render(contributor)
}

func (c Profile) Update(
	name string,
	email string,
	about string,
	photo []byte) revel.Result {

	id := c.Session["userid"]

	contributor := &models.Contributor{}
	c.Trx.Where("id = ?", id).First(&contributor)

	contributor.Name = name
	contributor.Email = email
	contributor.About = about

	c.Session["user"] = contributor.Email
	c.Session["username"] = contributor.Name

	if len(photo) > 0 {
		fileName := c.Params.Files["photo"][0].Filename
		uploadDr := revel.Config.StringDefault("upload.image.location", "/public/upload/")
		hashName := hashFileName(fileName, c.Session["user"])
		fullName := uploadDr + hashName

		if err := saveFile(photo, revel.BasePath+fullName); err != nil {
			revel.ERROR.Printf("[LGFATAL] Failed to upload user %d profile picture to %s. Error: %s",
				id,
				fullName,
				err,
			)

			c.Flash.Error(c.Message("errors.upload.image"))
			return c.Redirect(routes.Profile.Edit())
		}

		contributor.Photo = fullName
		c.Session["userphoto"] = fullName
	}

	c.Trx.Save(&contributor)

	if err := c.Trx.Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to get user in edit profile. Error: %s", err)

		c.Flash.Error(c.Message("errors.db.generic"))
		return c.Redirect(routes.Profile.Edit())
	}

	revel.INFO.Printf("[LGINFO] Successfully updated %s's profile.", email)

	c.Flash.Success(c.Message("profile.update.success"))
	return c.Redirect(routes.Profile.Edit())
}
