package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/movies", createMovie)
	r.GET("/movies", listMovies)
	r.GET("/movies/:id", getMovie)
	r.PUT("/movies/:id", updateMovie)
	r.DELETE("/movies/:id", deleteMovie)

	r.Run(":8080")
}
