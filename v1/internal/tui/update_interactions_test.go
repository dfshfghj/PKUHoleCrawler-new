package tui

import (
	"strings"
	"testing"

	"treehole/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

type stubPostsProvider struct {
	refreshPost          *models.Post
	listTags             []models.Tag
	togglePraisePID      int32
	toggleAttentionPID   int32
	createCommentPID     int32
	createCommentText    string
	createCommentQuoteID *int32
	refreshCalls         int
	canWrite             bool
	mode                 SessionMode
}

func (s *stubPostsProvider) ListPosts(cursor, limit, label int, keyword string) ([]models.Post, int, bool, error) {
	return nil, 0, false, nil
}

func (s *stubPostsProvider) GetPostDetail(pid int32, sortAsc bool) (*models.Post, []models.Comment, int32, bool, error) {
	return s.refreshPost, nil, 0, false, nil
}

func (s *stubPostsProvider) ListComments(pid int32, sortAsc bool, cursor int32) ([]models.Comment, int32, bool, error) {
	return nil, 0, false, nil
}

func (s *stubPostsProvider) SearchPosts(keyword string, cursor, limit, label int) ([]models.Post, int, bool, error) {
	return nil, 0, false, nil
}

func (s *stubPostsProvider) ListTags() ([]models.Tag, error) { return s.listTags, nil }

func (s *stubPostsProvider) RefreshPost(pid int32) (*models.Post, error) {
	s.refreshCalls++
	if s.refreshPost == nil {
		s.refreshPost = &models.Post{Pid: pid}
	}
	return s.refreshPost, nil
}

func (s *stubPostsProvider) TogglePraise(pid int32) error {
	s.togglePraisePID = pid
	return nil
}

func (s *stubPostsProvider) ToggleAttention(pid int32) error {
	s.toggleAttentionPID = pid
	return nil
}

func (s *stubPostsProvider) CreateComment(pid int32, text string, quoteID *int32) error {
	s.createCommentPID = pid
	s.createCommentText = text
	s.createCommentQuoteID = quoteID
	return nil
}

func (s *stubPostsProvider) CreatePost(text string) error { return nil }
func (s *stubPostsProvider) CanWrite() bool               { return s.canWrite }
func (s *stubPostsProvider) Mode() SessionMode            { return s.mode }

func TestBuildCommentContentMarksSelectedComment(t *testing.T) {
	page := NewPostsPageModel()
	page.CommentList = []models.Comment{
		{Cid: 1, Text: "First", Timestamp: 1100, NameTag: "user1"},
		{Cid: 2, Text: "Second", Timestamp: 1200, NameTag: "user2"},
	}
	page.SelectedCommentIdx = 1

	output := page.buildCommentContent(60)
	lines := strings.Split(output, "\n")
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "▸ ") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected selected comment marker in output, got %q", output)
	}
}

func TestHandlePostsKeyQuoteOpensComposerWithSelectedComment(t *testing.T) {
	m := newTestModel()
	m.Posts.ShowPostDetail = true
	m.Posts.CanWrite = true
	m.Posts.CurrentPost = &models.Post{Pid: 42}
	m.Posts.CommentList = []models.Comment{
		{Cid: 1, Text: "First", Timestamp: 1100, NameTag: "user1"},
		{Cid: 2, Text: "Second line", Timestamp: 1200, NameTag: "user2"},
	}
	m.Posts.SelectedCommentIdx = 1

	result, cmd := m.handlePostsKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatal("quote shortcut should not emit async command")
	}
	if result.Dialog != DialogComposer {
		t.Fatalf("dialog = %v, want composer", result.Dialog)
	}
	if result.Composer.Mode() != ComposerModeComment {
		t.Fatalf("composer mode = %v, want comment", result.Composer.Mode())
	}
	quote := result.Composer.QuoteTarget()
	if quote == nil || quote.Cid != 2 {
		t.Fatalf("quote target = %+v, want selected comment #2", quote)
	}
	if !strings.Contains(result.Composer.View(80), "引用 #2 user2: Second line") {
		t.Fatalf("composer view missing quote preview: %q", result.Composer.View(80))
	}
}

func TestHandleTagsDialogKeyTwoLevelSelectionAppliesChild(t *testing.T) {
	m := newTestModel()
	m.Dialog = DialogTags
	m.TagsDialog.SetTags([]models.Tag{
		{ID: 1, Name: "课程", ParentID: 0},
		{ID: 11, Label: "课程心得", ParentID: 1},
		{ID: 12, Label: "课程吐槽", ParentID: 1},
	})

	result, cmd := m.handleTagsDialogKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("entering child tag phase should not trigger load command")
	}
	if result.Dialog != DialogTags {
		t.Fatalf("dialog after entering child phase = %v, want tags", result.Dialog)
	}
	if result.Posts.ActiveTagID != 0 {
		t.Fatalf("active tag changed too early: %d", result.Posts.ActiveTagID)
	}

	result, _ = result.handleTagsDialogKey(tea.KeyMsg{Type: tea.KeyDown})
	result, cmd = result.handleTagsDialogKey(tea.KeyMsg{Type: tea.KeyEnter})
	if result.Dialog != DialogNone {
		t.Fatalf("dialog after applying child tag = %v, want none", result.Dialog)
	}
	if result.Posts.ActiveTagID != 12 {
		t.Fatalf("active tag = %d, want 12", result.Posts.ActiveTagID)
	}
	if result.Posts.ActiveTag != "课程吐槽" {
		t.Fatalf("active tag label = %q, want %q", result.Posts.ActiveTag, "课程吐槽")
	}
	if cmd == nil {
		t.Fatal("applying child tag should trigger reload command")
	}
}

func TestTogglePraiseCmdRefreshesPost(t *testing.T) {
	provider := &stubPostsProvider{refreshPost: &models.Post{Pid: 7}}

	msg := togglePraiseCmd(provider, 7)()
	result, ok := msg.(ActionResultMsg)
	if !ok {
		t.Fatalf("message type = %T, want ActionResultMsg", msg)
	}
	if provider.togglePraisePID != 7 {
		t.Fatalf("toggle praise pid = %d, want 7", provider.togglePraisePID)
	}
	if provider.refreshCalls != 1 {
		t.Fatalf("refresh calls = %d, want 1", provider.refreshCalls)
	}
	if result.Post == nil || result.Post.Pid != 7 {
		t.Fatalf("result post = %+v, want refreshed post #7", result.Post)
	}
}

func TestCreateCommentCmdPassesQuoteID(t *testing.T) {
	provider := &stubPostsProvider{}
	quote := &models.Comment{Cid: 456}

	msg := createCommentCmd(provider, 99, "hello", quote)()
	result, ok := msg.(ActionResultMsg)
	if !ok {
		t.Fatalf("message type = %T, want ActionResultMsg", msg)
	}
	if provider.createCommentPID != 99 || provider.createCommentText != "hello" {
		t.Fatalf("create comment payload = pid:%d text:%q", provider.createCommentPID, provider.createCommentText)
	}
	if provider.createCommentQuoteID == nil || *provider.createCommentQuoteID != 456 {
		t.Fatalf("quote id = %+v, want 456", provider.createCommentQuoteID)
	}
	if result.Error != nil || result.Kind != "comment" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
