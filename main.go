package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
	"wallpaper/cache"
	"wallpaper/cosmosdb"
	"wallpaper/qiniu"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	br "github.com/Blocked233/middleware/brotli"
	"github.com/Blocked233/middleware/tunnel"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

var (
	client *azcosmos.Client
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

	// Must Add funMap before load template
	r.FuncMap["escape"] = template.HTMLEscapeString
	r.LoadHTMLGlob("templates/*")

	r.Static("/assets", "./static")

	r.GET("/", func(ctx *gin.Context) {

		params := &wallpaper{TimeURL: make(map[string]string, 31)}

		// update parms.time

		updateMonth(params)

		// get all pictures of this month

		partitionKey := time.Now().Format("200601")
		query := fmt.Sprintf("SELECT * FROM c WHERE c.Month = '%s'", partitionKey)
		results, err := cosmosdb.QueryWallpaperItems(client, "bingWallpaper", "US", partitionKey, query)
		if err != nil {
			log.Printf("queryItems failed: %s\n", err)
		}

		params.HeadImgUrl = results[0].URL
		params.HeadImgCopyright = results[0].Copyright
		for _, item := range results {
			params.TimeURL[item.ID] = qiniu.Key2PublicUrl(item.ID + ".jpg")
		}

		ctx.HTML(200, "bingTemplate.html", params)

	})

	r.POST("/Message/Tun", func(ctx *gin.Context) {
		tunnel.GrpcServer.ServeHTTP(ctx.Writer, ctx.Request)
	})

	cachefile := r.Group("/cachefile")
	{
		cachefile.GET("/webp", func(ctx *gin.Context) {
			b, err := cache.Webp.Get(ctx.Query("key"))
			if err != nil {
				ctx.String(404, "Not Found", err)
				return
			}
			ctx.Data(200, "image/webp", b)

		})
	}

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
