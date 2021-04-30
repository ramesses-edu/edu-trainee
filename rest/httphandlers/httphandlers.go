package httphandlers

import (
	"edu-trainee/rest/authorization"
	"edu-trainee/rest/middleware"
	"edu-trainee/rest/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"

	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

var (
	DB    *gorm.DB
	reNum *regexp.Regexp = regexp.MustCompile(`\d+`)
)

func InitRoutes(router *http.ServeMux, db *gorm.DB) {
	router.Handle("/", mainHandler())
	router.Handle("/public", http.NotFoundHandler())
	router.Handle("/public/", publicHandler())
	router.Handle("/logout/", logoutHandler())
	router.Handle("/auth/", authentification())
	router.Handle("/posts", middleware.Autorization(http.HandlerFunc(postsHandler), db))
	router.Handle("/posts/", middleware.Autorization(http.HandlerFunc(postsHandler), db))
	router.Handle("/comments", middleware.Autorization(http.HandlerFunc(commentsHandler), db))
	router.Handle("/comments/", middleware.Autorization(http.HandlerFunc(commentsHandler), db))
	router.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("localhost/swagger/doc.json"),
	))
}

func authentification() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rPath := r.URL.Path
		reGoogleProvider := regexp.MustCompile(`\/auth\/google(\/)??`)
		reFacebookProvider := regexp.MustCompile(`\/auth\/facebook(\/)??`)
		reTwitterProvider := regexp.MustCompile(`\/auth\/twitter(\/)??`)
		reCallback := regexp.MustCompile(`\/auth\/callback(\/)??\w+`)
		switch {
		case reGoogleProvider.Match([]byte(rPath)):
			authorization.AuthGoogle(w, r)
		case reFacebookProvider.Match([]byte(rPath)):
			authorization.AuthFacebook(w, r)
		case reTwitterProvider.Match([]byte(rPath)):
			authorization.AuthTwitter(w, r)
		case reCallback.Match([]byte(rPath)):
			oauthCallback(w, r)
		}
	})
}

func oauthCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	reProviderGoogle := regexp.MustCompile(`\/auth\/callback\/google(\/)??`)
	reProviderFacebook := regexp.MustCompile(`\/auth\/callback\/facebook(\/)??`)
	reProviderTwitter := regexp.MustCompile(`\/auth\/callback\/twitter(\/)??`)
	switch {
	case reProviderGoogle.Match([]byte(r.URL.Path)):
		authorization.CallbackGoogle(w, r)
	case reProviderFacebook.Match([]byte(r.URL.Path)):
		authorization.CallbackFacebook(w, r)
	case reProviderTwitter.Match([]byte(r.URL.Path)):
		authorization.CallbackTwitter(w, r)
	}
}

func mainHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := authorization.GetCurrentUser(DB, r)
		t, err := template.ParseFiles("./templates/index.html")
		if err != nil {
			fmt.Println(err)
		}
		t.Execute(w, u)
	})
}

type myFileSystem struct {
	fs http.FileSystem
}

func (nfs myFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, _ := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

func publicHandler() http.Handler {
	return http.StripPrefix("/public/", http.FileServer(myFileSystem{fs: http.Dir("./static")}))
}

func logoutHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := authorization.GetCurrentUser(DB, r)
		if u.ID == 0 {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		u.AccessToken = authorization.CalculateSignature(authorization.GenerateAccessToken(), "provider")
		u.UpdateAccessToken(DB)
		cookie := http.Cookie{Name: "UAAT", Path: "/", MaxAge: -1}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusFound)
	})
}

func responseXML(r *http.Request) bool {
	for key := range r.Form {
		if key == "xml" {
			return true
		}
	}
	return false
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	rPath := r.URL.Path
	rePostsComments := regexp.MustCompile(`^\/posts\/\d+\/comments(\/)??$`)
	rePostsID := regexp.MustCompile(`^\/posts\/\d+(\/)??$`)
	rePosts := regexp.MustCompile(`^\/posts(\/)??$`)

	switch {
	case rePosts.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: //list posts with filters
			listPostsHTTP(w, r)
		case http.MethodPost: //create post in:json
			createPostHTTP(w, r)
		case http.MethodPut: //update post  in:json
			updatePostHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":""}`))
			return
		}
	case rePostsID.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: // get posts/{id}
			getPostByIDHTTP(w, r)
		case http.MethodDelete: // delete posts/{id}
			deletePostHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":""}`))
			return
		}
	case rePostsComments.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: // list comments like->/comments?postId={id}
			listPostCommentsHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":""}`))
			return
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
	}
}

