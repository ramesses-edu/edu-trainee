package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"gorm.io/gorm"
)

var reNum *regexp.Regexp = regexp.MustCompile(`\d+`)

func mainHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := getCurrentUser(a.db, r)
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
		u := getCurrentUser(a.db, r)
		if u.ID == 0 {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		u.AccessToken = calculateSignature(generateAccessToken(), "provider")
		u.updateAccessToken(a.db)
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
	var pp posts
	result := pp.listPosts(a.db, param)
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
		pp.responseXML(w, r)
	} else {
		pp.responseJSON(w, r)
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
	var p post
	result := p.getPost(a.db, param)
	if result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var pp posts = posts{Posts: []post{p}}
	if responseXML(r) {
		pp.responseXML(w, r)
	} else {
		pp.responseJSON(w, r)
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
	u := getCurrentUser(a.db, r)
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
	var p post
	err = json.Unmarshal(reqBody, &p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	p.UserID = u.ID
	result := p.createPost(a.db)
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
//@Param RequestPost body post true "JSON structure for updating post"
//@Success 200
//@Failure 400
//@Failure default
//@Router /posts/ [put]
//@Security ApiKeyAuth
func updatePostHTTP(w http.ResponseWriter, r *http.Request) {
	u := getCurrentUser(a.db, r)
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
	var p post
	err = json.Unmarshal(reqBody, &p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	p.UserID = u.ID
	result := p.updatePost(a.db)
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
	u := getCurrentUser(a.db, r)
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
	var p post = post{ID: pID, UserID: u.ID}
	result := p.deletePost(a.db)
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
	var cc comments
	result := cc.listComments(a.db, param)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	if responseXML(r) {
		cc.responseXML(w, r)
	} else {
		cc.responseJSON(w, r)
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
	var cc comments
	result := cc.listComments(a.db, param)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	if responseXML(r) {
		cc.responseXML(w, r)
	} else {
		cc.responseJSON(w, r)
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
	var cmnt comment
	result := cmnt.getComment(a.db, param)
	if result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":""}`))
		return
	}
	var cc comments = comments{Comments: []comment{cmnt}}
	if responseXML(r) {
		cc.responseXML(w, r)
	} else {
		cc.responseJSON(w, r)
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
	u := getCurrentUser(a.db, r)
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
	var c comment
	err = json.Unmarshal(reqBody, &c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	c.UserID = u.ID
	result := c.createComment(a.db)
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
//@Param RequestPost body comment true "JSON structure for creating post"
//@Success 200
//@Failure 400
//@Failure default
//@Router /comments/ [put]
//@Security ApiKeyAuth
func updateCommentHTTP(w http.ResponseWriter, r *http.Request) {
	u := getCurrentUser(a.db, r)
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
	var c comment
	err = json.Unmarshal(reqBody, &c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	c.UserID = u.ID
	result := c.updateComment(a.db)
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
	u := getCurrentUser(a.db, r)
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
	var c comment = comment{ID: cID, UserID: u.ID}
	result := c.deleteComment(a.db)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return
	}
	w.WriteHeader(http.StatusOK)
}
