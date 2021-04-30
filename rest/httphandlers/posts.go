package httphandlers

import (
	"edu-trainee/rest/authorization"
	"edu-trainee/rest/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"gorm.io/gorm"
)

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
		xmlWrite(w, pp)
	} else {
		jsonWrite(w, pp.Posts)
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
	if responseXML(r) {
		xmlWrite(w, p)
	} else {
		jsonWrite(w, p)
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
		xmlWrite(w, cc)
	} else {
		jsonWrite(w, cc.Comments)
	}
}
