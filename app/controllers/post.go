package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"io/ioutil"
	"net/http"
	"obrolansubuh.com/backend/app/routes"
	"obrolansubuh.com/models"
	"strconv"
	"strings"
	"time"
)

type Post struct {
	GormController
}

type Link struct {
	Rel string `json:"rel"`
	Uri string `json:"uri"`
}

type PostCreated struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Links []Link `json:"links"`
}

type PostUpdated struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type PostPublishToogled struct {
	ID      int64  `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type PostList struct {
	ID         int64
	Title      string
	Content    string
	Author     string
	Category   string
	CreatedAt  time.Time
	Published  bool
	EditLink   string
	ToggleLink string
}

func getUsersPost(uid string, allUsers bool, category int64, published string, db *gorm.DB) (posts []models.Post, err error) {
	db = db.Preload("Author").Order("created_at desc")
	if !allUsers {
		db = db.Where("author_id = ?", uid)
	}

	if category != 0 {
		db = db.Joins(
			"join post_categories as pc on posts.id = pc.post_id " +
				"join categories as c on c.id = pc.category_id")

		db = db.Where("c.id = ?", category)
	}

	published = strings.ToLower(published)
	if published != "all" && (published == "true" || published == "false") {
		pub := published == "true"

		db = db.Where("published = ?", pub)
	}

	// the Select call is needed because of a bug.
	// TODO: remove this call when the bug is fixed
	db = db.Select("posts.id, posts.title, posts.content, posts.published, posts.created_at, posts.updated_at, posts.author_id")
	if err = db.Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}

func (c Post) isAdmin() bool {
	return c.Session["usertype"] == "ADMIN"
}

func (c Post) JsonList(category int64, published string) revel.Result {
	// enforce using cookies here
	// so people can't just API call this easily
	uid := c.Session["userid"]
	posts, err := getUsersPost(uid, c.isAdmin(), category, published, c.Trx)

	if err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to get post list from database.")

		FR := FailRequest{Messages: []string{c.Message("errors.post.database")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	postList := make([]PostList, 0, len(posts))
	for _, post := range posts {
		var cat models.Category
		c.Trx.Model(&post).Association("Categories").Find(&cat)

		tmp := PostList{
			ID:         post.ID,
			Title:      post.Title,
			Author:     post.Author.Name,
			Category:   cat.Heading,
			Content:    post.Content,
			CreatedAt:  post.CreatedAt,
			Published:  post.Published,
			EditLink:   routes.Post.Edit(post.ID),
			ToggleLink: routes.Post.TogglePublished(post.ID),
		}
		postList = append(postList, tmp)
	}

	return c.RenderJson(postList)
}

func (c Post) List() revel.Result {
	ToolbarItems := []ToolbarItem{
		ToolbarItem{Id: "new-post", Text: c.Message("menu.post.new"), Icon: "note-add", Url: "Post.New"},
	}

	return c.Render(ToolbarItems)
}

func (c Post) TogglePublished(id int64) revel.Result {
	// Non admins are not allowed to publish
	if !c.isAdmin() {
		FR := FailRequest{
			Messages: []string{c.Message("errors.privilege.generic")},
		}

		c.Response.Status = http.StatusForbidden
		return c.RenderJson(FR)
	}

	// Gorm can't set boolean field to false via methods.
	//
	// see:
	// https://github.com/jinzhu/gorm/issues/398
	// https://github.com/jinzhu/gorm/issues/469
	c.Trx.Exec("UPDATE posts SET published = NOT published WHERE posts.id = ?", id)

	if err := c.Trx.Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to toogle post's update status. Check controllers/post.go:118.")

		PPT := PostPublishToogled{ID: id, Success: false, Message: c.Message("post.publishtoggle.fail")}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(PPT)
	}

	PPT := PostPublishToogled{ID: id, Success: true, Message: c.Message("post.publishtoggle.success")}
	return c.RenderJson(PPT)
}

func (c Post) New() revel.Result {
	ToolbarItems := []ToolbarItem{
		ToolbarItem{
			Id:   "save-draft",
			Text: c.Message("menu.post.savedraft"),
			Icon: "save",
			Url:  "Post.New",
		},
	}

	if c.Session["usertype"] == "ADMIN" {
		publishMenu := ToolbarItem{
			Id:   "publish-post",
			Text: c.Message("menu.post.publish"),
			Icon: "editor:publish",
			Url:  "Post.New",
		}

		ToolbarItems = append(ToolbarItems, publishMenu)

	}

	return c.Render(ToolbarItems)
}

func (c Post) Save(title string, content string, category int64, publish bool) revel.Result {
	// Non admins are not allowed to publish
	if !c.isAdmin() && publish {
		FR := FailRequest{
			Messages: []string{c.Message("errors.privilege.generic")},
		}

		c.Response.Status = http.StatusForbidden
		return c.RenderJson(FR)
	}

	c.Validation.Required(title).Message(c.Message("post.validation.title"))
	c.Validation.Required(content).Message(c.Message("post.validation.content"))
	c.Validation.Required(category).Message(c.Message("post.validation.category"))
	c.Validation.MaxSize(title, 1024).Message(c.Message("post.validation.title_length"))

	if c.Validation.HasErrors() {
		messages := make([]string, 0, len(c.Validation.Errors))
		for _, v := range c.Validation.ErrorMap() {
			messages = append(messages, v.String())
		}

		FR := FailRequest{Messages: messages}

		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(FR)
	}

	contributor, gcErr := c.GetContributor(c.Session["user"])

	if gcErr != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to get author info when creating new post. Likely a database problem.")

		FR := FailRequest{Messages: []string{c.Message("errors.post.database")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	// cleanups
	title = strings.Trim(title, " \n")
	content = strings.Trim(content, " \n")

	newPost := models.Post{
		Title:     title,
		Content:   content,
		Author:    contributor,
		Published: publish,
	}

	var cat models.Category
	c.Trx.Create(&newPost)
	c.Trx.Where("id = ?", category).First(&cat)
	c.Trx.Model(&newPost).Association("Categories").Append(&cat)

	if err := c.Trx.Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to save post in database.")

		FR := FailRequest{Messages: []string{c.Message("errors.post.database")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	revel.INFO.Printf("[LGINFO] Contributor %s created a post with id %d at %s.",
		c.Session["username"],
		newPost.ID,
		newPost.CreatedAt,
	)

	link := Link{Rel: "post/edit", Uri: routes.Post.Edit(newPost.ID)}
	PC := PostCreated{ID: newPost.ID, Title: newPost.Title, Links: []Link{link}}

	c.Response.Status = http.StatusCreated
	return c.RenderJson(PC)
}

func (c Post) Edit(id int64) revel.Result {
	ToolbarItems := []ToolbarItem{
		ToolbarItem{
			Id:       "update-post",
			Text:     c.Message("menu.post.update"),
			Icon:     "editor:publish",
			Url:      "Post.Update",
			UrlParam: strconv.FormatInt(id, 10),
		},
	}

	post := models.Post{}
	c.Trx.Preload("Author").Where("id = ?", id).First(&post)

	var cat models.Category
	c.Trx.Model(&post).Association("Categories").Find(&cat)

	// you can only edit your own post (except admin)
	userid, _ := strconv.ParseInt(c.Session["userid"], 10, 64)
	if post.Author.ID != userid && !c.isAdmin() {
		c.Flash.Error(c.Message("errors.post.privilege"))
		return c.Redirect(routes.Post.List())
	}

	return c.Render(post, cat, ToolbarItems)
}

func (c Post) Update(id int64) revel.Result {
	c.Validation.Required(id)

	if c.Validation.HasErrors() {
		FR := FailRequest{Messages: []string{c.Message("post.validation.id")}}

		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(FR)
	}

	var p models.Post
	data, ioerr := ioutil.ReadAll(c.Request.Body)
	if ioerr != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to read request body on %s. Error: %s",
			"Post.Update",
			ioerr)

		FR := FailRequest{Messages: []string{c.Message("errors.post.request")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	var tmpDat map[string]interface{}
	jserr := json.Unmarshal(data, &tmpDat)
	if jserr != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to decode JSON from client on %s. Error: %s.",
			"Post.Update",
			jserr)

		FR := FailRequest{Messages: []string{c.Message("errors.post.json")}}

		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(FR)
	}

	p.Title = strings.Trim(tmpDat["title"].(string), " \n")
	p.Content = strings.Trim(tmpDat["content"].(string), " \n")

	var oldPost models.Post
	c.Trx.Preload("Author").Where("id = ?", id).First(&oldPost)

	// you can only edit your own post (except admin)
	userid, _ := strconv.ParseInt(c.Session["userid"], 10, 64)
	if oldPost.Author.ID != userid && !c.isAdmin() {
		FR := FailRequest{
			Messages: []string{c.Message("errors.post.privilege")},
		}

		c.Response.Status = http.StatusForbidden
		return c.RenderJson(FR)
	}

	// only admins are allowed to change publish status
	if oldPost.Published != p.Published && !c.isAdmin() {
		FR := FailRequest{
			Messages: []string{c.Message("errors.privilege.generic")},
		}

		c.Response.Status = http.StatusForbidden
		return c.RenderJson(FR)
	}

	var oldCat, newCat models.Category
	newCatID, _ := strconv.ParseInt(tmpDat["category"].(string), 10, 64)

	c.Trx.Model(&oldPost).Association("Categories").Find(&oldCat)
	c.Trx.Where("id = ?", newCatID).First(&newCat)

	c.Trx.Model(&oldPost).Updates(p)
	c.Trx.Model(&oldPost).Association("Categories").Delete(oldCat)
	c.Trx.Model(&oldPost).Association("Categories").Replace(newCat)
	if err := c.Trx.Error; err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to save post in database.")

		FR := FailRequest{Messages: []string{c.Message("errors.post.database")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	revel.INFO.Printf("[LGINFO] Contributor %s updated a post with id %d at %s.",
		c.Session["username"],
		p.ID,
		p.CreatedAt,
	)

	message := fmt.Sprintf(c.Message("post.update.success"), p.Title)

	PU := PostUpdated{
		ID:      p.ID,
		Title:   p.Title,
		Message: message}
	c.Response.Status = http.StatusOK // should we use Created for update too?
	return c.RenderJson(PU)
}
