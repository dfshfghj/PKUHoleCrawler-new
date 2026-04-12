package tui

import (
	"fmt"
	"strconv"
	"strings"

	"treehole/internal/client"
	"treehole/internal/db"
	"treehole/internal/models"
)

type PostsProvider interface {
	ListPosts(cursor, limit, label int, keyword string) ([]models.Post, int, bool, error)
	GetPostDetail(pid int32, sortAsc bool) (*models.Post, []models.Comment, int32, bool, error)
	ListComments(pid int32, sortAsc bool, cursor int32) ([]models.Comment, int32, bool, error)
	SearchPosts(keyword string, cursor, limit, label int) ([]models.Post, int, bool, error)
	ListTags() ([]models.Tag, error)
	RefreshPost(pid int32) (*models.Post, error)
	TogglePraise(pid int32) error
	ToggleAttention(pid int32) error
	CreateComment(pid int32, text string, quoteID *int32) error
	CreatePost(text string) error
	CanWrite() bool
	Mode() SessionMode
}

type OfflinePostsProvider struct {
	database *db.Database
}

func NewOfflinePostsProvider(database *db.Database) *OfflinePostsProvider {
	return &OfflinePostsProvider{database: database}
}

func (p *OfflinePostsProvider) ListPosts(cursor, limit, label int, keyword string) ([]models.Post, int, bool, error) {
	if keyword != "" {
		return p.SearchPosts(keyword, cursor, limit, label)
	}
	if label != 0 {
		return nil, 0, false, fmt.Errorf("离线模式暂不支持标签筛选")
	}
	posts, err := p.database.GetPostsCursor(cursor, limit, false)
	if err != nil {
		return nil, 0, false, err
	}
	next := nextPostCursor(posts)
	return posts, next, len(posts) == limit, nil
}

func (p *OfflinePostsProvider) SearchPosts(keyword string, cursor, limit, label int) ([]models.Post, int, bool, error) {
	if label != 0 {
		return nil, 0, false, fmt.Errorf("离线模式暂不支持标签筛选")
	}
	posts, err := p.database.SearchPostsCursor(keyword, cursor, limit, false)
	if err != nil {
		return nil, 0, false, err
	}
	next := nextPostCursor(posts)
	return posts, next, len(posts) == limit, nil
}

func (p *OfflinePostsProvider) GetPostDetail(pid int32, sortAsc bool) (*models.Post, []models.Comment, int32, bool, error) {
	post, err := p.database.GetPostByPid(pid)
	if err != nil {
		return nil, nil, 0, false, err
	}
	comments, err := p.database.GetCommentsByPidCursor(pid, 0, 50, sortAsc)
	if err != nil {
		return nil, nil, 0, false, err
	}
	next := nextCommentCursor(comments)
	return post, comments, next, len(comments) == 50, nil
}

func (p *OfflinePostsProvider) ListComments(pid int32, sortAsc bool, cursor int32) ([]models.Comment, int32, bool, error) {
	comments, err := p.database.GetCommentsByPidCursor(pid, cursor, 50, sortAsc)
	if err != nil {
		return nil, 0, false, err
	}
	next := nextCommentCursor(comments)
	return comments, next, len(comments) == 50, nil
}

func (p *OfflinePostsProvider) ListTags() ([]models.Tag, error) {
	return nil, fmt.Errorf("离线模式暂不支持标签读取")
}

func (p *OfflinePostsProvider) RefreshPost(pid int32) (*models.Post, error) {
	return p.database.GetPostByPid(pid)
}

func (p *OfflinePostsProvider) TogglePraise(pid int32) error {
	return fmt.Errorf("离线模式不支持点赞")
}

func (p *OfflinePostsProvider) ToggleAttention(pid int32) error {
	return fmt.Errorf("离线模式不支持关注")
}

func (p *OfflinePostsProvider) CreateComment(pid int32, text string, quoteID *int32) error {
	return fmt.Errorf("离线模式不支持发评论")
}

func (p *OfflinePostsProvider) CreatePost(text string) error {
	return fmt.Errorf("离线模式不支持发帖")
}

