package pages

import (
	"fmt"
	"strings"
	"time"

	"treehole/internal/models"
)

type PostsModel struct {
	Width        int
	Height       int
	Posts        []models.Post
	PostsTotal   int
	PostsPage    int
	PostsPerPage int
	PostsCursor  int
	PostsLoading bool

	ShowPostDetail bool
	CurrentPost    *models.Post
	Comments       []models.Comment
	CommentsCursor int

	Searching    bool
	SearchInput  string
	SearchActive bool
	SearchTotal  int

	LastError string
}

func NewPostsModel() PostsModel {
	return PostsModel{
		PostsPerPage: 10,
		PostsPage:    1,
	}
}

func (m PostsModel) View() string {
	if m.ShowPostDetail {
		return m.postDetailView()
	}

	var b strings.Builder

	if m.SearchActive {
		b.WriteString(pTitleStyle.Render(fmt.Sprintf("搜索结果: %s", m.SearchInput)))
	} else {
		b.WriteString(pTitleStyle.Render("帖子列表"))
	}
	b.WriteString("\n\n")

	searchLabel := "按 / 搜索"
	if m.Searching {
		searchLabel = "输入关键词 (Enter搜索, Esc取消): " + m.SearchInput
		b.WriteString(pSearchInputFocused.Render(searchLabel))
	} else {
		b.WriteString(pSearchInput.Render(searchLabel))
	}
	b.WriteString("\n")

	if m.PostsLoading {
		b.WriteString("\n")
		b.WriteString(pLoadingStyle.Render("加载中..."))
		return b.String()
	}

	if m.LastError != "" {
		b.WriteString("\n")
		b.WriteString(pErrorStyle.Render("错误: " + m.LastError))
	}

	if len(m.Posts) == 0 {
		b.WriteString("\n")
		b.WriteString(pEmptyStyle.Render("暂无数据"))
	} else {
		for i, post := range m.Posts {
			style := pListItemStyle
			if i == m.PostsCursor {
				style = pListItemSelectedStyle
			}

			text := post.Text
			if len(text) > 80 {
				text = text[:80] + "..."
			}

			ts := time.Unix(int64(post.Timestamp), 0).Format("2006-01-02 15:04")

			line := fmt.Sprintf("#%-6d %-10s %s", post.Pid, ts, text)
			if post.Anonymous == 0 {
				line = fmt.Sprintf("#%-6d %-10s [实名] %s", post.Pid, ts, text)
			}

			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	if m.PostsTotal > 0 {
		totalPages := (m.PostsTotal + m.PostsPerPage - 1) / m.PostsPerPage
		b.WriteString("\n")
		b.WriteString(pPaginationStyle.Render(
			fmt.Sprintf("第 %d/%d 页 (共 %d 条) | ←→翻页 | ↑↓选择 | Enter查看 | /搜索",
				m.PostsPage, totalPages, m.PostsTotal),
		))
	}

	b.WriteString("\n")
	b.WriteString(pHelpStyle.Render("Tab: 切换页面 | q: 退出"))

	return b.String()
}

func (m PostsModel) postDetailView() string {
	var b strings.Builder

	if m.CurrentPost == nil {
		return "无帖子数据"
	}

	ts := time.Unix(int64(m.CurrentPost.Timestamp), 0).Format("2006-01-02 15:04")
	b.WriteString(pPostPidStyle.Render(fmt.Sprintf("#%d", m.CurrentPost.Pid)))
	b.WriteString("  ")
	b.WriteString(pPostMetaStyle.Render(ts))
	b.WriteString(fmt.Sprintf("  回复: %d  点赞: %d", m.CurrentPost.Reply, m.CurrentPost.Likenum))
	if m.CurrentPost.Tag != "" {
		b.WriteString("  标签: " + m.CurrentPost.Tag)
	}
	b.WriteString("\n\n")

	b.WriteString(pPostTextStyle.Render(m.CurrentPost.Text))
	b.WriteString("\n\n")

	b.WriteString(pDividerStyle.Render(strings.Repeat("─", 60)))
	b.WriteString("\n\n")

	b.WriteString(pSubtitleStyle.Render(fmt.Sprintf("评论 (%d):", len(m.Comments))))
	b.WriteString("\n\n")

	if len(m.Comments) == 0 {
		b.WriteString(pEmptyStyle.Render("暂无评论"))
	} else {
		for i, c := range m.Comments {
			style := pListItemStyle
			if i == m.CommentsCursor {
				style = pListItemSelectedStyle
			}

			cName := c.Name
			if cName == "" {
				cName = "匿名"
			}
			cTs := time.Unix(int64(c.Timestamp), 0).Format("15:04")
			cText := c.Text
			if len(cText) > 100 {
				cText = cText[:100] + "..."
			}

			line := fmt.Sprintf("%s %s: %s", cTs, cName, cText)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(pHelpStyle.Render("Esc: 返回列表 | ↑↓: 选择评论 | q: 退出"))

	return b.String()
}
