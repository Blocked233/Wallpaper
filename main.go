package main

import (
	"gee"
	"log"
	"net/http"

	"github.com/nutsdb/nutsdb"
)

func main() {
	// open db and access bing api
	go readJson()

	r := gee.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	r.GET("/", func(ctx *gee.Context) {
		http.ServeFile(ctx.Writer, ctx.Req, "./static/html/index.html")
	})

	auth := r.Group("/auth")
	{
		bucket := "account"

		auth.POST("/register", func(ctx *gee.Context) {

			if err := db.Update(
				func(tx *nutsdb.Tx) error {
					ctx.Req.ParseForm()
					log.Println(ctx.Req.Form)
					log.Println(ctx.PostForm("username"), ctx.PostForm("password"))

					if err := tx.Put(bucket, []byte(ctx.PostForm("username")), []byte(ctx.PostForm("password")), 0); err != nil {
						return err
					}

					ctx.String(http.StatusOK, "register success")

					return nil
				}); err != nil {
				log.Println(err)
			}

		})

		auth.POST("/login", func(ctx *gee.Context) {

			if err := db.Update(
				func(tx *nutsdb.Tx) error {

					password, err := tx.Get(bucket, []byte(ctx.PostForm("username")))
					if err != nil {
						return err
					}

					if string(password.Value) == ctx.PostForm("password") {
						ctx.String(http.StatusOK, "login success")
					}

					return nil
				}); err != nil {
				log.Println(err)
			}

		})

	}

	if err := r.Run(":9999"); err != nil {
		log.Println(err)
	}
}
