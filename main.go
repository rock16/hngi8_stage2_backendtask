package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	//port := os.Getenv("PORT")

	//if port == "" {
	//log.Fatal("$PORT must be set")
	//}

	router := gin.Default()
	router.Static("/", "./template")

	router.POST("/upload", func(c *gin.Context) {
		name := c.PostForm("name")
		email := c.PostForm("email")

		c.String(http.StatusOK, fmt.Sprintf("Thank you for getting in touch with me fields name= %s and email= %s.", name, email))
	})

	router.Run(":8080")
}
