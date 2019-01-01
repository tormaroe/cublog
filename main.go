package main

import (
	"fmt"
	"html/template"
	"os"

	"github.com/dinever/golf"
	"github.com/tormaroe/cublog/posts"
)

var blogState = posts.NewBlogState()

func mainHandler(ctx *golf.Context) {
	title, _ := ctx.App.Config.GetString("blogName", "The Blog")
	data := map[string]interface{}{
		"Title": title,
		"Posts": blogState.MainPagePosts(),
	}
	ctx.Loader("template").Render("index.html", data)
}

func pageHandler(ctx *golf.Context) {
	post, err := blogState.FindPost(ctx.Param("page"))
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

func requireAuthentication(next golf.HandlerFunc) golf.HandlerFunc {
	fn := func(ctx *golf.Context) {
		if loggedIn, err := ctx.Session.Get("loggedIn"); err == nil && loggedIn.(bool) {
			next(ctx)
		} else {
			ctx.Redirect("/login")
		}
	}
	return fn
}

func requireParent(next golf.HandlerFunc) golf.HandlerFunc {
	fn := func(ctx *golf.Context) {
		if loggedIn, err := ctx.Session.Get("loggedIn"); err == nil && loggedIn.(bool) {
			next(ctx)
		} else {
			ctx.Redirect("/login")
		}
	}
	return fn
}

func loginHandler(ctx *golf.Context) {
	ctx.Loader("template").Render("login.html", make(map[string]interface{}))
}

func loginHandlerPost(ctx *golf.Context) {
	user := ctx.Request.FormValue("user")
	password := ctx.Request.FormValue("password")
	fmt.Println("Login request by " + user)

	parentPassword, _ := ctx.App.Config.GetString("parentPassword", "øæoai3wryfuoøi<es")
	childPassword, _ := ctx.App.Config.GetString("childPassword", "poaøirefghuaoiuhf")

	if password == parentPassword {
		ctx.Session.Set("loggedIn", true)
		ctx.Session.Set("user", user)
		ctx.Session.Set("parent", true)
	} else if password == childPassword {
		ctx.Session.Set("loggedIn", true)
		ctx.Session.Set("user", user)
		ctx.Session.Set("parent", false)
	} else {
		ctx.Loader("template").Render("login.html", map[string]interface{}{
			"Message":    "Username or password was wrong, please try again!",
			"HasMessage": true,
		})
	}

	ctx.Redirect("/admin")
}

func logoutHandler(ctx *golf.Context) {
	ctx.Session.Delete("loggedIn")
	ctx.Session.Delete("user")
	ctx.Session.Delete("parent")
	ctx.Redirect("/")
}

func adminHandler(ctx *golf.Context) {
	data := map[string]interface{}{
		"Title": "Admin",
		"Posts": blogState.AdminPagePosts(),
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
	err := ctx.Request.ParseForm() // Not needed?!
	if err != nil {
		panic(err)
	}
	post := posts.New(ctx.Request.FormValue("PostTitle"), ctx.Request.FormValue("PostSlug"), ctx.Request.FormValue("PostBody"))
	err = blogState.AddAndSave(post)
	if err != nil {
		panic(err)
	}

	ctx.Redirect("/admin")
}

func editPostHandler(ctx *golf.Context) {
	post, err := blogState.FindPost(ctx.Param("page"))
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
	post, err := blogState.FindPost(ctx.Param("page"))
	if err != nil {
		ctx.Abort(404)
		return
	}
	post.Title = ctx.Request.FormValue("PostTitle")
	post.Path = ctx.Request.FormValue("PostSlug")
	post.Body = template.HTML(ctx.Request.FormValue("PostBody"))
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

	if err = blogState.Load(); err != nil {
		panic(err)
	}

	app := golf.New()
	app.Config, err = golf.ConfigFromJSON(file)
	app.View.SetTemplateLoader("template", "www/templates/")

	app.SessionManager = golf.NewMemorySessionManager()
	app.Use(golf.SessionMiddleware)

	app.Static("/www/static/", "static")

	app.Get("/", mainHandler)
	app.Get("/p/:page/", pageHandler)

	app.Get("/login", loginHandler)
	app.Post("/login", loginHandlerPost)
	app.Get("/logout", logoutHandler)

	authChain := golf.NewChain(requireAuthentication)
	parentChain := golf.NewChain(requireAuthentication)
	parentChain.Append(requireParent)

	app.Get("/admin", authChain.Final(adminHandler))

	app.Get("/admin/new", authChain.Final(newPostHandler))
	app.Post("/admin/new", authChain.Final(insertPostHandler))
	app.Get("/admin/:page/edit", authChain.Final(editPostHandler))
	app.Post("/admin/:page/edit", authChain.Final(updatePostHandler))

	app.Put("/admin/:page/approve", parentChain.Final(approvePostHandler))
	app.Put("/admin/:page/publish", authChain.Final(publishPostHandler))
	app.Delete("/admin/:page", authChain.Final(deletePostHandler))

	app.Run(":9000")
}
