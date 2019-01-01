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

func adminHandler(ctx *golf.Context) {
	data := map[string]interface{}{
		"Title": "Admin",
		"Posts": allPosts,
	}
	ctx.Loader("template").Render("admin.html", data)
}

func newPostHandler(ctx *golf.Context) {
	data := map[string]interface{}{
		"Title": "New post",
	}
	ctx.Loader("template").Render("post-form.html", data)
}

func insertPostHandler(ctx *golf.Context) {
	err := ctx.Request.ParseForm()
	if err != nil {
		panic(err)
	}
	post := posts.New(ctx.Request.FormValue("PostTitle"), ctx.Request.FormValue("PostSlug"), ctx.Request.FormValue("PostBody"))
	if err = post.Save(); err != nil {
		panic(err)
	}
	allPosts = append(allPosts, post)
	ctx.Redirect("/admin")
}

func editPostHandler(ctx *golf.Context) {
	post, err := findPost(ctx.Param("page"))
	if err != nil {
		ctx.Abort(404)
		return
	}

	data := map[string]interface{}{
		"Title": "Edit: " + post.Title,
		"Post":  post,
	}
	ctx.Loader("template").Render("post-form.html", data)
}

func updatePostHandler(ctx *golf.Context) {
	err := ctx.Request.ParseForm()
	if err != nil {
		panic(err)
	}
	post, err := findPost(ctx.Param("page"))
	if err != nil {
		ctx.Abort(404)
		return
	}
	post.Title = ctx.Request.FormValue("PostTitle")
	post.Path = ctx.Request.FormValue("PostSlug")
	post.Body = ctx.Request.FormValue("PostBody")
	post.Approved = false
	if err = post.Save(); err != nil {
		panic(err)
	}
	ctx.Redirect("/admin")
}

func approvePostHandler(ctx *golf.Context) {
	ctx.Send("TODO")
}

func publishPostHandler(ctx *golf.Context) {
	ctx.Send("TODO")
}

func deletePostHandler(ctx *golf.Context) {
	ctx.Send("TODO")
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

	app.Get("/admin", adminHandler)

	app.Get("/admin/new", newPostHandler)
	app.Post("/admin/new", insertPostHandler)
	app.Get("/admin/:page/edit", editPostHandler)
	app.Post("/admin/:page/edit", updatePostHandler)

	app.Put("/admin/:page/approve", approvePostHandler)
	app.Put("/admin/:page/publish", publishPostHandler)
	app.Delete("/admin/:page", deletePostHandler)

	app.Run(":45001")
}
