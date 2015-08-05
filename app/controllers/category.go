package controllers

import (
	"github.com/revel/revel"
	"net/http"
	//"obrolansubuh.com/backend/app/routes"
	"obrolansubuh.com/models"
)

type Category struct {
	GormController
}

func (c Category) New() revel.Result {
	ToolbarItems := []ToolbarItem{
		ToolbarItem{Id: "save-category", Text: "Save", Icon: "note-add", Url: "Category.New"},
	}

	return c.Render(ToolbarItems)
}

func (c Category) Save(heading, description, image string) revel.Result {
	c.Validation.Required(heading).Message("Kurang Heading")
	c.Validation.Required(description).Message("Kurang Description")

	if c.Validation.HasErrors() {
		messages := make([]string, 0, len(c.Validation.Errors))
		for _, v := range c.Validation.ErrorMap() {
			messages = append(messages, v.String())
		}

		FR := FailRequest{Messages: messages}

		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(FR)
	}

	newCat := models.Category{
		Heading:     heading,
		Description: description,
		Image:       image,
	}

	if err := c.Trx.Create(&newCat).Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to save new category in database.")

		FR := FailRequest{Messages: []string{"Error database"}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	actions := JsonResponse{
		Actions: []Action{
			// The Rel should be updated when we've implemented them
			Action{Uri: "Edit", Rel: "/category/edit/1"},
			Action{Uri: "Delete", Rel: "/category/delete/1"},
		},
	}

	c.Response.Status = http.StatusCreated
	return c.RenderJson(actions)
}
