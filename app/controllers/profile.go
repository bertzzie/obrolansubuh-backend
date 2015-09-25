package controllers

import (
	"github.com/revel/revel"
	"obrolansubuh.com/backend/app/routes"
	"obrolansubuh.com/models"
	"regexp"
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
	handle string,
	email string,
	about string,
	photo []byte) revel.Result {

	c.Validation.Required(name).Message(c.Message("contributor.validation.name"))
	c.Validation.Required(handle).Message(c.Message("contributor.validation.handle.required"))
	c.Validation.Match(handle, regexp.MustCompile(`^\w*$`)).Message(c.Message("contributor.validation.handle.invalid"))
	c.Validation.Required(email).Message("contributor.validation.email.required")
	c.Validation.Email(email).Message("contributor.validation.email.invalid")

	id := c.Session["userid"]

	/*
		// existing contributor check
		_, dupe := c.GetContributor(email)
		if dupe == nil {
			c.Flash.Error(c.Message("contributor.validation.email.duplicate"))
			return c.Redirect(routes.Profile.Edit())
		}

		_, handleDupe := c.GetContributorByHandle(handle)
		if handleDupe == nil {
			c.Flash.Error(c.Message("contributor.validation.handle.duplicate"))
			return c.Redirect(routes.Profile.Edit())
		}
	*/
	if c.IsEmailDupe(id, email) {
		c.Flash.Error(c.Message("contributor.validation.email.duplicate"))
		return c.Redirect(routes.Profile.Edit())
	}

	if c.IsHandleDupe(id, handle) {
		c.Flash.Error(c.Message("contributor.validation.handle.duplicate"))
		return c.Redirect(routes.Profile.Edit())
	}

	contributor := &models.Contributor{}
	c.Trx.Where("id = ?", id).First(&contributor)

	contributor.Name = name
	contributor.Handle = handle
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

func (c Profile) ChangePassword() revel.Result {
	return c.Render()
}

func (c Profile) UpdatePassword(currentPassword, newPassword, retypePassword string) revel.Result {
	id := c.Session["userid"]
	contributor := &models.Contributor{}
	c.Trx.Where("id = ?", id).First(&contributor)

	if contributor.CheckPassword(currentPassword) {
		if newPassword == retypePassword {
			contributor.SetPassword(newPassword)
			c.Trx.Save(&contributor)

			revel.INFO.Printf("[LGINFO] User %s changed their password.", contributor.Email)

			c.Flash.Success(c.Message("profile.password.success"))
			return c.Redirect(routes.Profile.ChangePassword())
		}

		c.Flash.Error(c.Message("profile.password.retypefailed"))
		return c.Redirect(routes.Profile.ChangePassword())
	}

	c.Flash.Error(c.Message("profile.password.wrong"))
	return c.Redirect(routes.Profile.ChangePassword())
}

func (c Profile) IsEmailDupe(id, email string) bool {
	con := &models.Contributor{}
	tx := c.Trx.Where("id != ? AND email = ?", id, email).First(&con)

	return tx.Error == nil
}

func (c Profile) IsHandleDupe(id, handle string) bool {
	con := &models.Contributor{}
	tx := c.Trx.Where("id != ? AND handle = ?", id, handle).First(&con)

	return tx.Error == nil
}
