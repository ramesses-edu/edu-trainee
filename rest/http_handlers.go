package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {

}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	rPath := r.URL.Path
	rePostsComments := regexp.MustCompile(`^\/posts\/\d+\/comments(\/)??$`)
	rePostsID := regexp.MustCompile(`^\/posts\/\d+(\/)??$`)
	rePosts := regexp.MustCompile(`^\/posts(\/)??$`)

	var param map[string]interface{} = make(map[string]interface{})
	respXML := false
	for key := range r.Form {
		if key == "xml" {
			respXML = true
		}
	}

	switch {
	case rePosts.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: //list posts with filters
			userId := r.FormValue("userId")
			if userId != "" {
				var err error
				param["userId"], err = strconv.Atoi(r.FormValue("userId"))
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			var pp posts
			result := pp.getPosts(a.db, param)
			if result.Error != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if respXML {
				pp.responseXML(w, r)
			} else {
				pp.responseJSON(w, r)
			}
		case http.MethodPost: //create post in:json
			reqBody, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var p post
			err = json.Unmarshal(reqBody, &p)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			result := p.createPost(a.db)
			if result.Error != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"id": %d}`, p.ID)
		case http.MethodPut: //update post  in:json
			reqBody, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var p post
			err = json.Unmarshal(reqBody, &p)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			result := p.updatePost(a.db)
			if result.Error != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"id": %d}`, p.ID)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case rePostsID.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: // get posts/{id}
			var err error
			param["id"], err = strconv.Atoi(reNum.FindString(rPath))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var pp posts
			result := pp.getPosts(a.db, param)
			if result.Error != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if respXML {
				pp.responseXML(w, r)
			} else {
				pp.responseJSON(w, r)
			}
		case http.MethodDelete: // delete posts/{id}
			pID, err := strconv.Atoi(reNum.FindString(rPath))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var p post = post{ID: pID}
			result := p.deletePost(a.db)
			if result.Error != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case rePostsComments.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: // list comments like->/comments?postId={id}
			var err error
			param["postId"], err = strconv.Atoi(reNum.FindString(rPath))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var cc comments
			result := cc.getComments(a.db, param)
			if result.Error != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if respXML {
				cc.responseXML(w, r)
			} else {
				cc.responseJSON(w, r)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "[]")
	}
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	rPath := r.URL.Path
	reCommentsID := regexp.MustCompile(`^\/comments\/\d+(\/)??$`)
	reComments := regexp.MustCompile(`^\/comments(\/)??$`)

	respXML := false
	var param map[string]interface{} = make(map[string]interface{})
	for key := range r.Form {
		if key == "xml" {
			respXML = true
		}
	}

	switch {
	case reComments.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: // list comments with filters
			postId := r.FormValue("postId")
			if postId != "" {
				var err error
				param["postId"], err = strconv.Atoi(r.FormValue("postId"))
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			var cc comments
			result := cc.getComments(a.db, param)
			if result.Error != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if respXML {
				cc.responseXML(w, r)
			} else {
				cc.responseJSON(w, r)
			}
		case http.MethodPost: // create comment in:json
			reqBody, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var c comment
			err = json.Unmarshal(reqBody, &c)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			result := c.createComment(a.db)
			if result.Error != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"id": %d}`, c.ID)
		case http.MethodPut: // update comment in:json
			reqBody, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var c comment
			err = json.Unmarshal(reqBody, &c)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			result := c.updateComment(a.db)
			if result.Error != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"id": %d}`, c.ID)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case reCommentsID.Match([]byte(rPath)):
		switch r.Method {
		case http.MethodGet: // get comments/{id}
			var err error
			param["id"], err = strconv.Atoi(reNum.FindString(rPath))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var cc comments
			result := cc.getComments(a.db, param)
			if result.Error != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if respXML {
				cc.responseXML(w, r)
			} else {
				cc.responseJSON(w, r)
			}
		case http.MethodDelete: // delete comments/{id}
			cID, err := strconv.Atoi(reNum.FindString(rPath))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var c comment = comment{ID: cID}
			result := c.deleteComment(a.db)
			if result.Error != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "[]")
	}
}