//@Summary List posts
//@Description get posts
//@Produce json
//@Param userId query integer false "posts filter by user"
//@Param xml query string false "show data like XML"
//@success 200
//@Failure 400,404
//@Failure 500
//@Failure default
//@Router /posts/ [get]
//@Security ApiKeyAuth
func listPostsHTTP(w http.ResponseWriter, r *http.Request) {
	var param map[string]interface{} = make(map[string]interface{})
	userId := r.FormValue("userId")
	if userId != "" {
		var err error
		param["userId"], err = strconv.Atoi(r.FormValue("userId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":""}`))
			return
		}
	}
	var pp models.Posts
	result := pp.ListPosts(DB, param)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":""}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	if responseXML(r) {
		pp.ResponseXML(w, r)
	} else {
		pp.ResponseJSON(w, r)
	}
}

//@Summary Show a posts
//@Description get post by ID
//@Produce json
//@Param id path integer true "Post ID"
//@Param xml query string false "show data like XML"
//@success 200
//@Failure 400,404
//@Failure 500
//@Failure default
//@Router /posts/{id} [get]
//@Security ApiKeyAuth
func getPostByIDHTTP(w http.ResponseWriter, r *http.Request) {
	var param map[string]interface{} = make(map[string]interface{})
	var err error
	param["id"], err = strconv.Atoi(reNum.FindString(r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var p models.Post
	result := p.GetPost(DB, param)
	if result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var pp models.Posts = models.Posts{Posts: []models.Post{p}}
	if responseXML(r) {
		pp.ResponseXML(w, r)
	} else {
		pp.ResponseJSON(w, r)
	}
}

type createPostStruct struct {
	Title string
	Body  string
}

//@Summary Create post
//@Description create post
//@Accept json
//@Produce json
//@Param RequestPost body createPostStruct true "JSON structure for creating post"
//@Success 200,201
//@Failure 400
//@Failure default
//@Router /posts/ [POST]
//@Security ApiKeyAuth
func createPostHTTP(w http.ResponseWriter, r *http.Request) {
	u := authorization.GetCurrentUser(DB, r)
	if u.ID == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var p models.Post
	err = json.Unmarshal(reqBody, &p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	p.UserID = u.ID
	result := p.CreatePost(DB)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	jsonP, _ := json.MarshalIndent(p, "", "  ")
	fmt.Fprintln(w, string(jsonP))
}

//@Summary Update post
//@Description update post
//@Accept json
//@Produce json
//@Param RequestPost body models.Post true "JSON structure for updating post"
//@Success 200
//@Failure 400
//@Failure default
//@Router /posts/ [put]
//@Security ApiKeyAuth
func updatePostHTTP(w http.ResponseWriter, r *http.Request) {
	u := authorization.GetCurrentUser(DB, r)
	if u.ID == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var p models.Post
	err = json.Unmarshal(reqBody, &p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	p.UserID = u.ID
	result := p.UpdatePost(DB)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	jsonP, _ := json.MarshalIndent(p, "", "  ")
	fmt.Fprintln(w, string(jsonP))
}

//@Summary Delete post
//@Description delete post by ID
//@Param id path int true "ID of deleting post"
//@Success 200
//@Failure default
//@Router /posts/{id} [delete]
//@Security ApiKeyAuth
func deletePostHTTP(w http.ResponseWriter, r *http.Request) {
	u := authorization.GetCurrentUser(DB, r)
	if u.ID == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}

	pID, err := strconv.Atoi(reNum.FindString(r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var p models.Post = models.Post{ID: pID, UserID: u.ID}
	result := p.DeletePost(DB)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	w.WriteHeader(http.StatusOK)
}

//@Summary List comments of post
//@Description List comments like request /comments?postId={id}
//@Param id path int true "ID of post"
//@Param xml query string false "show data like XML"
//@Router /posts/{id}/comments [get]
//@Success 200
//@Failure default
//@Security ApiKeyAuth
func listPostCommentsHTTP(w http.ResponseWriter, r *http.Request) {
	var param map[string]interface{} = make(map[string]interface{})
	var err error
	param["postId"], err = strconv.Atoi(reNum.FindString(r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var cc models.Comments
	result := cc.ListComments(DB, param)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	if responseXML(r) {
		cc.ResponseXML(w, r)
	} else {
		cc.ResponseJSON(w, r)
	}
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	rPath := r.URL.Path
	reCommentsID := regexp.MustCompile(`^\/comments\/\d+(\/)??$`)
	reComments := regexp.MustCompile(`^\/comments(\/)??$`)

	switch {
	case reComments.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: // list comments with filters
			listCommentsHTTP(w, r)
		case http.MethodPost: // create comment in:json
			createCommentHTTP(w, r)
		case http.MethodPut: // update comment in:json
			updateCommentHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":""}`))
			return
		}
	case reCommentsID.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: // get comments/{id}
			getCommentByIDHTTP(w, r)
		case http.MethodDelete: // delete comments/{id}
			deleteCommentHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":""}`))
			return
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
	}
}

