package crawler

import (
	"encoding/json"
	"fmt"
	"log"

	"treehole/internal/client"
	"treehole/internal/db"
	"treehole/internal/models"
)

// FetchResult holds the result of fetching a single page
type FetchResult struct {
	PostCount    int
	CommentCount int
}

// FetchAndSave fetches one page of posts and saves them to the database
func FetchAndSave(c *client.Client, database *db.Database, page int) (FetchResult, error) {
	var result FetchResult

	posts, comments, err := fetchPostsV3(c, page, 100)
	if err != nil {
		return result, err
	}

	result.PostCount = len(posts)
	result.CommentCount = len(comments)

	if err := database.UpsertPosts(posts); err != nil {
		log.Printf("[Crawler] 写入帖子失败: %v", err)
	}

	if len(comments) > 0 {
		if err := database.UpsertComments(comments); err != nil {
			log.Printf("[Crawler] 写入评论失败: %v", err)
		}
	}

	return result, nil
}

func fetchPostsV3(c *client.Client, page int, limit int) ([]models.Post, []models.Comment, error) {
	log.Printf("[Crawler] 正在请求 API: page=%d, limit=%d", page, limit)
	resp, err := c.GetPostsList(page, limit, 20, 1)
	if err != nil {
		log.Printf("[Crawler] API 请求失败: %v", err)
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[Crawler] API 返回非200状态码: %d", resp.StatusCode)
		return nil, nil, fmt.Errorf("get posts failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Printf("[Crawler] JSON 解析失败: %v", err)
		return nil, nil, err
	}

	code, ok := result["code"].(float64)
	if !ok || code != 20000 {
		message, _ := result["message"].(string)
		log.Printf("[Crawler] API 业务错误: code=%v, message=%s", code, message)
		return nil, nil, fmt.Errorf("API error: %s", message)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		log.Printf("[Crawler] 无效的 data 格式")
		return nil, nil, fmt.Errorf("invalid data format")
	}

	postsData, ok := data["list"].([]interface{})
	if !ok {
		log.Printf("[Crawler] 无效的 posts data 格式")
		return nil, nil, fmt.Errorf("invalid posts data format")
	}

	log.Printf("[Crawler] 解析到 %d 条帖子数据", len(postsData))

	var posts []models.Post
	var comments []models.Comment

	for _, item := range postsData {
		postMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var pid, anonymous, extra, timestamp, reply, likenum, label int
		var text, postType, tag, mediaIds string

		if v, ok := postMap["pid"].(float64); ok {
			pid = int(v)
		}
		if v, ok := postMap["text"].(string); ok {
			text = v
		}
		if v, ok := postMap["anonymous"].(float64); ok {
			anonymous = int(v)
		}
		if v, ok := postMap["type"].(string); ok {
			postType = v
		} else {
			postType = "text"
		}
		if v, ok := postMap["extra"].(float64); ok {
			extra = int(v)
		}
		if v, ok := postMap["timestamp"].(float64); ok {
			timestamp = int(v)
		}
		if v, ok := postMap["reply"].(float64); ok {
			reply = int(v)
		}
		if v, ok := postMap["likenum"].(float64); ok {
			likenum = int(v)
		}
		if v, ok := postMap["tag"].(string); ok {
			tag = v
		}
		if v, ok := postMap["label"].(float64); ok {
			label = int(v)
		}
		if v, ok := postMap["media_ids"].(string); ok {
			mediaIds = v
		}

		post := models.Post{
			Pid: pid, Text: text, Anonymous: anonymous, Type: postType,
			Extra: extra, Timestamp: timestamp, Reply: reply, Likenum: likenum,
			Tag: tag, IsComment: 1, Label: label, MediaIds: mediaIds,
		}
		posts = append(posts, post)

		if commentList, ok := postMap["comment_list"].([]interface{}); ok {
			for _, commentItem := range commentList {
				commentMap, ok := commentItem.(map[string]interface{})
				if !ok {
					continue
				}

				var cid, pidVal, cTimestamp int
				var cText, nameTag, cTag, cMediaIds string

				if v, ok := commentMap["cid"].(float64); ok {
					cid = int(v)
				}
				if v, ok := commentMap["pid"].(float64); ok {
					pidVal = int(v)
				}
				if v, ok := commentMap["text"].(string); ok {
					cText = v
				}
				if v, ok := commentMap["timestamp"].(float64); ok {
					cTimestamp = int(v)
				}
				if v, ok := commentMap["name_tag"].(string); ok {
					nameTag = v
				}
				if v, ok := commentMap["tag"].(string); ok {
					cTag = v
				}
				if v, ok := commentMap["media_ids"].(string); ok {
					cMediaIds = v
				}

				commentID := commentMap["comment_id"]
				var quote int
				if commentID != nil {
					if v, ok := commentID.(float64); ok {
						quote = int(v)
					}
				}

				comment := models.Comment{
					Cid: cid, Pid: pidVal, Name: nameTag, Text: cText,
					Timestamp: cTimestamp, Tag: cTag, Quote: quote, MediaIds: cMediaIds,
				}
				comments = append(comments, comment)
			}
		}
	}

	return posts, comments, nil
}
