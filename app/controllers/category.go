package controllers

import (
	"github.com/revel/revel"
	"net/http"
	"obrolansubuh.com/backend/app/routes"
	"obrolansubuh.com/models"
	"strconv"
)

type Category struct {
	GormController
}

type CategoryList struct {
	ID          int64
	Heading     string
	Description string
	EditLink    string
}

func (c Category) JsonList() revel.Result {
	var categories []models.Category
	if err := c.Trx.Find(&categories).Error; err != nil {

		revel.ERROR.Printf("[LGFATAL] Failed to get category list from database. Error: %s", err)

		FR := FailRequest{Messages: []string{c.Message("errors.db.generic")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	catList := make([]CategoryList, 0, len(categories))
	for _, category := range categories {
		tmp := CategoryList{
			ID:          category.ID,
			Heading:     category.Heading,
			Description: category.Description,
			EditLink:    routes.Category.Edit(category.ID),
		}
		catList = append(catList, tmp)
	}

	return c.RenderJson(catList)
}

func (c Category) List() revel.Result {
	ToolbarItems := []ToolbarItem{
		ToolbarItem{
			Id:   "new-category",
			Text: c.Message("menu.category.new"),
			Icon: "note-add",
			Url:  "Category.New",
		},
	}

	return c.Render(ToolbarItems)
}

func (c Category) New() revel.Result {
	ToolbarItems := []ToolbarItem{
		ToolbarItem{
			Id:   "save-category",
			Text: c.Message("category.button.save"),
			Icon: "note-add",
			Url:  "Category.New",
		},
	}

	return c.Render(ToolbarItems)
}

func (c Category) Save(heading, description, image string) revel.Result {
	c.Validation.Required(heading).Message(c.Message("category.validation.heading"))

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

		FR := FailRequest{Messages: []string{c.Message("errors.db.generic")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	revel.INFO.Printf("[LGINFO] Contributor %s created a category with id %d.",
		c.Session["username"],
		newCat.ID,
	)

	actions := JsonResponse{
		Actions: []Action{
			// The Rel should be updated when we've implemented them
			Action{Uri: "Edit", Rel: routes.Category.Edit(newCat.ID)},
		},
	}

	c.Response.Status = http.StatusCreated
	return c.RenderJson(actions)
}

func (c Category) Edit(id int64) revel.Result {
	ToolbarItems := []ToolbarItem{
		ToolbarItem{
			Id:   "update-category",
			Text: c.Message("menu.category.update"),
			Icon: "editor:publish",
			Url:  "Category.Edit",
		},
	}

	var Category models.Category
	if err := c.Trx.Where("id = ?", id).First(&Category).Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to get Category ID %d data. Error: %s",
			id,
			err,
		)

		c.Flash.Error(c.Message("errors.db.generic"))
		return c.Redirect(routes.Category.List())
	}

	return c.Render(ToolbarItems, Category)
}

func (c Category) Update(id, heading, description, image string) revel.Result {
	c.Validation.Required(id).Message(c.Message("category.validation.id"))
	c.Validation.Required(heading).Message(c.Message("category.validation.heading"))

	if c.Validation.HasErrors() {
		messages := make([]string, 0, len(c.Validation.Errors))
		for _, v := range c.Validation.ErrorMap() {
			messages = append(messages, v.String())
		}

		FR := FailRequest{Messages: messages}

		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(FR)
	}

	realID, iderr := strconv.ParseInt(id, 10, 64)
	if iderr != nil {
		revel.ERROR.Printf("[LGWRD] ID parse error. Value: %s, should be valid int64 value. Parser Error: %s",
			id,
			iderr,
		)

		FR := FailRequest{Messages: []string{c.Message("errors.category.parseid")}}

		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(FR)
	}

	cat := models.Category{
		Heading:     heading,
		Description: description,
		Image:       image,
	}

	if err := c.Trx.Table("categories").Where("id = ?", realID).Updates(&cat).Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Error in updating new category %s. Error: %s",
			heading,
			err,
		)

		FR := FailRequest{Messages: []string{c.Message("errors.db.generic")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	actions := JsonResponse{
		Actions: []Action{
			Action{Uri: "New", Rel: routes.Category.New()},
			Action{Uri: "List", Rel: routes.Category.List()},
		},
	}

	c.Response.Status = http.StatusOK
	return c.RenderJson(actions)
}
