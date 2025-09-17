package routes

import (
	"repowipe/controllers"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine){
	base := r.Group("/api")
	{
		base.GET("/verify/user",controllers.VerifyUser)
		base.POST("/set/access/token",controllers.SetAccessToken)
		base.POST("/fetch/repos",controllers.FetchAllRepos)
		base.GET("/search/repo",controllers.SearchRepos)
	}
}