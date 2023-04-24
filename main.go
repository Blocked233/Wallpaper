package main

import (
	"log"
	"net/http"
	"os"
	"time"
	"wallpaper/cosmosdb"

	br "github.com/Blocked233/middleware/brotli"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

func init() {

	var err error

	client, err = cosmosdb.NewClient()
	if err != nil {
		log.Fatal("Failed to create Azure Cosmos DB db client: ", err)
	}

	// Create a database
	err = cosmosdb.CreateDatabase(client, databaseName)
	if err != nil {
		log.Printf("createDatabase failed: %s\n", err)
	}

	// Create different containers
	err = cosmosdb.CreateContainer(client, databaseName, "US", partitionKey)
	if err != nil {
		log.Printf("createContainer failed: %s\n", err)
	}

	err = cosmosdb.CreateContainer(client, databaseName, "Account", "/Username")
	if err != nil {
		panic(err)
	}
}

func main() {

	go update()

	r := gin.Default()
	r.Use(gin.Logger(), gin.Recovery(), br.Brotli(6))

	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	r.GET("/", func(ctx *gin.Context) {
		http.ServeFile(ctx.Writer, ctx.Request, "./static/html/index.html")
	})

	auth := r.Group("/auth")
	{

		auth.POST("/register", func(ctx *gin.Context) {
			err := accountRegister(ctx.PostForm("username"), ctx.PostForm("password"), ctx.PostForm("email"))
			if err != nil {
				ctx.JSON(200, gin.H{
					"status":  "error",
					"message": err.Error(),
				})
				return
			}

			ctx.JSON(200, gin.H{
				"status":  "success",
				"message": "Register success",
			})

			time.Sleep(1 * time.Second)
			ctx.Redirect(http.StatusMovedPermanently, "/")
		})

		auth.POST("/login", func(ctx *gin.Context) {

			err := accountLogin(ctx.PostForm("username"), ctx.PostForm("password"))
			if err != nil {
				ctx.JSON(200, gin.H{
					"status":  "error",
					"message": err.Error(),
				})
				return
			}

			ctx.JSON(200, gin.H{
				"status":  "success",
				"message": "Login success",
			})

			time.Sleep(1 * time.Second)
			ctx.Redirect(http.StatusMovedPermanently, "/")

		})

	}

	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(os.Args...),
		Cache:      autocert.DirCache("secret-dir"),
	}

	// 80 and 443
	log.Fatal(autotls.RunWithManager(r, &m))

}
