package main

import (
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/go-resty/resty"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
)

type Med struct {
	Title         string
	SetId         string
	PublishedDate string
}

func extractString(value []byte, key string) string {
	ans, _ := jsonparser.GetString(value, key)
	return ans
}

func main() {
	e := echo.New()
	e.Debug = true

	e.GET("/fetch", handleFetch)
	e.GET("/", handleDisplay)
	e.Logger.Fatal(e.Start(":3000"))
}

func handleDisplay(c echo.Context) error {
	db, err := gorm.Open("sqlite3", "med.db")
	if err != nil {
		panic("failed to connect database")
	}

	defer db.Close()
	var meds []Med
	db.Table("meds").Select("*").Scan(&meds)

	return c.JSON(http.StatusOK, &meds)
}

func handleFetch(c echo.Context) error {
	client := resty.New()
	resp, _ := client.R().
		EnableTrace().
		Get("https://dailymed.nlm.nih.gov/dailymed/services/v2/spls.json")

	var data []byte = resp.Body()

	var Meds []Med

	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		Meds = append(Meds, Med{
			Title:         extractString(value, "title"),
			SetId:         extractString(value, "setid"),
			PublishedDate: extractString(value, "published_date"),
		})

	}, "data")

	db, err := gorm.Open("sqlite3", "med.db")
	if err != nil {
		panic("failed to connect database")
	}

	defer db.Close()

	db.DropTableIfExists(&Med{})
	db.AutoMigrate(&Med{})

	for i := 0; i < len(Meds); i++ {
		db.Create(&Meds[i])
	}

	return c.String(http.StatusOK, "Fetched Data Successfully!")
}
