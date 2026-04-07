package handles

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"treehole/internal/db"
	"treehole/internal/models"
	"treehole/server/utils"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	utils.RespondSuccess(c, gin.H{"status": "ok", "message": "PKU Hole API is running"})
}

func GetPost(database *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		pid, err := strconv.ParseInt(c.Param("pid"), 10, 32)
		if err != nil {
			utils.RespondError(c, 400, "InvalidParam", err)
			return
		}

		p, err := database.GetPostByPid(int32(pid))
		if err != nil {
			utils.RespondError(c, 404, "NotFound", err)
			return
		}

		username := "anonymous"
		if !p.Anonymous {
			username = ""
		}

		postData := map[string]interface{}{
			"id":        p.Pid,
			"text":      p.Text,
			"userid":    65535,
			"username":  username,
			"timestamp": p.Timestamp,
			"reply":     p.Reply,
			"follownum": p.Likenum,
			"is_follow": 0,
			"likenum":   p.PraiseNum,
			"is_like":   0,
			"type":      p.Type,
			"tags":      []string{},
			"media_ids": p.MediaIds,
		}

		utils.RespondSuccess(c, postData)
	}
}

// GetPosts 统一处理帖子获取和搜索
func GetPosts(database *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyword := c.Query("keyword")
		orderBy := c.Query("order_by")

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "25"))
		if err != nil || limit < 1 {
			limit = 25
		}
		if limit > 100 {
			limit = 100
		}

		cursor, _ := strconv.Atoi(c.DefaultQuery("begin", "0"))

		var posts []models.Post
		var dbErr error

		if orderBy != "" && keyword == "" {
			posts, dbErr = database.GetPostsOrderBy(orderBy, cursor, limit)
		} else if orderBy != "" && keyword != "" {
			posts, dbErr = database.SearchPostsOrderBy(keyword, orderBy, cursor, limit)
		} else if keyword != "" {
			posts, dbErr = database.SearchPostsCursor(keyword, cursor, limit, false)
		} else {
			posts, dbErr = database.GetPostsCursor(cursor, limit, false)
		}

		if dbErr != nil {
			utils.RespondError(c, 500, "ServerError", dbErr)
			return
		}

		postData := make([]map[string]interface{}, len(posts))
		for i, p := range posts {
			username := "anonymous"
			if !p.Anonymous {
				username = ""
			}

			postData[i] = map[string]interface{}{
				"id":        p.Pid,
				"text":      p.Text,
				"userid":    65535,
				"username":  username,
				"timestamp": p.Timestamp,
				"reply":     p.Reply,
				"follownum": p.Likenum,
				"is_follow": 0,
				"likenum":   p.PraiseNum,
				"is_like":   0,
				"type":      p.Type,
				"tags":      []string{},
				"media_ids": p.MediaIds,
			}
		}

		utils.RespondSuccess(c, postData)
	}
}

func GetComments(database *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		pid, err := strconv.ParseInt(c.Param("pid"), 10, 32)
		if err != nil {
			utils.RespondError(c, 400, "InvalidParam", err)
			return
		}

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "25"))
		if err != nil || limit < 1 {
			limit = 25
		}
		if limit > 100 {
			limit = 100
		}

		cursor, _ := strconv.ParseInt(c.DefaultQuery("begin", "0"), 10, 32)

		// 添加sort参数，0表示升序(asc)，1表示降序(desc)，默认为0（升序）
		sortParam := c.DefaultQuery("sort", "0")
		sortAsc := true
		if sortParam == "1" {
			sortAsc = false
		}

		comments, err := database.GetCommentsByPidCursor(int32(pid), int32(cursor), limit, sortAsc)
		if err != nil {
			utils.RespondError(c, 500, "ServerError", err)
			return
		}

		commentData := make([]map[string]interface{}, len(comments))
		for i, cmt := range comments {
			commentData[i] = map[string]interface{}{
				"cid":       cmt.Cid,
				"pid":       cmt.Pid,
				"userid":    65535,
				"username":  cmt.NameTag,
				"text":      cmt.Text,
				"timestamp": cmt.Timestamp,
				"quote":     cmt.Quote,
				"media_ids": cmt.MediaIds,
			}
		}

		utils.RespondSuccess(c, commentData)
	}
}

func GetImage(c *gin.Context) {
	idStr := c.Query("id")
	pidStr := c.Query("pid")

	if idStr == "" && pidStr == "" {
		utils.RespondError(c, 400, "InvalidParam", errors.New("missing id or pid parameter"))
		return
	}

	var filename string
	dir := "data/images"

	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			utils.RespondError(c, 400, "InvalidParam", errors.New("invalid id"))
			return
		}

		filename = findImageFile(dir, strconv.Itoa(id))
	} else {
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			utils.RespondError(c, 400, "InvalidParam", errors.New("invalid pid"))
			return
		}
		filename = findImageFile(dir, strconv.Itoa(pid))
	}

	if filename == "" {
		c.Status(http.StatusNotFound)
		return
	}

	c.File(filename)
}

func findImageFile(dir, base string) string {
	exts := []string{".jpg", ".png", ".gif", ".webp"}
	for _, ext := range exts {
		path := filepath.Join(dir, base+ext)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
