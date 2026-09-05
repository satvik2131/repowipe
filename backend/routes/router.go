package routes

import (
	"repowipe/controllers"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	base := r.Group("/api")

	// ── Auth (static paths before :provider) ──────────────────────────────────
	auth := base.Group("/auth")
	{
		auth.GET("/connections", controllers.GetConnections)
		auth.POST("/logout", controllers.Logout)
		auth.GET("/:provider/login", controllers.GetProviderLoginURL)
		auth.POST("/:provider/callback", controllers.ProviderCallback)
		auth.DELETE("/:provider", controllers.UnlinkProvider)
	}

	base.POST("/set/access/token", controllers.SetAccessToken)
	base.GET("/verify/user", controllers.VerifyUser)

	// Legacy GitHub-only paths (before /:provider/*)
	base.POST("/fetch/repos", controllers.FetchAllRepos)
	base.GET("/search/repo", controllers.SearchRepos)
	base.DELETE("/delete/repos", controllers.DeleteRepos)

	// Transfers (before /:provider/*)
	base.POST("/transfers", controllers.StartTransfer)
	base.GET("/transfers/:id", controllers.GetTransfer)

	// Provider-scoped repos
	base.POST("/:provider/repos", controllers.FetchProviderRepos)
	base.GET("/:provider/search", controllers.SearchProviderRepos)
	base.DELETE("/:provider/repos", controllers.DeleteProviderRepos)
}
