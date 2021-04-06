package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	echo "github.com/labstack/echo/v4"
)

func echoStart(ctx context.Context, addr string) {
	e := echo.New()
	initRoutes(e)
	e.Start(addr)
	<-ctx.Done()
	e.Shutdown(ctx)
}
func initRoutes(e *echo.Echo) {
	e.GET("/*", func(c echo.Context) error {
		return c.String(http.StatusBadRequest, "")
	})
	gPosts := e.Group("/posts")
	gPosts.Use(emwAutorization)
	gPosts.GET("", listPosts)
	gPosts.GET("/:id", getPostById)
	gPosts.GET("/:id/comments", listPostComments)
	gPosts.POST("", createPost)
	gPosts.PUT("", updatePost)
	gPosts.DELETE("/:id", deletePost)

	gComments := e.Group("/comments")
	gComments.Use(emwAutorization)
	gComments.GET("", listComments)
	gComments.GET("/:id", getCommentByID)
	gComments.POST("", createComment)
	gComments.PUT("", updateComment)
	gComments.DELETE("/:id", deleteComment)
}
func emwAutorization(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		accessToken := ""
		accessTokenCookie, err := c.Cookie("UAAT")
		if err == nil {
			accessToken = accessTokenCookie.Value
		}
		if accessToken == "" {
			accessToken = c.Request().Header.Get("APIKey")
		}
		if accessToken == "" {
			c.String(http.StatusNetworkAuthenticationRequired, `{"error":""}`)
			return nil
		}
		hashAccTok := calculateSignature(accessToken, "provider")
		var u user
		result := u.getUser(a.db, map[string]interface{}{
			"access_token": hashAccTok,
		})
		if result.Error != nil || result.RowsAffected == 0 {
			c.String(http.StatusNetworkAuthenticationRequired, `{"error":""}`)
			return nil
		}
		return next(c)
	}
}

func listPosts(c echo.Context) error {
	respXML := false
	for key := range c.QueryParams() {
		if key == "xml" {
			respXML = true
		}
	}
	var param map[string]interface{} = make(map[string]interface{})
	if userId := c.QueryParam("userId"); userId != "" {
		var err error
		param["userId"], err = strconv.Atoi(userId)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "")
		}
	}
	var pp posts
	result := pp.listPosts(a.db, param)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if respXML {
		return c.XML(http.StatusOK, &pp)
	}
	return c.JSON(http.StatusOK, &pp.Posts)
}
func getPostById(c echo.Context) error {
	respXML := false
	for key := range c.QueryParams() {
		if key == "xml" {
			respXML = true
		}
	}
	var param map[string]interface{} = make(map[string]interface{})
	var err error
	param["id"], err = strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "")
	}
	var p post
	result := p.getPost(a.db, param)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	var pp posts = posts{Posts: []post{p}}
	if respXML {
		return c.XML(http.StatusOK, &pp)
	}
	return c.JSON(http.StatusOK, &pp.Posts)
}
func listPostComments(c echo.Context) error {
	respXML := false
	for key := range c.QueryParams() {
		if key == "xml" {
			respXML = true
		}
	}
	var param map[string]interface{} = make(map[string]interface{})
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	param["postId"] = postId
	var cc comments
	result := cc.listComments(a.db, param)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if respXML {
		return c.XML(http.StatusOK, &cc)
	}
	return c.JSON(http.StatusOK, &cc.Comments)
}
func createPost(c echo.Context) error {
	reqBody, err := ioutil.ReadAll(c.Request().Body)
	defer c.Request().Body.Close()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var p post
	err = json.Unmarshal(reqBody, &p)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	result := p.createPost(a.db)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	return c.String(http.StatusCreated, fmt.Sprintf(`{"id": %d}`, p.ID))
}
func updatePost(c echo.Context) error {
	reqBody, err := ioutil.ReadAll(c.Request().Body)
	defer c.Request().Body.Close()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var p post
	err = json.Unmarshal(reqBody, &p)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	result := p.updatePost(a.db)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, fmt.Sprintf(`{"id": %d}`, p.ID))
}
func deletePost(c echo.Context) error {
	pID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var p post = post{ID: pID}
	result := p.deletePost(a.db)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.String(http.StatusOK, "")
}
func listComments(c echo.Context) error {
	respXML := false
	for key := range c.QueryParams() {
		if key == "xml" {
			respXML = true
		}
	}
	var param map[string]interface{} = make(map[string]interface{})
	if postId := c.QueryParam("postId"); postId != "" {
		var err error
		param["postId"], err = strconv.Atoi(postId)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "")
		}
	}
	var cc comments
	result := cc.listComments(a.db, param)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if respXML {
		return c.XML(http.StatusOK, &cc)
	}
	return c.JSON(http.StatusOK, &cc.Comments)
}
func getCommentByID(c echo.Context) error {
	respXML := false
	for key := range c.QueryParams() {
		if key == "xml" {
			respXML = true
		}
	}
	var param map[string]interface{} = make(map[string]interface{})
	var err error
	param["id"], err = strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "")
	}
	var cmnt comment
	result := cmnt.getComment(a.db, param)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	var cc comments = comments{Comments: []comment{cmnt}}
	if respXML {
		return c.XML(http.StatusOK, &cc)
	}
	return c.JSON(http.StatusOK, &cc.Comments)
}
func createComment(c echo.Context) error {
	reqBody, err := ioutil.ReadAll(c.Request().Body)
	defer c.Request().Body.Close()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var cmt comment
	err = json.Unmarshal(reqBody, &cmt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	result := cmt.createComment(a.db)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	return c.String(http.StatusCreated, fmt.Sprintf(`{"id": %d}`, cmt.ID))
}
func updateComment(c echo.Context) error {
	reqBody, err := ioutil.ReadAll(c.Request().Body)
	defer c.Request().Body.Close()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var cmt comment
	err = json.Unmarshal(reqBody, &cmt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	result := cmt.updateComment(a.db)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, fmt.Sprintf(`{"id": %d}`, cmt.ID))
}
func deleteComment(c echo.Context) error {
	cID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var cmt comment = comment{ID: cID}
	result := cmt.deleteComment(a.db)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.String(http.StatusOK, "")
}