func (p *OfflinePostsProvider) CanWrite() bool    { return false }
func (p *OfflinePostsProvider) Mode() SessionMode { return SessionModeOffline }

type OnlinePostsProvider struct {
	client *client.Client
}

func NewOnlinePostsProvider(c *client.Client) *OnlinePostsProvider {
	return &OnlinePostsProvider{client: c}
}

func (p *OnlinePostsProvider) ListPosts(cursor, limit, label int, keyword string) ([]models.Post, int, bool, error) {
	page := cursorToPage(cursor)
	pid, plainKeyword := splitPIDSearch(keyword)
	posts, total, err := p.client.ListPostsV3(client.V3ListPostsParams{
		Page:          page,
		Limit:         limit,
		CommentLimit:  10,
		CommentStream: 1,
		Keyword:       plainKeyword,
		Label:         label,
		Pid:           pid,
	})
	if err != nil {
		return nil, 0, false, err
	}
	hasMore := page*limit < total
	return posts, page, hasMore, nil
}

func (p *OnlinePostsProvider) SearchPosts(keyword string, cursor, limit, label int) ([]models.Post, int, bool, error) {
	return p.ListPosts(cursor, limit, label, keyword)
}

func (p *OnlinePostsProvider) GetPostDetail(pid int32, sortAsc bool) (*models.Post, []models.Comment, int32, bool, error) {
	post, err := p.client.GetPostGet(pid)
	if err != nil {
		return nil, nil, 0, false, err
	}
	sort := 0
	if !sortAsc {
		sort = 1
	}
	comments, total, err := p.client.ListCommentsV3(pid, 1, 50, sort, 1)
	if err != nil {
		return nil, nil, 0, false, err
	}
	return post, comments, 1, 50 < total, nil
}

func (p *OnlinePostsProvider) ListComments(pid int32, sortAsc bool, cursor int32) ([]models.Comment, int32, bool, error) {
	page := commentCursorToPage(cursor)
	sort := 0
	if !sortAsc {
		sort = 1
	}
	comments, total, err := p.client.ListCommentsV3(pid, page, 50, sort, 1)
	if err != nil {
		return nil, 0, false, err
	}
	return comments, int32(page), page*50 < total, nil
}

func (p *OnlinePostsProvider) ListTags() ([]models.Tag, error) {
	return p.client.GetTagsTreeV3()
}

func (p *OnlinePostsProvider) RefreshPost(pid int32) (*models.Post, error) {
	return p.client.GetPostGet(pid)
}

func (p *OnlinePostsProvider) TogglePraise(pid int32) error {
	return p.client.TogglePraiseV3(pid)
}

func (p *OnlinePostsProvider) ToggleAttention(pid int32) error {
	return p.client.ToggleAttentionV3(pid)
}

func (p *OnlinePostsProvider) CreateComment(pid int32, text string, quoteID *int32) error {
	_, err := p.client.CreateCommentV3WithQuote(pid, text, quoteID)
	return err
}

func (p *OnlinePostsProvider) CreatePost(text string) error {
	_, err := p.client.CreatePostV3(client.CreatePostPayload{Type: "text", Kind: 0, RewardCost: 0, Text: text})
	return err
}

func (p *OnlinePostsProvider) CanWrite() bool {
	status := p.client.ProbeSession()
	return status.CanWriteOnline
}

func (p *OnlinePostsProvider) Mode() SessionMode { return SessionModeOnline }

func splitPIDSearch(keyword string) (int32, string) {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" || !strings.HasPrefix(trimmed, "#") {
		return 0, trimmed
	}
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return 0, trimmed
	}
	pidText := strings.TrimPrefix(parts[0], "#")
	if pidText == "" {
		return 0, trimmed
	}
	pid, err := strconv.Atoi(pidText)
	if err != nil {
		return 0, trimmed
	}
	return int32(pid), strings.Join(parts[1:], " ")
}

func cursorToPage(cursor int) int {
	if cursor <= 0 {
		return 1
	}
	return cursor + 1
}

func commentCursorToPage(cursor int32) int {
	if cursor <= 0 {
		return 1
	}
	return int(cursor) + 1
}
