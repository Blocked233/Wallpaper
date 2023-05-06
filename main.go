package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"time"
	"wallpaper/cache"
	"wallpaper/config"
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

	config := config.GetConfig()

	cache.NewRedisClient(config.RedisPassword)
	qiniu.NewQiniu(config.Domin, config.AccessKey, config.SecretKey)

	client, err = cosmosdb.NewClient(config.Endpoint, config.Key)
	if err != nil {
		log.Fatal("Failed to create Azure Cosmos DB db client: ", err)
	}

	// Create a database
	err = cosmosdb.CreateDatabase(client, databaseName)
	if err != nil {
		log.Printf("createDatabase failed: %s\n", err)
	}

	// Create different containers
	err = cosmosdb.CreateContainer(client, databaseName, "US", "/Month")
	if err != nil {
		log.Printf("createContainer failed: %s\n", err)
	}

	err = cosmosdb.CreateContainer(client, databaseName, "Account", "/Username")
	if err != nil {
		panic(err)
	}

	// cache group init
	cache.NewWallPaperGroup()
}

func main() {

	go update()

	r := gin.Default()
	r.Use(br.Brotli(6))

	// Must Add funMap before load template
	r.FuncMap["escape"] = template.HTMLEscapeString
	r.LoadHTMLGlob("templates/*")

	r.Static("/assets", "./static")

	r.POST("/Message/Tun", func(ctx *gin.Context) {
		tunnel.GrpcServer.ServeHTTP(ctx.Writer, ctx.Request)
	})

	r.GET("/", func(ctx *gin.Context) {

		params := &wallpaperParams{TimeURL: make(map[string]string, 31)}

		// update parms.time

		updateMonth(time.Now(), params)

		// get all pictures of this month

		results := getWallpapers(time.Now())

		if len(results) == 0 {
			ctx.AbortWithStatus(404)
		}
		params.HeadImgUrl = results[0].URL
		params.HeadImgCopyright = results[0].Copyright
		for _, item := range results {
			params.TimeURL[item.ID] = qiniu.Key2PublicUrl(item.ID + ".jpg")
		}

		ctx.HTML(200, "bingTemplate.html", params)

	})

	debug := r.Group("/debug")
	{
		debug.GET("/pprof", func(ctx *gin.Context) {
			pprof.Index(ctx.Writer, ctx.Request)
		})
		debug.GET("/pprof/cmdline", func(ctx *gin.Context) {
			pprof.Cmdline(ctx.Writer, ctx.Request)
		})
		debug.GET("/pprof/profile", func(ctx *gin.Context) {
			pprof.Profile(ctx.Writer, ctx.Request)
		})
		debug.GET("/pprof/symbol", func(ctx *gin.Context) {
			pprof.Symbol(ctx.Writer, ctx.Request)
		})

		debug.GET("/pprof/trace", func(ctx *gin.Context) {
			pprof.Trace(ctx.Writer, ctx.Request)
		})
	}

	wallpaper := r.Group("/wallpaper")
	{
		wallpaper.GET("/date", func(ctx *gin.Context) {

			d := ctx.Query("time")
			if d == "" {
				ctx.String(404, "Not Found")
				return
			}

			t, err := time.Parse("2006-01", d)
			if err != nil {
				ctx.String(404, "Not Found")
				return
			}

			params := &wallpaperParams{TimeURL: make(map[string]string, 31)}

			updateMonth(t, params)
			t = t.AddDate(0, 1, -1)

			results := getWallpapers(t)
			if len(results) == 0 {
				ctx.AbortWithStatus(404)
			}

			params.HeadImgUrl = results[0].URL
			params.HeadImgCopyright = results[0].Copyright
			for _, item := range results {
				params.TimeURL[item.ID] = qiniu.Key2PublicUrl(item.ID + ".jpg")
			}

			ctx.HTML(200, "bingTemplate.html", params)
		})
	}

	cachefile := r.Group("/cachefile")
	{
		cachefile.GET("/webp", func(ctx *gin.Context) {
			b, err := cache.Wallpaper.Get(ctx.Query("key"))
			if err != nil {
				ctx.String(404, "Not Found", err)
				return
			}

			result := cosmosdb.WallpaperItem{}
			json.Unmarshal(b, &result)
			ctx.Data(200, "image/webp", result.Webp)

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

//go env -w GOOS=linux
//go build -ldflags="-s -w" ./
//go tool pprof -http=:8088 https://graphs.eu.org/debug/pprof/profile -seconds 20
