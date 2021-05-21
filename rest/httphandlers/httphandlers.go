package httphandlers

import (
	"edu-trainee/rest/application"
	"edu-trainee/rest/authorization"
	"edu-trainee/rest/middleware"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"text/template"

	httpSwagger "github.com/swaggo/http-swagger"
)

var (
	reNum *regexp.Regexp = regexp.MustCompile(`\d+`)
)

func InitRoutes(router *http.ServeMux) {
	router.Handle("/", mainHandler())
	router.Handle("/public", http.NotFoundHandler())
	router.Handle("/public/", publicHandler())
	router.Handle("/logout/", logoutHandler())
	router.Handle("/auth/", authentification())
	router.Handle("/posts", middleware.Autorization(http.HandlerFunc(postsHandler)))
	router.Handle("/posts/", middleware.Autorization(http.HandlerFunc(postsHandler)))
	router.Handle("/comments", middleware.Autorization(http.HandlerFunc(commentsHandler)))
	router.Handle("/comments/", middleware.Autorization(http.HandlerFunc(commentsHandler)))
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
	DB := application.CurrentApplication().DB
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
	DB := application.CurrentApplication().DB
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
	if _, ok := r.Form["xml"]; ok {
		return true
	}
	return false
}

func xmlWrite(w http.ResponseWriter, data interface{}) error {
	xmlB, err := xml.MarshalIndent(data, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return err
	}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(xmlB))
	return nil
}

func jsonWrite(w http.ResponseWriter, data interface{}) error {
	jsonB, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":""}`))
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonB))
	return nil
}
