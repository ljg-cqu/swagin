package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/long2ice/swagin"
	"github.com/long2ice/swagin/security"
)

func main() {
	// Use customize Gin engine
	r := gin.New()

	// Registering func(c *gin.Context) is accepted,
	// but the OpenAPI generator will ignore the operation and it won't appear in the specification.
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	app := swagin.NewFromEngine(r, NewSwagger())
	subApp := swagin.NewFromEngine(r, NewSwagger())

	/*
		You can use default Gin engin:
			app := swagin.New(NewSwagger())
			subApp := swagin.New(NewSwagger())
	*/

	subApp.GET("/noModel", noModel)
	app.Mount("/sub", subApp)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))
	queryGroup := app.Group("/query", swagin.Tags("Query"))
	queryGroup.GET("/list", queryList)
	queryGroup.GET("/:id", queryPath)
	queryGroup.DELETE("", query)

	app.GET("/noModel", noModel)

	formGroup := app.Group("/form", swagin.Tags("Form"), swagin.Security(&security.Bearer{}))
	formGroup.POST("/encoded", formEncode)
	formGroup.PUT("", body)
	formGroup.POST("/file", file)

	port := ":8084"

	fmt.Printf("Now you can visit http://127.0.0.1%v/docs or http://127.0.0.1%v/redoc to see the api docs. Have fun!", port, port)
	if err := app.Run(port); err != nil {
		panic(err)
	}
}
