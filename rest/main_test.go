package main

import (
	"bytes"
	"edu-trainee/rest/application"
	"edu-trainee/rest/authorization"
	"edu-trainee/rest/httphandlers"
	"edu-trainee/rest/models"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestMain(t *testing.M) {
	a = application.New()
	defer application.Close()
	httphandlers.InitRoutes(a.Router)

	createTestUser(a.DB)
	code := t.Run()
	clearTableUsers()
	clearTablePosts()
	clearTableComments()
	os.Exit(code)
}

func createTestUser(db *gorm.DB) {
	var u models.User = models.User{Login: "test", Name: "test", Provider: "test", AccessToken: authorization.CalculateSignature("test", "provider")}
	result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&u)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
}
func clearTableUsers() {
	tx := a.DB.Begin()
	tx.Exec("DELETE FROM users")
	tx.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
	tx.Commit()
}
func clearTablePosts() {
	tx := a.DB.Begin()
	tx.Exec("DELETE FROM posts")
	tx.Exec("ALTER TABLE posts AUTO_INCREMENT = 1")
	tx.Commit()
}
func clearTableComments() {
	tx := a.DB.Begin()
	tx.Exec("DELETE FROM comments")
	tx.Exec("ALTER TABLE comments AUTO_INCREMENT = 1")
	tx.Commit()
}
func execRequest(request *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, request)
	return rr
}
func checkRespCode(t *testing.T, expectedCode, actualCode int) {
	if expectedCode != actualCode {
		t.Errorf("Expected response code: %d. Actual response code: %d", expectedCode, actualCode)
	}
}
func TestEmptyPostTable(t *testing.T) {
	clearTablePosts()
	request, _ := http.NewRequest(http.MethodGet, "/posts/", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	if body := resp.Body.String(); body != "[]" {
		t.Errorf("Expected empty array []. Got %s", body)
	}
}
func TestEmptyCommentsTable(t *testing.T) {
	clearTableComments()
	request, _ := http.NewRequest(http.MethodGet, "/comments/", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	if body := resp.Body.String(); body != "[]" {
		t.Errorf("Expected empty array []. Got %s", body)
	}
}
func TestUnautorizedAccess(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/posts/", nil)
	resp := execRequest(request)
	checkRespCode(t, http.StatusNetworkAuthenticationRequired, resp.Code)
	body := resp.Body.Bytes()
	umBody := make(map[string]interface{})
	json.Unmarshal(body, &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
	request, _ = http.NewRequest(http.MethodGet, "/comments/", nil)
	resp = execRequest(request)
	checkRespCode(t, http.StatusNetworkAuthenticationRequired, resp.Code)
	body = resp.Body.Bytes()
	umBody = make(map[string]interface{})
	json.Unmarshal(body, &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
}
func TestGetNonExistenPost(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/posts/234", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusNotFound, resp.Code)
	umBody := make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
}
func TestGetNonExistenComment(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/comments/234", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusNotFound, resp.Code)
	umBody := make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
}
func addPosts(count int) {
	tx := a.DB.Begin()
	for i := 0; i < count; i++ {
		tx.Exec("INSERT INTO posts (title,body,userId) VALUES(?,?,?)", "title", "body", 1)
	}
	tx.Commit()
}
func addComments(count, postid int) {
	tx := a.DB.Begin()
	for i := 0; i < count; i++ {
		tx.Exec("INSERT INTO comments (name,email,body,postId,userId) VALUES(?,?,?,?,?)", "name", "email", "body", postid, 1)
	}
	tx.Commit()
}
func TestNonAllowedMethod(t *testing.T) {
	//method connect
	request, _ := http.NewRequest(http.MethodConnect, "/posts", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusMethodNotAllowed, resp.Code)
	umBody := make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
	request, _ = http.NewRequest(http.MethodConnect, "/posts/1", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusMethodNotAllowed, resp.Code)
	umBody = make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
	request, _ = http.NewRequest(http.MethodConnect, "/posts/1/comments", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusMethodNotAllowed, resp.Code)
	umBody = make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
	request, _ = http.NewRequest(http.MethodConnect, "/comments", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusMethodNotAllowed, resp.Code)
	umBody = make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
	request, _ = http.NewRequest(http.MethodConnect, "/comments/1", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusMethodNotAllowed, resp.Code)
	umBody = make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
}
func TestListPosts(t *testing.T) {
	clearTablePosts()
	addPosts(10)
	request, _ := http.NewRequest(http.MethodGet, "/posts", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBody []models.Post
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if len(umBody) == 0 {
		t.Errorf("ListPosts:zero-array in response. Expected length = 10")
	}
	//with filter userId
	request, _ = http.NewRequest(http.MethodGet, "/posts?userId=1", nil)
	request.Header.Add("APIkey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if len(umBody) == 0 {
		t.Errorf("ListPosts with filter: zero-array in response. Expected length = 10")
	}
	//in xml format
	request, _ = http.NewRequest(http.MethodGet, "/posts?xml", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBodyXML models.Posts
	xml.Unmarshal(resp.Body.Bytes(), &umBodyXML)
	if len(umBodyXML.Posts) == 0 {
		t.Errorf("ListPost in xml: zero array in response. Expected length = 10")
	}
}
func TestListPostsErrors(t *testing.T) {
	clearTablePosts()
	addPosts(10)
	//nonexisten user
	request, _ := http.NewRequest(http.MethodGet, "/posts?userId=5", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	if body := resp.Body.String(); body != "[]" {
		t.Errorf("Expected empty array []. Got %s", body)
	}
	//nondigital userId
	request, _ = http.NewRequest(http.MethodGet, "/posts?userId=qwe", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusBadRequest, resp.Code)
	umBody := make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
}
func TestListComments(t *testing.T) {
	clearTableComments()
	clearTablePosts()
	addPosts(1)
	addComments(10, 1)
	request, _ := http.NewRequest(http.MethodGet, "/comments", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBody []models.Comment
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if len(umBody) == 0 {
		t.Errorf("ListPosts:zero-array in response. Expected length = 10")
	}
	//with filter postId
	request, _ = http.NewRequest(http.MethodGet, "/posts?postId=1", nil)
	request.Header.Add("APIkey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if len(umBody) == 0 {
		t.Errorf("ListPosts with filter: zero-array in response. Expected length = 10")
	}
	//in xml format
	request, _ = http.NewRequest(http.MethodGet, "/comments?xml", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBodyXML models.Comments
	xml.Unmarshal(resp.Body.Bytes(), &umBodyXML)
	if len(umBodyXML.Comments) == 0 {
		t.Errorf("ListPost in xml: zero array in response. Expected length = 10")
	}
}
func TestListCommentsError(t *testing.T) {
	clearTableComments()
	clearTablePosts()
	addPosts(1)
	addComments(10, 1)
	//nonexisten post
	request, _ := http.NewRequest(http.MethodGet, "/comments?postId=5", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	if body := resp.Body.String(); body != "[]" {
		t.Errorf("Expected empty array []. Got %s", body)
	}
	//nondigital postId
	request, _ = http.NewRequest(http.MethodGet, "/comments?postId=qwe", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusBadRequest, resp.Code)
	umBody := make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
}
func TestListPostComments(t *testing.T) {
	clearTableComments()
	clearTablePosts()
	addPosts(1)
	addComments(10, 1)
	request, _ := http.NewRequest(http.MethodGet, "/posts/1/comments", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBody []models.Comment
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if len(umBody) == 0 {
		t.Errorf("ListPosts:zero-array in response. Expected length = 10")
	}
	//in xml
	request, _ = http.NewRequest(http.MethodGet, "/posts/1/comments?xml", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBodyXML models.Comments
	xml.Unmarshal(resp.Body.Bytes(), &umBodyXML)
	if len(umBodyXML.Comments) == 0 {
		t.Errorf("ListPost in xml: zero array in response. Expected length = 10")
	}
}
func TestListPostCommentsErrors(t *testing.T) {
	clearTableComments()
	clearTablePosts()
	addPosts(1)
	addComments(10, 1)
	//nonexisten post
	request, _ := http.NewRequest(http.MethodGet, "/posts/5/comments", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	if body := resp.Body.String(); body != "[]" {
		t.Errorf("Expected empty array []. Got %s", body)
	}
	//nondigital postId
	request, _ = http.NewRequest(http.MethodGet, "/posts/qwe/comments", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusBadRequest, resp.Code)
	umBody := make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
}
func TestGetPost(t *testing.T) {
	clearTablePosts()
	addPosts(10)
	request, _ := http.NewRequest(http.MethodGet, "/posts/1", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBody models.Post
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if umBody.ID != 1 {
		t.Errorf("ListPosts: post.id expected 1. got %d", umBody.ID)
	}
	//in xml
	request, _ = http.NewRequest(http.MethodGet, "/posts/1?xml", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBodyXML models.Post
	xml.Unmarshal(resp.Body.Bytes(), &umBodyXML)
	if umBodyXML.ID != 1 {
		t.Errorf("ListPosts: post.id expected 1. got %d", umBodyXML.ID)
	}
}
func TestGetComment(t *testing.T) {
	clearTableComments()
	clearTablePosts()
	addPosts(1)
	addComments(10, 1)
	request, _ := http.NewRequest(http.MethodGet, "/comments/1", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBody models.Comment
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if umBody.ID != 1 {
		t.Errorf("ListPosts: comment.id expected 1. got %d", umBody.ID)
	}
	//in xml
	request, _ = http.NewRequest(http.MethodGet, "/comments/1?xml", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var umBodyXML models.Comment
	xml.Unmarshal(resp.Body.Bytes(), &umBodyXML)
	if umBodyXML.ID != 1 {
		t.Errorf("ListPosts: comment.id expected 1. got %d", umBodyXML.ID)
	}
}
func TestCreatePost(t *testing.T) {
	clearTablePosts()
	rBody := []byte(`{"title":"test","body":"test"}`)
	request, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(rBody))
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusCreated, resp.Code)
	var m map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &m)
	if m["userId"] != 1.0 {
		t.Errorf("Expected post userId to be '1'. Got '%v'", m["userId"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected post id to be '1'. Got '%v'", m["id"])
	}
	if m["title"] != "test" {
		t.Errorf("Expected post title to be 'test'. Got '%v'", m["title"])
	}
	if m["body"] != "test" {
		t.Errorf("Expected post body to be 'test'. Got '%v'", m["body"])
	}
}
func TestCreateComment(t *testing.T) {
	clearTableComments()
	clearTablePosts()
	addPosts(1)
	rBody := []byte(`{"name":"test","body":"test","email":"test@test.test","postId":1}`)
	request, _ := http.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(rBody))
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusCreated, resp.Code)
	var m map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &m)
	if m["userId"] != 1.0 {
		t.Errorf("Expected comment userId to be '1'. Got '%v'", m["userId"])
	}
	if m["postId"] != 1.0 {
		t.Errorf("Expected comment postId to be '1'. Got '%v'", m["postId"])
	}
	if m["id"] != 1.0 {
		t.Errorf("Expected comment id to be '1'. Got '%v'", m["id"])
	}
	if m["name"] != "test" {
		t.Errorf("Expected comment name to be 'test'. Got '%v'", m["name"])
	}
	if m["body"] != "test" {
		t.Errorf("Expected comment body to be 'test'. Got '%v'", m["body"])
	}
	if m["email"] != "test@test.test" {
		t.Errorf("Expected comment email to be 'test'. Got '%v'", m["email"])
	}
}

func TestUpdatePost(t *testing.T) {
	clearTablePosts()
	addPosts(1)
	request, _ := http.NewRequest(http.MethodGet, "/posts/1", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	var origin models.Post
	json.Unmarshal(resp.Body.Bytes(), &origin)
	reqData := []byte(`{"id":1, "title":"test update", "body":"test update"}`)
	request, _ = http.NewRequest(http.MethodPut, "/posts", bytes.NewBuffer(reqData))
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var update models.Post
	json.Unmarshal(resp.Body.Bytes(), &update)
	if origin.ID != update.ID {
		t.Errorf("Origin id: %d and updated id: %d mismutch", origin.ID, update.ID)
	}
	if origin.Title == update.Title {
		t.Errorf("Expected the title to change from '%v' to 'test update'. Got '%v'", origin.Title, update.Title)
	}
	if origin.Body == update.Body {
		t.Errorf("Expected the body to change from '%v' to 'test update'. Got '%v'", origin.Body, update.Body)
	}
}
func TestUpdateComment(t *testing.T) {
	clearTableComments()
	clearTablePosts()
	addPosts(1)
	addComments(1, 1)
	request, _ := http.NewRequest(http.MethodGet, "/comments/1", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	var origin models.Comment
	json.Unmarshal(resp.Body.Bytes(), &origin)
	reqData := []byte(`{"id":1, "postId":1 , "name":"test update", "body":"test update", "email":"testupd@test.test"}`)
	request, _ = http.NewRequest(http.MethodPut, "/comments", bytes.NewBuffer(reqData))
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	var update models.Comment
	json.Unmarshal(resp.Body.Bytes(), &update)
	if origin.ID != update.ID {
		t.Errorf("Origin id: %d and updated id: %d mismutch", origin.ID, update.ID)
	}
	if origin.Name == update.Name {
		t.Errorf("Expected the name to change from '%v' to 'test update'. Got '%v'", origin.Name, update.Name)
	}
	if origin.Body == update.Body {
		t.Errorf("Expected the body to change from '%v' to 'test update'. Got '%v'", origin.Body, update.Body)
	}
	if origin.Email == update.Email {
		t.Errorf("Expected the email to change from '%v' to 'test update'. Got '%v'", origin.Body, update.Body)
	}
}
func TestDeletePost(t *testing.T) {
	clearTablePosts()
	addPosts(1)
	request, _ := http.NewRequest(http.MethodDelete, "/posts/1", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	request, _ = http.NewRequest(http.MethodGet, "/posts/1", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusNotFound, resp.Code)
	umBody := make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
}
func TestDeleteComment(t *testing.T) {
	clearTableComments()
	clearTablePosts()
	addPosts(1)
	addComments(1, 1)
	request, _ := http.NewRequest(http.MethodDelete, "/comments/1", nil)
	request.Header.Add("APIKey", "test")
	resp := execRequest(request)
	checkRespCode(t, http.StatusOK, resp.Code)
	request, _ = http.NewRequest(http.MethodGet, "/comments/1", nil)
	request.Header.Add("APIKey", "test")
	resp = execRequest(request)
	checkRespCode(t, http.StatusNotFound, resp.Code)
	umBody := make(map[string]interface{})
	json.Unmarshal(resp.Body.Bytes(), &umBody)
	if _, ok := umBody["error"]; !ok {
		t.Errorf("Expected error JSON message. Got %s", resp.Body.String())
	}
}
