package main

import (
	"context"
	"log"
	"os"

	"edu-trainee/rest/application"
	"edu-trainee/rest/authorization"
	"edu-trainee/rest/docs"
	"edu-trainee/rest/echofeature"

	"edu-trainee/rest/httphandlers"

	"github.com/joho/godotenv"
)

var a application.Application

func init() {
	// loads values from .env into the system
	if err := godotenv.Load("./config.env"); err != nil {
		log.Print("No .env file found")
		os.Exit(1)
	}
}

// @title Education Forum API
// @version 1.0
// @description This is a education forum server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email romgrishin@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:80
// @BasePath /
// @query.collection.format multi

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name APIKey
func main() {

	a = application.Application{}
	a.InitApplication()
	sql, err := a.DB.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sql.Close()
	httphandlers.InitRoutes(a.Router, a.DB)
	authorization.A = a
	httphandlers.DB = a.DB

	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	//echo
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go echofeature.EchoStart(ctx, "localhost:8081")

	a.ListenAndServe()
}
