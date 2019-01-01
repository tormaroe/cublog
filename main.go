package main

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/dinever/golf"
	"github.com/tormaroe/cublog/posts"
)

var blogState = posts.NewBlogState()

func defaultTemplateData(ctx *golf.Context, title string) map[string]interface{} {
	return map[string]interface{}{
		"Title":      title,
		"IsLoggedIn": isLoggedIn(ctx),
		"IsParent":   isParent(ctx),
	}
}

func mainHandler(ctx *golf.Context) {
	title, _ := ctx.App.Config.GetString("blogName", "The Blog")
	data := defaultTemplateData(ctx, title)
	data["Posts"] = blogState.MainPagePosts()
	ctx.Loader("template").Render("index.html", data)
}

func pageHandler(ctx *golf.Context) {
	post, err := blogState.FindPost(ctx.Param("page"))
	if err != nil {
		ctx.Abort(404)
		return
	}
	data := defaultTemplateData(ctx, post.Title)
	data["Body"] = post.Body
	ctx.Loader("template").Render("post.html", data)
}

func isLoggedIn(ctx *golf.Context) bool {
	loggedIn, err := ctx.Session.Get("loggedIn")
	return err == nil && loggedIn.(bool)
}

func isParent(ctx *golf.Context) bool {
	parent, err := ctx.Session.Get("parent")
	return isLoggedIn(ctx) && err == nil && parent.(bool)
}

func requireAuthentication(next golf.HandlerFunc) golf.HandlerFunc {
	fn := func(ctx *golf.Context) {
		if isLoggedIn(ctx) {
			next(ctx)
		} else {
			ctx.Redirect("/login")
		}
	}
	return fn
}

func requireParent(next golf.HandlerFunc) golf.HandlerFunc {
	fn := func(ctx *golf.Context) {
		if isParent(ctx) {
			next(ctx)
		} else {
			ctx.Redirect("/login")
		}
	}
	return fn
}

func loginHandler(ctx *golf.Context) {
	ctx.Loader("template").Render("login.html", defaultTemplateData(ctx, "Login"))
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
	data := defaultTemplateData(ctx, "Admin")
	data["Posts"] = blogState.AdminPagePosts()
	ctx.Loader("template").Render("admin.html", data)
}

func newPostHandler(ctx *golf.Context) {
	ctx.Loader("template").Render("post-form.html", defaultTemplateData(ctx, "New post"))
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

	data := defaultTemplateData(ctx, "Edit: "+post.Title)
	data["Post"] = post
	ctx.Loader("template").Render("post-form.html", data)
}

func updatePostHandler(ctx *golf.Context) {
	err := ctx.Request.ParseForm()
	if err != nil {
		panic(err)
	}
	post, err := blogState.FindPost(ctx.Param("page"))
	if err != nil {
		ctx.SendStatus(404)
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
	post, err := blogState.FindPost(ctx.Param("page"))
	if err != nil {
		ctx.SendStatus(404)
		return
	}
	post.Approved = true
	post.Save()
	ctx.SendStatus(200)
}

func publishPostHandler(ctx *golf.Context) {
	post, err := blogState.FindPost(ctx.Param("page"))
	if err != nil {
		ctx.SendStatus(404)
		return
	}
	post.Published = true
	post.PublishedDate = time.Now()
	post.Save()
	ctx.SendStatus(200)
}

func deletePostHandler(ctx *golf.Context) {

	post, err := blogState.FindPost(ctx.Param("page"))
	if err != nil {
		ctx.SendStatus(404)
		return
	}
	post.Deleted = true
	post.Save()
	ctx.SendStatus(200)
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
