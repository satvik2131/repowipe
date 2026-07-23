package routes

import (
	"repowipe/controllers"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	base := r.Group("/api")

	// ── Auth ─────────────────────────────────────────────────────────────────
	auth := base.Group("/auth")
	{
		// Returns the full GitHub OAuth URL — client never needs to know client_id
		auth.GET("/github/login", controllers.GetGithubLoginURL)

		// Receives the temporary code from the GitHub OAuth callback and
		// exchanges it for an access token, then sets a session cookie.
		auth.POST("/github/callback", controllers.SetAccessToken)

		// Clears Redis session, revokes GitHub token, and expires the cookie.
		auth.POST("/logout", controllers.Logout)
	}

	// Kept for backward-compat if anything still points at the old path.
	base.POST("/set/access/token", controllers.SetAccessToken)

	// ── Session ───────────────────────────────────────────────────────────────
	base.GET("/verify/user", controllers.VerifyUser)

	// ── Repos ─────────────────────────────────────────────────────────────────
	base.POST("/fetch/repos", controllers.FetchAllRepos)
	base.GET("/search/repo", controllers.SearchRepos)
	base.DELETE("/delete/repos", controllers.DeleteRepos)
}