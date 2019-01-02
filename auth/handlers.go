package auth

import (
	"errors"
	"fmt"

	"github.com/dinever/golf"
)

func IsLoggedIn(ctx *golf.Context) bool {
	loggedIn, err := ctx.Session.Get("loggedIn")
	return err == nil && loggedIn.(bool)
}

func IsParent(ctx *golf.Context) bool {
	parent, err := ctx.Session.Get("parent")
	return IsLoggedIn(ctx) && err == nil && parent.(bool)
}

func RequireAuthentication(next golf.HandlerFunc) golf.HandlerFunc {
	fn := func(ctx *golf.Context) {
		if IsLoggedIn(ctx) {
			next(ctx)
		} else {
			ctx.Redirect("/login")
		}
	}
	return fn
}

func RequireParent(next golf.HandlerFunc) golf.HandlerFunc {
	fn := func(ctx *golf.Context) {
		if IsParent(ctx) {
			next(ctx)
		} else {
			ctx.Redirect("/login")
		}
	}
	return fn
}

func LoginHandler(ctx *golf.Context) {
	ctx.Loader("template").Render("login.html", map[string]interface{}{
		"Title": "Login",
	})
}

func findUser(admins []BlogAdmin, user string) (BlogAdmin, error) {
	for _, a := range admins {
		if a.Username == user {
			return a, nil
		}
	}
	return BlogAdmin{}, errors.New("User not found")
}

// LoginHandlerPost takes a slice of BlogAdmins and returns a
// HandlerFunc which is a closure over the admins. The admins
// will be consulted when validating a login request.
// LoginHandlerPost assumes form valkues "user" and "password".
// If login is successful, the following session variables will be
// set: "loggedIn" (bool), "user" (string, and "parent" (bool).
func LoginHandlerPost(admins []BlogAdmin) golf.HandlerFunc {
	return func(ctx *golf.Context) {
		user := ctx.Request.FormValue("user")
		password := ctx.Request.FormValue("password")
		fmt.Println("Login request by " + user)

		admin, err := findUser(admins, user)
		if err != nil {
			fmt.Println("Did not find user")
			goto noMatch
		}

		if password != admin.Password {
			fmt.Println("No password match")
			goto noMatch
		}

		ctx.Session.Set("loggedIn", true)
		ctx.Session.Set("user", admin.Username)
		ctx.Session.Set("parent", admin.IsParent)

		ctx.Redirect("/admin")
		return

	noMatch:
		ctx.Loader("template").Render("login.html", map[string]interface{}{
			"Message":    "Username or password was wrong, please try again!",
			"HasMessage": true,
		})
	}
}

func LogoutHandler(ctx *golf.Context) {
	ctx.Session.Delete("loggedIn")
	ctx.Session.Delete("user")
	ctx.Session.Delete("parent")
	ctx.Redirect("/")
}
