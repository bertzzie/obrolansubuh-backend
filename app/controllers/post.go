package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"io/ioutil"
	"net/http"
	"obrolansubuh.com/backend/app/routes"
	"obrolansubuh.com/models"
	"path/filepath"
	"strconv"
	"time"
)

type Post struct {
	GormController
}

type UploadedFile struct {
	Url         string `json:"url"`
	ContentType string `json:"type"`
	ContentSize int    `json:"size"`
	DeleteURL   string `json:"deleteUrl"`
	DeleteType  string `json:"deleteType"`
}

type FileUploadResponse struct {
	Files []UploadedFile `json:"files"`
}

type FailedUpload struct {
	Name  string `json:"name"`
	Size  int    `json:"size"`
	Error string `json:"error"`
}

type FileUploadError struct {
	Files []FailedUpload `json:"files"`
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

func getUsersPost(uid string, db *gorm.DB) (posts []models.Post, err error) {
	db.Where("author_id = ?", uid).Find(&posts)
	if err = db.Error; err != nil {
		return nil, err
	}

	return posts, nil
}

type PostList struct {
	Title      string
	Content    string
	CreatedAt  time.Time
	Published  bool
	EditLink   string
	ToggleLink string
}

func (c Post) JsonList() revel.Result {
	// enforce using cookies here
	// so people can't just API call this easily
	uid := c.Session["userid"]
	posts, err := getUsersPost(uid, c.Trx)

	if err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to get post list from database.")

		FR := FailRequest{Messages: []string{c.Message("errors.post.database")}}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FR)
	}

	postList := make([]PostList, 0, len(posts))
	for _, post := range posts {
		tmp := PostList{
			Title:      post.Title,
			Content:    post.Content,
			CreatedAt:  post.CreatedAt,
			Published:  post.Published,
			EditLink:   routes.Post.Edit(post.ID),
			ToggleLink: "/toggle",
		}
		postList = append(postList, tmp)
	}

	return c.RenderJson(postList)
}

func (c Post) List() revel.Result {
	return c.Render()
}

func (c Post) New() revel.Result {
	ToolbarItems := []ToolbarItem{
		ToolbarItem{Id: "publish-post", Text: "Publish", Icon: "editor:publish", Url: "Post.New"},
		ToolbarItem{Id: "save-draft", Text: "Save Draft", Icon: "save", Url: "Post.New"},
	}

	return c.Render(ToolbarItems)
}

func (c Post) Save(title string, content string, publish bool) revel.Result {
	c.Validation.Required(title).Message(c.Message("post.validation.title"))
	c.Validation.Required(content).Message(c.Message("post.validation.content"))
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

	newPost := models.Post{
		Title:     title,
		Content:   content,
		Author:    contributor,
		Published: publish,
	}

	c.Trx.Create(&newPost)
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
		ToolbarItem{Id: "update-post", Text: "Update", Icon: "editor:publish", Url: "Post.Update", UrlParam: strconv.FormatInt(id, 10)},
	}

	post := models.Post{}
	c.Trx.Where("id = ?", id).First(&post)

	return c.Render(post, ToolbarItems)
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
		c.RenderJson(FR)
	}

	jserr := json.Unmarshal(data, &p)
	if jserr != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to decode JSON from client on %s. Error: %s.",
			"Post.Update",
			jserr)

		FR := FailRequest{Messages: []string{c.Message("errors.post.json")}}

		c.Response.Status = http.StatusBadRequest
		c.RenderJson(FR)
	}

	c.Trx.Table("posts").Where("id = ?", p.ID).Updates(p)
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

func (c Post) ImageUpload(image []byte) revel.Result {
	fileType := c.Params.Files["image"][0].Header["Content-Type"]
	fileName := c.Params.Files["image"][0].Filename

	hostname := revel.Config.StringDefault("server.hostname", "http://localhost:9000")
	uploadDr := revel.Config.StringDefault("upload.image.location", "/public/upload/")

	hashName := hashFileName(fileName, c.Session["user"])

	fullName := uploadDr + hashName

	if err := saveFile(image, revel.BasePath+fullName); err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to write uploaded file %s by %s. Error: %s",
			revel.BasePath+fullName,
			c.Session["user"],
			err)

		failedUpload := FailedUpload{
			Name:  fullName,
			Size:  len(image),
			Error: c.Message("errors.upload.image"),
		}

		FUR := FileUploadError{
			Files: []FailedUpload{failedUpload},
		}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FUR)
	} else {
		fileInfo := UploadedFile{
			Url:         hostname + fullName,
			ContentType: fileType[0],
			ContentSize: len(image),
			DeleteURL:   hostname + "/post/image/delete/" + fileName,
			DeleteType:  "DELETE",
		}

		FUR := FileUploadResponse{
			Files: []UploadedFile{fileInfo},
		}

		return c.RenderJson(FUR)
	}
}

func saveFile(file []byte, destination string) error {
	err := ioutil.WriteFile(destination, file, 0644) // Permission: -rw-r--r--

	if err != nil {
		return err
	}

	return nil
}

// Hash function to change filename so we always have unique
// filename for uploaded file.
//
// Hash function is a simple
func hashFileName(filename string, username string) string {
	ext := filepath.Ext(filename)

	fullname := filename + "_" + username + "_" + time.Now().Format("20060102150405")
	sum := md5.Sum([]byte(fullname))
	hash := hex.EncodeToString(sum[:])

	result := hash + ext

	return result

}
