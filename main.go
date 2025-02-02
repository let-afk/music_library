package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/swaggo/swag/example/basic/docs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Music Library API
// @version 1.0
// @description API for managing an online music library.
// @host localhost:8080
// @BasePath /

type Song struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Group       string `json:"group"`
	Song        string `json:"song"`
	ReleaseDate string `json:"release_date"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

var db *gorm.DB

func initDB() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	db.AutoMigrate(&Song{})
}

// @Summary Get all songs with filtering and pagination
// @Description Get list of all songs with optional filtering and pagination
// @Produce json
// @Param group query string false "Group Name"
// @Param song query string false "Song Name"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} Song
// @Router /songs [get]
func getSongs(c *gin.Context) {
	var songs []Song
	query := db

	if group := c.Query("group"); group != "" {
		query = query.Where("group = ?", group)
	}
	if song := c.Query("song"); song != "" {
		query = query.Where("song = ?", song)
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	query.Limit(limit).Offset(offset).Find(&songs)

	c.JSON(http.StatusOK, songs)
}

// @Summary Get song lyrics with pagination
// @Description Get lyrics of a song with pagination (verses per page)
// @Produce json
// @Param id path int true "Song ID"
// @Param page query int true "Page number"
// @Param per_page query int true "Verses per page"
// @Success 200 {object} map[string]string
// @Router /songs/{id}/lyrics [get]
func getSongLyrics(c *gin.Context) {
	id := c.Param("id")
	var song Song
	if err := db.First(&song, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	verses := strings.Split(song.Text, "\n")
	c.JSON(http.StatusOK, gin.H{"lyrics": verses})
}

// @Summary Add a new song
// @Description Add a new song to the library
// @Accept json
// @Produce json
// @Param song body Song true "Song Data"
// @Success 201 {object} Song
// @Router /songs [post]
func addSong(c *gin.Context) {
	var song Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	db.Create(&song)
	c.JSON(http.StatusCreated, song)
}

// @Summary Delete a song
// @Description Delete a song by ID
// @Param id path int true "Song ID"
// @Success 200 {object} map[string]string
// @Router /songs/{id} [delete]
func deleteSong(c *gin.Context) {
	id := c.Param("id")
	db.Delete(&Song{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Song deleted"})
}

// @Summary Update a song
// @Description Update details of an existing song by ID
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param song body Song true "Updated Song Data"
// @Success 200 {object} Song
// @Router /songs/{id} [put]
func updateSong(c *gin.Context) {
	id := c.Param("id")
	var song Song
	if err := db.First(&song, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	db.Save(&song)
	c.JSON(http.StatusOK, song)
}

func main() {
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	} else {
		logrus.Info(".env file loaded")
	}

	initDB()

	r := gin.Default()

	r.GET("/songs", getSongs)
	r.GET("/songs/:id/lyrics", getSongLyrics)
	r.POST("/songs", addSong)
	r.DELETE("/songs/:id", deleteSong)
	r.PUT("/songs/:id", updateSong)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logrus.Infof("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		logrus.Fatalf("Error starting server: %v", err)
	}
}
