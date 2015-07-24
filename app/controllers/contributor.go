package controllers

import (
	"fmt"
	"github.com/revel/revel"
	"obrolansubuh.com/backend/app/routes"
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

func (c Contributor) UpdateProfile(
	id int64,
	name string,
	email string,
	about string,
	photo []byte) revel.Result {

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
			c.Flash.Error("GAGAL")
			return c.Redirect(routes.Contributor.EditProfile(id))
		}

		contributor.Photo = fullName
		c.Session["userphoto"] = fullName
	}

	c.Trx.Save(&contributor)

	c.Flash.Success("Berhasil!")
	return c.Redirect(routes.Contributor.EditProfile(id))
}
