package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"wallpaper/cosmosdb"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type Image struct {
	Startdate     string        `json:"startdate"`
	Fullstartdate string        `json:"fullstartdate"`
	Enddate       string        `json:"enddate"`
	URL           string        `json:"url"`
	Urlbase       string        `json:"urlbase"`
	Copyright     string        `json:"copyright"`
	Copyrightlink string        `json:"copyrightlink"`
	Title         string        `json:"title"`
	Quiz          string        `json:"quiz"`
	Wp            bool          `json:"wp"`
	Hsh           string        `json:"hsh"`
	Drk           int           `json:"drk"`
	Top           int           `json:"top"`
	Bot           int           `json:"bot"`
	Hs            []interface{} `json:"hs"`
}

// bing api json struct
type AutoGenerated struct {
	Images   []Image `json:"images"`
	Tooltips struct {
		Loading  string `json:"loading"`
		Previous string `json:"previous"`
		Next     string `json:"next"`
		Walle    string `json:"walle"`
		Walls    string `json:"walls"`
	} `json:"tooltips"`
}

// template variables struct
type wallpaper struct {
	Time             [12]string // Page Tail Month
	HeadImgUrl       string
	HeadImgCopyright string
	TimeURL          map[string]string // Time and Download URL
}

var (
	databaseName = "bingWallpaper"
	partitionKey = "/Month"
	client       *azcosmos.Client

	bingURL         = "https://www.bing.com"
	wallpaperParams = &wallpaper{TimeURL: make(map[string]string, 31)}
)

const apiAddr = "https://www.bing.com/HPImageArchive.aspx?format=js&n=15&pid=hp&mkt=en-US&uhd=1&uhdwidth=384&uhdheight=216"

func update() {

	for {

		resp, err := http.Get(apiAddr)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Hour)
			continue
		}

		jsonmsg := AutoGenerated{}
		data, _ := io.ReadAll(resp.Body)
		json.Unmarshal(data, &jsonmsg)

		// error handling
		if len(jsonmsg.Images) == 0 {
			time.Sleep(time.Hour)
			continue
		}

		wallpaperParams.HeadImgUrl = bingURL + jsonmsg.Images[0].Urlbase + "_UHD.jpg"
		wallpaperParams.HeadImgCopyright = jsonmsg.Images[0].Copyright

		for _, val := range jsonmsg.Images {
			upload2DB(val)
		}

		updateHTML()

		time.Sleep(24 * time.Hour)
	}
}

func upload2DB(val Image) {

	// Create a Item
	item := cosmosdb.WallpaperItem{
		ID:        val.Enddate,     // 20220216
		Month:     val.Enddate[:6], // 202202
		Copyright: val.Copyright,
		URL:       bingURL + val.Urlbase + "_UHD.jpg",
	}

	// upload to cosmosdb

	err := cosmosdb.CreateItem(client, databaseName, "US", item.Month, item)
	if err != nil {
		log.Printf("createItem failed: %s\n", err)
	}
}

func updateMonth() {
	year, month, _ := time.Now().Date()

	for i := 0; i < 12; i++ {

		wallpaperParams.Time[i] = fmt.Sprintf("%d-%02d", year, month)

		month = month - 1

		if month == 0 {
			year = year - 1
			month = 12
		}

	}
}

func updateHTML() {

	// update Page Tail Month

	updateMonth()

	// get all pictures of this month

	partitionKey := time.Now().Format("200601")
	query := fmt.Sprintf("SELECT * FROM c WHERE c.Month = '%s'", partitionKey)
	results, err := cosmosdb.QueryWallpaperItems(client, "bingWallpaper", "US", partitionKey, query)
	if err != nil {
		log.Printf("queryItems failed: %s\n", err)
	}

	for _, item := range results {
		wallpaperParams.TimeURL[item.ID] = item.URL
	}

	// template

	indexHTML, err := template.ParseFiles("./templates/bingTemplate.html")
	if err != nil {
		log.Println(err)
		return
	}

	// daily html
	out, err := os.Create("./static/html/" + wallpaperParams.Time[0] + ".html")
	if err != nil {
		log.Println(err)
		return
	}
	defer out.Close()

	err = indexHTML.Execute(out, wallpaperParams)
	if err != nil {
		log.Println(err)
		return
	}

	// index html
	index, err := os.Create("./static/html/index.html")
	if err != nil {
		log.Println(err)
		return
	}
	defer out.Close()

	err = indexHTML.Execute(index, wallpaperParams)
	if err != nil {
		log.Println(err)
		return
	}

	// register html
	registerHTML, err := template.ParseFiles("./templates/registerTemplate.html")
	if err != nil {
		log.Println(err)
		return
	}

	register, err := os.Create("./static/html/register.html")
	if err != nil {
		log.Println(err)
		return
	}
	defer out.Close()

	err = registerHTML.Execute(register, wallpaperParams)
	if err != nil {
		log.Println(err)
		return
	}

	// login html
	loginHTML, err := template.ParseFiles("./templates/loginTemplate.html")
	if err != nil {
		log.Println(err)
		return
	}

	login, err := os.Create("./static/html/login.html")
	if err != nil {
		log.Println(err)
		return
	}
	defer out.Close()

	err = loginHTML.Execute(login, wallpaperParams)
	if err != nil {
		log.Println(err)
		return
	}

}
