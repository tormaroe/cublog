package main

import (
	"os"

	"github.com/dinever/golf"
)

func mainHandler(ctx *golf.Context) {
	title, _ := ctx.App.Config.GetString("blogName", "The Blog")
	data := map[string]interface{}{
		"Title": title,
	}
	ctx.Loader("template").Render("index.html", data)
}

func pageHandler(ctx *golf.Context) {
	ctx.Send("Page: " + ctx.Param("page"))
}

func main() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}

	app := golf.New()
	app.Config, err = golf.ConfigFromJSON(file)
	app.View.SetTemplateLoader("template", "templates/")
	app.Static("/static/", "static")
	app.Get("/", mainHandler)
	app.Get("/p/:page/", pageHandler)
	app.Run(":9000")
}
