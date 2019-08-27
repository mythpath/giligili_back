package api

import (
	"github.com/gin-gonic/gin"
	"selfText/giligili_back/service"
)

func DailyRank(c *gin.Context) {
	dailyRankService := service.DailyRankService{}

	if err := c.ShouldBind(&dailyRankService); err == nil {
		res := dailyRankService.Get()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}
