package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dinever/golf"
	"github.com/tormaroe/cublog/posts"
)

var allPosts = []*posts.BlogPost{}

func mainHandler(ctx *golf.Context) {
	title, _ := ctx.App.Config.GetString("blogName", "The Blog")
	data := map[string]interface{}{
		"Title": title,
		"Posts": allPosts,
	}
	ctx.Loader("template").Render("index.html", data)
}

func findPost(path string) (*posts.BlogPost, error) {
	for i := range allPosts {
		if allPosts[i].Path == path {
			return allPosts[i], nil
		}
	}
	return nil, errors.New("Post not found")
}

func pageHandler(ctx *golf.Context) {
	post, err := findPost(ctx.Param("page"))
	if err != nil {
		ctx.Abort(404)
		return
	}

	data := map[string]interface{}{
		"Title": post.Title,
		"Body":  post.Body,
	}
	ctx.Loader("template").Render("post.html", data)
}

func main() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}

	allPosts, err = posts.LoadAll()
	if err != nil {
		panic(err)
	}
	for _, p := range allPosts {
		fmt.Println("Loaded post: " + p.Title)
	}

	app := golf.New()
	app.Config, err = golf.ConfigFromJSON(file)
	app.View.SetTemplateLoader("template", "www/templates/")
	app.Static("/www/static/", "static")
	app.Get("/", mainHandler)
	app.Get("/p/:page/", pageHandler)
	app.Run(":9000")
}
