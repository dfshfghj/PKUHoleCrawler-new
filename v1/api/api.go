package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"treehole/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// Post represents a post in the PKU Hole database with the required response format
type Post struct {
	Pid       int         `json:"pid"`
	Text      string      `json:"text"`
	Type      string      `json:"type"`
	Timestamp int         `json:"timestamp"`
	Reply     int         `json:"reply"`
	Likenum   int         `json:"likenum"`
	Extra     int         `json:"extra"`
	Anonymous int         `json:"anonymous"`
	IsTop     int         `json:"is_top"`
	Label     int         `json:"label"`
	Status    int         `json:"status"`
	IsComment int         `json:"is_comment"`
	Tag       string      `json:"tag"`
	IsFollow  int         `json:"is_follow"`
	IsProtect int         `json:"is_protect"`
	ImageSize []int       `json:"image_size"`
	LabelInfo interface{} `json:"label_info"`
}

// PaginationData represents the pagination structure in the response
type PaginationData struct {
	CurrentPage  int     `json:"current_page"`
	Data         []Post  `json:"data"`
	FirstPageUrl string  `json:"first_page_url"`
	From         int     `json:"from"`
	LastPage     int     `json:"last_page"`
	LastPageUrl  string  `json:"last_page_url"`
	NextPageUrl  string  `json:"next_page_url"`
	Path         string  `json:"path"`
	PerPage      string  `json:"per_page"`
	PrevPageUrl  *string `json:"prev_page_url"` // Use pointer to allow null
	To           int     `json:"to"`
	Total        int     `json:"total"`
}

// ApiResponse represents the complete API response structure
type ApiResponse struct {
	Code      int            `json:"code"`
	Data      PaginationData `json:"data"`
	Message   string         `json:"message"`
	Success   bool           `json:"success"`
	Timestamp int64          `json:"timestamp"`
}

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("./data.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)

	log.Println("Database connected successfully")
}

func getPkuHolePosts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "25")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 25
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	var posts []models.Post
	if err := db.Order("pid DESC").Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50000,
			"data":      nil,
			"message":   "Database query failed",
			"success":   false,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	var apiPosts []Post
	for _, p := range posts {
		apiPosts = append(apiPosts, Post{
			Pid:       p.Pid,
			Text:      p.Text,
			Type:      p.Type,
			Timestamp: p.Timestamp,
			Reply:     p.Reply,
			Likenum:   p.Likenum,
			Extra:     p.Extra,
			Anonymous: p.Anonymous,
			IsTop:     p.IsTop,
			Label:     p.Label,
			Status:    p.Status,
			IsComment: p.IsComment,
			Tag:       p.Tag,
			IsFollow:  0,
			IsProtect: p.IsProtect,
			ImageSize: []int{p.ImageSizeX, p.ImageSizeY},
			LabelInfo: nil,
		})
	}

	totalCount := 7619157
	totalPages := (totalCount + limit - 1) / limit
	from := offset + 1
	to := offset + len(posts)
	if to > totalCount {
		to = totalCount
	}

	basePath := "http://treehole.pku.edu.cn/api/pku_hole"
	firstPageUrl := fmt.Sprintf("%s?page=1", basePath)
	lastPageUrl := fmt.Sprintf("%s?page=%d", basePath, totalPages)

	var nextPageUrl string
	var prevPageUrl *string

	if page < totalPages {
		nextPageUrl = fmt.Sprintf("%s?page=%d", basePath, page+1)
	} else {
		nextPageUrl = ""
	}

	if page > 1 {
		prevUrl := fmt.Sprintf("%s?page=%d", basePath, page-1)
		prevPageUrl = &prevUrl
	} else {
		prevPageUrl = nil
	}

	paginationData := PaginationData{
		CurrentPage:  page,
		Data:         apiPosts,
		FirstPageUrl: firstPageUrl,
		From:         from,
		LastPage:     totalPages,
		LastPageUrl:  lastPageUrl,
		NextPageUrl:  nextPageUrl,
		Path:         basePath,
		PerPage:      strconv.Itoa(limit),
		PrevPageUrl:  prevPageUrl,
		To:           to,
		Total:        totalCount,
	}

	response := ApiResponse{
		Code:      20000,
		Data:      paginationData,
		Message:   "success",
		Success:   true,
		Timestamp: time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

func main() {
	initDB()

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":      20000,
			"data":      gin.H{"status": "ok", "message": "PKU Hole API is running"},
			"message":   "success",
			"success":   true,
			"timestamp": time.Now().Unix(),
		})
	})

	r.GET("/pku_hole", getPkuHolePosts)

	port := "8081"
	fmt.Printf("Starting PKU Hole API server on port %s...\n", port)
	fmt.Printf("API endpoint: http://localhost:%s/pku_hole?page=1&limit=25\n", port)
	fmt.Printf("Health check: http://localhost:%s/health\n", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