//@Summary List comments
//@description list comments with filtering
//@Param postId query int false "ID of post"
//@Param xml query string false "show data like XML"
//@Success 200
//@Failure default
//@Router /comments/ [get]
//@Security ApiKeyAuth
func listCommentsHTTP(w http.ResponseWriter, r *http.Request) {
	var param map[string]interface{} = make(map[string]interface{})
	postId := r.FormValue("postId")
	if postId != "" {
		var err error
		param["postId"], err = strconv.Atoi(r.FormValue("postId"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":""}`))
			return
		}
	}
	var cc models.Comments
	result := cc.ListComments(DB, param)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	if responseXML(r) {
		cc.ResponseXML(w, r)
	} else {
		cc.ResponseJSON(w, r)
	}
}

//@Summary Show comment
//@description Get comment by ID
//@Param id path int true "ID of comment"
//@Param xml query string false "show data like XML"
//@Success 200
//@Failure default
//@Router /comments/{id} [get]
//@Security ApiKeyAuth
func getCommentByIDHTTP(w http.ResponseWriter, r *http.Request) {
	var param map[string]interface{} = make(map[string]interface{})
	var err error
	param["id"], err = strconv.Atoi(reNum.FindString(r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var cmnt models.Comment
	result := cmnt.GetComment(DB, param)
	if result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var cc models.Comments = models.Comments{Comments: []models.Comment{cmnt}}
	if responseXML(r) {
		cc.ResponseXML(w, r)
	} else {
		cc.ResponseJSON(w, r)
	}
}

type createCommentStruct struct {
	PostID int `json:"postId"`
	Name   string
	Email  string
	Body   string
}

//@Summary Create comment
//@description create comment
//@Accept json
//@Produce json
//@Param RequestPost body createCommentStruct true "JSON structure for creating post"
//@Success 200,201
//@Failure 400
//@Failure default
//@Router /comments/ [post]
//@Security ApiKeyAuth
func createCommentHTTP(w http.ResponseWriter, r *http.Request) {
	u := authorization.GetCurrentUser(DB, r)
	if u.ID == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var c models.Comment
	err = json.Unmarshal(reqBody, &c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	c.UserID = u.ID
	result := c.CreateComment(DB)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	jsonC, _ := json.MarshalIndent(c, "", "  ")
	fmt.Fprintln(w, string(jsonC))
}

//@Summary Update comment
//@description update comment
//@Accept json
//@Produce json
//@Param RequestPost body models.Comment true "JSON structure for creating post"
//@Success 200
//@Failure 400
//@Failure default
//@Router /comments/ [put]
//@Security ApiKeyAuth
func updateCommentHTTP(w http.ResponseWriter, r *http.Request) {
	u := authorization.GetCurrentUser(DB, r)
	if u.ID == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var c models.Comment
	err = json.Unmarshal(reqBody, &c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	c.UserID = u.ID
	result := c.UpdateComment(DB)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	jsonC, _ := json.MarshalIndent(c, "", "  ")
	fmt.Fprintln(w, string(jsonC))
}

//@Summary Delete comment
//@descripton delete comment by ID
//@Param id path int true "ID of deleting comment"
//@Success 200
//@Failure default
//@Router /comments/{id} [delete]
//@Security ApiKeyAuth
func deleteCommentHTTP(w http.ResponseWriter, r *http.Request) {
	u := authorization.GetCurrentUser(DB, r)
	if u.ID == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}

	cID, err := strconv.Atoi(reNum.FindString(r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var c models.Comment = models.Comment{ID: cID, UserID: u.ID}
	result := c.DeleteComment(DB)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	w.WriteHeader(http.StatusOK)
}
