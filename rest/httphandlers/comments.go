package httphandlers

import (
	"edu-trainee/rest/application"
	"edu-trainee/rest/authorization"
	"edu-trainee/rest/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

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
	DB:= application.CurrentApplication().DB
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
		xmlWrite(w, cc)
	} else {
		jsonWrite(w, cc.Comments)
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
	DB := application.CurrentApplication().DB
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

	if responseXML(r) {
		xmlWrite(w, cmnt)
	} else {
		jsonWrite(w, cmnt)
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
	DB := application.CurrentApplication().DB
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
	DB := application.CurrentApplication().DB
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
	DB := application.CurrentApplication().DB
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
