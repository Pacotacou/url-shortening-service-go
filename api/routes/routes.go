package routes

import (
	"api/helpers"
	"api/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	charset         = "abcdefghijklmnopkrstuvwxyz1234567890"
	shortcodeLength = 6
)

func LoadRoutes(router *gin.Engine, urlRepo *models.URLRepository) {

	setupRouter(router)

	// ping
	router.GET("/ping", pingHandler)

	// create a new short URL
	router.POST("/shorten", func(c *gin.Context) {
		shortenPostHandler(c, urlRepo)
	})

	// retrieve an original URL from a short URL
	router.GET("/shorten/:shortcode", func(c *gin.Context) {
		shortenGetHandler(c, urlRepo)
	})
	// update an existing short URL
	router.PUT("/shorten/:shortcode", func(c *gin.Context) {
		shortenPutHandler(c, urlRepo)
	})

	// delete an existing short URL
	router.DELETE("/shorten/:shortcode", func(c *gin.Context) {
		shortenDeleteHandler(c, urlRepo)
	})

	// get statistics on the short URL
	router.GET("/shorten/:shortcode/stats", func(c *gin.Context) {
		shortenStatsGetHandler(c, urlRepo)
	})
}

func shortenStatsGetHandler(c *gin.Context, urlRepo *models.URLRepository) {
	//get shortcode
	shortcode := c.Param("shortcode")
	//get stats from database
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	url, err := urlRepo.GetByShortcode(shortcode, ctx)
	if err != nil {
		fmt.Println("Error: ", err)
		c.JSON(404, "url not found")
		return
	}
	//response
	c.JSON(200, gin.H{
		"id":          url.ID,
		"url":         url.OriginalURL,
		"shortCode":   url.Shortcode,
		"createdAt":   url.CreatedAt,
		"updatedAt":   url.UpdatedAt,
		"accessCount": url.AccessCount,
	})
}

func shortenDeleteHandler(c *gin.Context, urlRepo *models.URLRepository) {
	//get shortcode
	shortcode := c.Param("shortcode")
	//verify url in database
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err := urlRepo.GetByShortcode(shortcode, ctx)
	if err != nil {
		fmt.Println("Error:", err)
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	// delete from database
	err = urlRepo.DeleteURL(shortcode, ctx)
	if err != nil {
		fmt.Println("Error: ", err)
		c.JSON(500, gin.H{"error": "server internal error"})
		return
	}
	// response
	c.JSON(204, gin.H{"message": "successfully deleted"})

}

func shortenPutHandler(c *gin.Context, urlRepo *models.URLRepository) {
	//get shortcode and context
	shortcode := c.Param("shortcode")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	//get new original url
	var body struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Error parsing body: ", err)
		c.JSON(300, gin.H{"error": "error in the request body"})
		return
	}

	// update shortcode with new url in database
	url, err := urlRepo.UpdateURL(shortcode, body.URL, ctx)
	if err != nil {
		fmt.Println("Error updating url:", err)
		c.JSON(500, gin.H{"error": "server internal error"})
		return
	}

	//return response
	c.JSON(200, gin.H{
		"id":        url.ID,
		"url":       url.OriginalURL,
		"shortCode": url.Shortcode,
		"createdAt": url.CreatedAt,
		"updatedAt": url.UpdatedAt,
	})
}

func shortenGetHandler(c *gin.Context, urlRepo *models.URLRepository) {
	shortcode := c.Param("shortcode")
	ctx := c.Request.Context()
	//Look for shortened url in database
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	url, err := urlRepo.GetByShortcode(shortcode, ctx)
	if err != nil {
		fmt.Println("Error:", err)
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	//increase count
	err = urlRepo.UpdateAccessCountByShortcode(shortcode, ctx)
	if err != nil {
		fmt.Println("Error incrementing count:", err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	//return url
	c.JSON(201, gin.H{
		"id":        url.ID,
		"url":       url.OriginalURL,
		"shortCode": url.Shortcode,
		"createdAt": url.CreatedAt,
		"updatedAt": url.UpdatedAt,
	})

}

func setupRouter(router *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true

	router.Use(cors.New(config))
}

func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func shortenPostHandler(c *gin.Context, urlRepo *models.URLRepository) {
	// get url from the body of the request
	var body struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Error parsing body: ", err)
		c.JSON(300, gin.H{"error": "error in the request body"})
		return
	}

	// validate the url
	isValidUrl, err := helpers.ValidateURL(body.URL)
	if err != nil {
		fmt.Println("Error validating URL", err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	if !isValidUrl {
		fmt.Println("The URL is not valid")
		c.JSON(300, gin.H{"error": "the url is not valid"})
		return
	}

	// generate short code
	var shortcode string
	for {
		sc, err := helpers.GenerateShortcode(shortcodeLength, charset)
		if err != nil {
			fmt.Println("Error:", err)
			c.JSON(500, gin.H{"error": "internal server error"})
			return
		}
		isInDatabase, err := helpers.ShortcodeInDatabase(sc, urlRepo)
		if err != nil {
			fmt.Println("Error:", err)
			c.JSON(500, gin.H{"error": "internal server error"})
			return
		}
		if !isInDatabase {
			shortcode = sc
			break
		}
	}

	savedURL, err := helpers.SaveShortenedUrl(shortcode, body.URL, urlRepo)
	if err != nil {
		fmt.Println("Error: ", err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(200, savedURL)
}
