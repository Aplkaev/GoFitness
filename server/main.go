package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HelloRequest struct {
	UserID    int64  `json:"userId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
}

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.POST("/api/hello", func(c *gin.Context) {
		var req HelloRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Здесь можно проверять initData по https://docs.telegram-mini-apps.com/platform/init-data
		// но для локальной разработки пропускаем

		c.JSON(http.StatusOK, gin.H{
			"message": "Привет, " + req.FirstName + "! (user #" + string(req.UserID) + ")",
			"serverTime": time.Now().Format(time.RFC3339),
		})
	})

	r.Run(":3001") // Важно: 3001, а не 3000 (чтобы не конфликтовать с vite)
}