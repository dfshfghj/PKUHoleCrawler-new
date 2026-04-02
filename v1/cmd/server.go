package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"treehole/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var (
	serverPort string
	serverHost string
)

func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the API server",
		Long:  `启动 PKU Hole API 服务器，提供 RESTful 接口。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
	}

	cmd.Flags().StringVarP(&serverPort, "port", "p", "8081", "server port")
	cmd.Flags().StringVar(&serverHost, "host", "0.0.0.0", "server host")

	return cmd
}

func runServer() error {
	database, cleanup, err := initDB()
	if err != nil {
		return err
	}
	defer cleanup()

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

	r.GET("/pku_hole", func(c *gin.Context) {
		getPkuHolePosts(database, c)
	})

	r.GET("/post/:pid", func(c *gin.Context) {
		getPostByPid(database, c)
	})

	r.GET("/post/:pid/comments", func(c *gin.Context) {
		getCommentsByPid(database, c)
	})

	r.GET("/search", func(c *gin.Context) {
		searchPosts(database, c)
	})

	addr := fmt.Sprintf("%s:%s", serverHost, serverPort)
	log.Printf("Starting PKU Hole API server on %s...", addr)
	log.Printf("API endpoints:")
	log.Printf("  GET http://%s:%s/pku_hole?page=1&limit=25", serverHost, serverPort)
	log.Printf("  GET http://%s:%s/post/:pid", serverHost, serverPort)
	log.Printf("  GET http://%s:%s/post/:pid/comments", serverHost, serverPort)
	log.Printf("  GET http://%s:%s/search?q=keyword", serverHost, serverPort)
	log.Printf("  GET http://%s:%s/health", serverHost, serverPort)

	if err := r.Run(addr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

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
	PrevPageUrl  *string `json:"prev_page_url"`
	To           int     `json:"to"`
	Total        int     `json:"total"`
}

type ApiResponse struct {
	Code      int            `json:"code"`
	Data      PaginationData `json:"data"`
	Message   string         `json:"message"`
	Success   bool           `json:"success"`
	Timestamp int64          `json:"timestamp"`
}

func getPkuHolePosts(database *db.Database, c *gin.Context) {
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

	posts, err := database.GetPosts(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50000,
			"data":      nil,
			"message":   "Database query failed",
			"success":   false,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	totalCount, err := database.GetPostCount()
	if err != nil {
		totalCount = 0
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

	totalPages := 1
	if totalCount > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}
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
	}

	if page > 1 {
		prevUrl := fmt.Sprintf("%s?page=%d", basePath, page-1)
		prevPageUrl = &prevUrl
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

func getPostByPid(database *db.Database, c *gin.Context) {
	pidStr := c.Param("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"data":      nil,
			"message":   "Invalid pid",
			"success":   false,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	post, err := database.GetPostByPid(pid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":      40400,
			"data":      nil,
			"message":   "Post not found",
			"success":   false,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	apiPost := Post{
		Pid:       post.Pid,
		Text:      post.Text,
		Type:      post.Type,
		Timestamp: post.Timestamp,
		Reply:     post.Reply,
		Likenum:   post.Likenum,
		Extra:     post.Extra,
		Anonymous: post.Anonymous,
		IsTop:     post.IsTop,
		Label:     post.Label,
		Status:    post.Status,
		IsComment: post.IsComment,
		Tag:       post.Tag,
		IsFollow:  0,
		IsProtect: post.IsProtect,
		ImageSize: []int{post.ImageSizeX, post.ImageSizeY},
		LabelInfo: nil,
	}

	response := ApiResponse{
		Code:    20000,
		Message: "success",
		Success: true,
		Data: PaginationData{
			CurrentPage:  1,
			Data:         []Post{apiPost},
			FirstPageUrl: "",
			From:         1,
			LastPage:     1,
			LastPageUrl:  "",
			NextPageUrl:  "",
			Path:         fmt.Sprintf("http://treehole.pku.edu.cn/api/post/%d", pid),
			PerPage:      "1",
			PrevPageUrl:  nil,
			To:           1,
			Total:        1,
		},
		Timestamp: time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

func getCommentsByPid(database *db.Database, c *gin.Context) {
	pidStr := c.Param("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"data":      nil,
			"message":   "Invalid pid",
			"success":   false,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	comments, err := database.GetCommentsByPid(pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50000,
			"data":      nil,
			"message":   "Database query failed",
			"success":   false,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	type Comment struct {
		Cid       int    `json:"cid"`
		Pid       int    `json:"pid"`
		Name      string `json:"name"`
		Text      string `json:"text"`
		Timestamp int    `json:"timestamp"`
		Tag       string `json:"tag"`
		Quote     int    `json:"quote"`
	}

	var apiComments []Comment
	for _, cmt := range comments {
		apiComments = append(apiComments, Comment{
			Cid:       cmt.Cid,
			Pid:       cmt.Pid,
			Name:      cmt.Name,
			Text:      cmt.Text,
			Timestamp: cmt.Timestamp,
			Tag:       cmt.Tag,
			Quote:     cmt.Quote,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      20000,
		"data":      apiComments,
		"message":   "success",
		"success":   true,
		"timestamp": time.Now().Unix(),
	})
}

func searchPosts(database *db.Database, c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      40000,
			"data":      nil,
			"message":   "Missing search query parameter 'q'",
			"success":   false,
			"timestamp": time.Now().Unix(),
		})
		return
	}

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

	posts, err := database.SearchPosts(keyword, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":      50000,
			"data":      nil,
			"message":   "Search failed",
			"success":   false,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	totalCount, err := database.SearchPostsCount(keyword)
	if err != nil {
		totalCount = 0
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

	totalPages := 1
	if totalCount > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}
	from := offset + 1
	to := offset + len(posts)
	if to > totalCount {
		to = totalCount
	}

	basePath := "http://treehole.pku.edu.cn/api/search"
	firstPageUrl := fmt.Sprintf("%s?q=%s&page=1", basePath, keyword)
	lastPageUrl := fmt.Sprintf("%s?q=%s&page=%d", basePath, keyword, totalPages)

	var nextPageUrl string
	var prevPageUrl *string

	if page < totalPages {
		nextPageUrl = fmt.Sprintf("%s?q=%s&page=%d", basePath, keyword, page+1)
	}

	if page > 1 {
		prevUrl := fmt.Sprintf("%s?q=%s&page=%d", basePath, keyword, page-1)
		prevPageUrl = &prevUrl
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
