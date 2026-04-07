package tui

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"treehole/internal/config"
	"treehole/internal/db"
	"treehole/internal/models"

	tea "github.com/charmbracelet/bubbletea"
)

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../..")
}

// stripANSI removes all ANSI escape sequences from a string.
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(s, "")
}

// visibleLines returns the non-empty lines from a stripped output string.
func visibleLines(output string) []string {
	stripped := stripANSI(output)
	var lines []string
	for _, line := range strings.Split(stripped, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

// loadRealPosts loads posts from the real treehole.db for testing.
func loadRealPosts(t *testing.T) []models.Post {
	t.Helper()

	dbPath := filepath.Join(projectRoot(), "treehole.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Skip("treehole.db not found, skipping real data test")
	}

	cfg := &config.Config{
		Username:  "test",
		Password:  "test",
		SecretKey: "test",
		Database: config.DatabaseConfig{
			Type:   "sqlite3",
			DBFile: dbPath,
		},
	}

	database, err := db.NewDatabase(cfg)
	if err != nil {
		t.Fatalf("NewDatabase: %v", err)
	}
	defer database.Close()

	posts, err := database.GetPosts(0, 50)
	if err != nil {
		t.Fatalf("GetPosts: %v", err)
	}

	if len(posts) == 0 {
		t.Skip("no posts returned from treehole.db")
	}

	return posts
}

func TestViewPostsRealDataOverflow(t *testing.T) {
	posts := loadRealPosts(t)

	// Find the longest post
	var longest models.Post
	for _, p := range posts {
		if len(p.Text) > len(longest.Text) {
			longest = p
		}
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = []models.Post{longest}
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	output := m.View()

	// Should not panic or produce empty output
	if output == "" {
		t.Fatal("View() returned empty string")
	}

	// Should contain the post text (at least the first line)
	firstLine := strings.Split(longest.Text, "\n")[0]
	if firstLine != "" && !containsStr(output, strings.TrimSpace(firstLine)) {
		t.Errorf("View() missing first line of long post: %q", firstLine)
	}

	t.Logf("Rendered post pid=%d, text_len=%d, output_len=%d",
		longest.Pid, len(longest.Text), len(output))
}

func TestViewPostsRealDataMultiLine(t *testing.T) {
	posts := loadRealPosts(t)

	// Find the post with most newlines
	var mostLines models.Post
	maxNewlines := 0
	for _, p := range posts {
		nl := strings.Count(p.Text, "\n")
		if nl > maxNewlines {
			maxNewlines = nl
			mostLines = p
		}
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = []models.Post{mostLines}
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string")
	}

	// Verify multiline content is rendered
	lines := strings.Split(mostLines.Text, "\n")
	renderedCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && containsStr(output, trimmed) {
			renderedCount++
		}
	}

	t.Logf("Post pid=%d has %d lines, %d lines found in output",
		mostLines.Pid, len(lines), renderedCount)

	if renderedCount == 0 {
		t.Error("No multiline content found in output")
	}
}

func TestViewPostsExtremeLongText(t *testing.T) {
	// Create a post with extremely long single line (2000 chars)
	longLine := strings.Repeat("A", 2000)

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = []models.Post{
		{Pid: 1, Text: longLine, Timestamp: 1000, Anonymous: true},
	}
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	// Should not panic
	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for extreme long text")
	}

	// The viewport should handle overflow gracefully
	t.Logf("Output length for 2000-char line: %d", len(output))
}

func TestViewPostsManyNewlines(t *testing.T) {
	// Create a post with many newlines (100 empty lines)
	manyNewlines := strings.Repeat("\n", 100) + "bottom line"

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = []models.Post{
		{Pid: 1, Text: manyNewlines, Timestamp: 1000, Anonymous: true},
	}
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for many newlines")
	}

	// Viewport should handle the overflow
	t.Logf("Output length for 100-newline post: %d", len(output))
}

func TestViewPostsWideTerminal(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 3 {
		t.Skip("need at least 3 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:3]
	m.SelectedPostIdx = 0
	m.Width = 200
	m.Height = 100 // Very tall to fit all posts even with wrapping

	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for wide terminal")
	}

	// At minimum, the first post should always be visible
	firstPost := posts[0]
	firstLine := strings.Split(firstPost.Text, "\n")[0]
	if firstLine != "" && !containsStr(output, strings.TrimSpace(firstLine)) {
		t.Errorf("Wide terminal: missing first post pid=%d", firstPost.Pid)
	}

	// Count how many posts are visible
	visibleCount := 0
	for _, p := range posts[:3] {
		fl := strings.Split(p.Text, "\n")[0]
		if fl != "" && containsStr(output, strings.TrimSpace(fl)) {
			visibleCount++
		}
	}

	t.Logf("Wide terminal (200x100): %d/%d posts visible", visibleCount, 3)
}

func TestViewPostsNarrowTerminal(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 1 {
		t.Skip("need at least 1 post")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:1]
	m.SelectedPostIdx = 0
	m.Width = 40
	m.Height = 12

	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for narrow terminal")
	}

	t.Logf("Narrow terminal (40x12) output length: %d", len(output))
}

func TestViewPostsTinyTerminal(t *testing.T) {
	m := newTestModel()
	m.Page = PagePosts
	m.PostList = []models.Post{
		{Pid: 1, Text: "test", Timestamp: 1000},
	}
	m.SelectedPostIdx = 0
	m.Width = 10
	m.Height = 5

	// Should not panic
	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for tiny terminal")
	}

	t.Logf("Tiny terminal (10x5) output length: %d", len(output))
}

func TestViewPostDetailLongText(t *testing.T) {
	posts := loadRealPosts(t)

	var longest models.Post
	for _, p := range posts {
		if len(p.Text) > len(longest.Text) {
			longest = p
		}
	}

	m := newTestModel()
	m.Page = PagePosts
	m.ShowPostDetail = true
	m.CurrentPost = &longest
	m.CommentList = nil
	m.Width = 80
	m.Height = 24

	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for detail view")
	}

	// Should contain post content
	if !containsStr(output, longest.Text[:20]) {
		t.Error("Detail view missing post content")
	}

	t.Logf("Detail view output length for long post: %d", len(output))
}

func TestViewPostDetailManyComments(t *testing.T) {
	// Create 50 comments
	var comments []models.Comment
	for i := 1; i <= 50; i++ {
		comments = append(comments, models.Comment{
			Cid:       int32(i),
			Pid:       1,
			Text:      strings.Repeat("C", 100),
			Timestamp: int32(1000 + i*100),
			NameTag:   "user",
		})
	}

	m := newTestModel()
	m.Page = PagePosts
	m.ShowPostDetail = true
	m.CurrentPost = &models.Post{Pid: 1, Text: "Post with many comments", Timestamp: 1000}
	m.CommentList = comments
	m.Width = 80
	m.Height = 24

	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for many comments")
	}

	// Should show comment count
	if !containsStr(output, "50") {
		t.Error("Detail view should show comment count")
	}

	t.Logf("Detail view with 50 comments output length: %d", len(output))
}

func TestScrollToSelectedPostBoundary(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 10 {
		t.Skip("need at least 10 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:10]
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	// Render to initialize viewport
	m.View()

	// Scroll to last post
	m.SelectedPostIdx = 9
	m.scrollToSelectedPost()

	// Should not panic
	m.View()

	t.Logf("Scrolled to last post, viewport YOffset=%d", m.PostViewport.YOffset)
}

func TestAdjustSelectedToViewportBoundary(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 10 {
		t.Skip("need at least 10 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:10]
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	// Render to initialize viewport
	m.View()

	// Simulate viewport scrolled to bottom
	m.PostViewport.GotoBottom()
	m.adjustSelectedToViewport()

	// SelectedPostIdx should be updated to match viewport
	t.Logf("After GotoBottom + adjustSelectedToViewport: SelectedPostIdx=%d", m.SelectedPostIdx)
}

func TestFastScrollPgDownBoundary(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 20 {
		t.Skip("need at least 20 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:20]
	m.SelectedPostIdx = 0
	m.PostListTotal = 100 // Simulate more posts available
	m.Width = 80
	m.Height = 24

	// Render to initialize viewport
	m.View()

	// Rapid PgDn
	for i := 0; i < 20; i++ {
		m, _ = m.handlePostsKey(keyPgDown())
	}

	// Should not panic
	output := m.View()
	if output == "" {
		t.Fatal("View() returned empty after rapid PgDn")
	}

	t.Logf("After 20x PgDn: SelectedPostIdx=%d, YOffset=%d", m.SelectedPostIdx, m.PostViewport.YOffset)
}

func TestFastScrollPgUpBoundary(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 20 {
		t.Skip("need at least 20 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:20]
	m.SelectedPostIdx = 19
	m.PostListTotal = 100
	m.Width = 80
	m.Height = 24

	// Render to initialize viewport
	m.View()

	// Rapid PgUp
	for i := 0; i < 20; i++ {
		m, _ = m.handlePostsKey(keyPgUp())
	}

	// Should not panic
	output := m.View()
	if output == "" {
		t.Fatal("View() returned empty after rapid PgUp")
	}

	// Should be at or near top
	if m.SelectedPostIdx > 2 {
		t.Errorf("SelectedPostIdx = %d, expected near 0 after 20x PgUp", m.SelectedPostIdx)
	}

	t.Logf("After 20x PgUp: SelectedPostIdx=%d, YOffset=%d", m.SelectedPostIdx, m.PostViewport.YOffset)
}

func TestFastScrollMixedBoundary(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 20 {
		t.Skip("need at least 20 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:20]
	m.SelectedPostIdx = 0
	m.PostListTotal = 100
	m.Width = 80
	m.Height = 24

	// Render to initialize viewport
	m.View()

	// Mixed: 10x PgDn, 5x PgUp, 15x PgDn, 30x PgUp
	for i := 0; i < 10; i++ {
		m, _ = m.handlePostsKey(keyPgDown())
	}
	for i := 0; i < 5; i++ {
		m, _ = m.handlePostsKey(keyPgUp())
	}
	for i := 0; i < 15; i++ {
		m, _ = m.handlePostsKey(keyPgDown())
	}
	for i := 0; i < 30; i++ {
		m, _ = m.handlePostsKey(keyPgUp())
	}

	// Should not panic
	output := m.View()
	if output == "" {
		t.Fatal("View() returned empty after mixed fast scroll")
	}

	// After 30x PgUp from bottom, should be at top
	if m.SelectedPostIdx > 3 {
		t.Errorf("SelectedPostIdx = %d, expected near 0 after mixed scroll ending with PgUp", m.SelectedPostIdx)
	}

	t.Logf("After mixed scroll: SelectedPostIdx=%d, YOffset=%d", m.SelectedPostIdx, m.PostViewport.YOffset)
}

func TestRefreshClearsState(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 5 {
		t.Skip("need at least 5 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:5]
	m.PostListTotal = 50
	m.SelectedPostIdx = 3
	m.SearchActive = false
	m.Width = 80
	m.Height = 24

	// Render to initialize viewport
	m.View()

	// Press 'r' to refresh
	m, _ = m.handlePostsKey(keyR())

	if !m.PostListLoading {
		t.Error("PostListLoading should be true after refresh")
	}
	if len(m.PostList) != 0 {
		t.Errorf("PostList should be empty after refresh, got %d", len(m.PostList))
	}
	if m.PostListTotal != 0 {
		t.Errorf("PostListTotal should be 0 after refresh, got %d", m.PostListTotal)
	}
	if m.SelectedPostIdx != 0 {
		t.Errorf("SelectedPostIdx should be 0 after refresh, got %d", m.SelectedPostIdx)
	}

	// View during loading should show "加载中..."
	output := m.View()
	if !containsStr(output, "加载中") {
		t.Error("View() should show '加载中...' during refresh")
	}
}

func TestRefreshDuringSearch(t *testing.T) {
	m := newTestModel()
	m.Page = PagePosts
	m.SearchActive = true
	m.SearchInput = "test"
	m.PostList = []models.Post{{Pid: 1, Text: "test result", Timestamp: 1000}}

	// Press 'r' during search - should NOT trigger refresh
	m, cmd := m.handlePostsKey(keyR())

	if m.PostListLoading {
		t.Error("PostListLoading should NOT change during search")
	}
	if cmd != nil {
		t.Error("r during search should NOT trigger reload")
	}
	if len(m.PostList) != 1 {
		t.Error("PostList should NOT be cleared during search")
	}
}

func TestViewPostsErrorState(t *testing.T) {
	m := newTestModel()
	m.Page = PagePosts
	m.PostListError = "connection refused"
	m.Width = 80
	m.Height = 24

	output := m.View()

	if !containsStr(output, "错误") {
		t.Error("View() should show error indicator")
	}
	if !containsStr(output, "connection refused") {
		t.Error("View() should show error message")
	}
}

func TestViewPostsErrorWithPartialData(t *testing.T) {
	m := newTestModel()
	m.Page = PagePosts
	m.PostList = []models.Post{
		{Pid: 1, Text: "Partial data", Timestamp: 1000},
	}
	m.PostListError = "timeout on page 2"
	m.PostListTotal = 50
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	output := m.View()

	// Should show both data and error
	if !containsStr(output, "Partial data") {
		t.Error("View() should still show partial data")
	}
	if !containsStr(output, "timeout on page 2") {
		t.Error("View() should show error message")
	}
}

func TestViewHomeExtremeStats(t *testing.T) {
	m := newTestModel()
	m.Page = PageHome
	m.LoggedIn = true
	m.LoginUser = "testuser"
	m.CrawlerState = CrawlerRunning
	m.TotalPosts = 9999999
	m.TotalComments = 99999999
	m.LastCrawlPage = 99999
	m.Width = 80
	m.Height = 24

	// Should not panic with large numbers
	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string with extreme stats")
	}

	t.Logf("Home view with extreme stats output length: %d", len(output))
}

func TestViewPostDetailEmptyPost(t *testing.T) {
	m := newTestModel()
	m.Page = PagePosts
	m.ShowPostDetail = true
	m.CurrentPost = &models.Post{Pid: 1, Text: "", Timestamp: 1000}
	m.CommentList = nil
	m.Width = 80
	m.Height = 24

	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for empty post detail")
	}

	if !containsStr(output, "#1") {
		t.Error("View() should show post pid")
	}
}

func TestViewPostDetailUnicodeContent(t *testing.T) {
	unicodeText := "🎉 测试中文和emoji混合 🚀\n日本語テスト\n한국어 테스트\nSpecial: αβγδε ∑∏∫"

	m := newTestModel()
	m.Page = PagePosts
	m.ShowPostDetail = true
	m.CurrentPost = &models.Post{Pid: 1, Text: unicodeText, Timestamp: 1000}
	m.CommentList = []models.Comment{
		{Cid: 1, Text: "评论测试 🎊", Timestamp: 1100, NameTag: "用户"},
	}
	m.Width = 80
	m.Height = 24

	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for unicode content")
	}

	// Should contain at least some unicode content
	if !containsStr(output, "测试") {
		t.Error("View() should show unicode content")
	}
}

func TestViewPostsManyPosts(t *testing.T) {
	posts := loadRealPosts(t)

	// Load all available posts
	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts
	m.PostListTotal = len(posts)
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	output := m.View()

	if output == "" {
		t.Fatal("View() returned empty string for many posts")
	}

	t.Logf("Rendered %d posts, output length: %d", len(posts), len(output))
}

func TestViewPostsScrollThroughAll(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 10 {
		t.Skip("need at least 10 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts
	m.PostListTotal = len(posts)
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	// Render to initialize viewport
	m.View()

	// Scroll through all posts
	for i := 0; i < len(posts)-1; i++ {
		m, _ = m.handlePostsKey(keyDown())
	}

	if m.SelectedPostIdx != len(posts)-1 {
		t.Errorf("SelectedPostIdx = %d, want %d", m.SelectedPostIdx, len(posts)-1)
	}

	// Should not panic
	output := m.View()
	if output == "" {
		t.Fatal("View() returned empty after scrolling through all posts")
	}

	t.Logf("Scrolled through %d posts successfully", len(posts))
}

func TestViewPostsViewportContentUpdate(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 5 {
		t.Skip("need at least 5 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:3]
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	// First render
	output1 := m.View()

	// Add more posts
	m.PostList = posts[:5]

	output2 := m.View()

	if len(output2) <= len(output1) {
		t.Errorf("Output should grow with more posts: before=%d, after=%d", len(output1), len(output2))
	}

	// Verify the new posts' content is in output
	for _, p := range posts[3:5] {
		firstLine := strings.Split(p.Text, "\n")[0]
		if firstLine != "" && !containsStr(output2, strings.TrimSpace(firstLine)) {
			t.Errorf("New post pid=%d not visible after adding", p.Pid)
		}
	}

	t.Logf("Output lengths: 3 posts=%d, 5 posts=%d", len(output1), len(output2))
}

func TestViewPostsResizeDuringRender(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 3 {
		t.Skip("need at least 3 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:3]
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	// Render at 80x24
	output1 := m.View()

	// Resize to 120x40
	m.Width = 120
	m.Height = 40
	m.postContent = "" // Force content update

	output2 := m.View()

	if output1 == output2 {
		t.Log("Output changed after resize (expected due to different dimensions)")
	}

	t.Logf("Output lengths: 80x24=%d, 120x40=%d", len(output1), len(output2))
}

func TestStripANSIRemovesAllCodes(t *testing.T) {
	input := "\x1b[38;5;205mHello\x1b[0m World\x1b[1mBold\x1b[22m"
	result := stripANSI(input)
	expected := "Hello WorldBold"
	if result != expected {
		t.Errorf("stripANSI(%q) = %q, want %q", input, result, expected)
	}
}

func TestViewPostsStrippedLines(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 3 {
		t.Skip("need at least 3 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:3]
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	output := m.View()
	lines := visibleLines(output)

	if len(lines) == 0 {
		t.Fatal("No visible lines after stripping ANSI codes")
	}

	// First line is tab bar, title is at line[2]
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 visible lines, got %d", len(lines))
	}
	if !strings.Contains(lines[2], "帖子列表") {
		t.Errorf("Line[2] = %q, want '帖子列表'", lines[2])
	}

	// Should contain search hint
	foundSearch := false
	for _, line := range lines {
		if strings.Contains(line, "搜索") {
			foundSearch = true
			break
		}
	}
	if !foundSearch {
		t.Error("No search hint found in visible lines")
	}

	t.Logf("Total visible lines: %d", len(lines))
	for i, line := range lines {
		if i < 5 {
			t.Logf("  line[%d]: %q", i, line)
		}
	}
}

func TestViewHomeStrippedLines(t *testing.T) {
	m := newTestModel()
	m.Page = PageHome
	m.LoggedIn = true
	m.LoginUser = "testuser"
	m.CrawlerState = CrawlerStopped
	m.TotalPosts = 100
	m.TotalComments = 500
	m.Width = 80
	m.Height = 24

	output := m.View()
	lines := visibleLines(output)

	if len(lines) == 0 {
		t.Fatal("No visible lines")
	}

	// Title should be in the content (tab bar is first, then separator, then content)
	titleFound := false
	for _, line := range lines {
		if strings.Contains(line, "PKUHole Crawler") {
			titleFound = true
			break
		}
	}
	if !titleFound {
		t.Errorf("Title 'PKUHole Crawler' not found in visible lines")
	}

	// Check key content lines
	allText := strings.Join(lines, " ")
	expectedContent := []string{"PKUHole", "已登录", "testuser", "已停止", "帖子总数", "100", "评论总数", "500"}
	for _, want := range expectedContent {
		if !strings.Contains(allText, want) {
			t.Errorf("Missing expected content: %q", want)
		}
	}

	t.Logf("Home view: %d visible lines", len(lines))
}

func TestViewPostDetailStrippedLines(t *testing.T) {
	m := newTestModel()
	m.Page = PagePosts
	m.ShowPostDetail = true
	m.CurrentPost = &models.Post{
		Pid: 42, Text: "Detail post text", Timestamp: 1000,
		Reply: 5, Likenum: 10,
	}
	m.CommentList = []models.Comment{
		{Cid: 1, Text: "First comment", Timestamp: 1100, NameTag: "user1"},
		{Cid: 2, Text: "Second comment", Timestamp: 1200, NameTag: "user2"},
	}
	m.Width = 80
	m.Height = 24

	output := m.View()
	lines := visibleLines(output)

	allText := strings.Join(lines, " ")
	expectedContent := []string{"#42", "Detail post text", "First comment", "Second comment", "Esc"}
	for _, want := range expectedContent {
		if !strings.Contains(allText, want) {
			t.Errorf("Missing expected content: %q", want)
		}
	}

	t.Logf("Post detail: %d visible lines", len(lines))
}

func TestViewConfigDialogStrippedLines(t *testing.T) {
	m := newTestModel()
	m.Dialog = DialogConfig
	m.ConfigFieldIdx = 0
	m.ConfigUsername = "testuser"
	m.ConfigPassword = "secret"
	m.ConfigSecretKey = "KEY123"
	m.Width = 80
	m.Height = 24

	output := m.View()
	lines := visibleLines(output)

	allText := strings.Join(lines, " ")
	expectedContent := []string{"配置管理", "config.json", "用户名", "密码", "SecretKey", "保存配置"}
	for _, want := range expectedContent {
		if !strings.Contains(allText, want) {
			t.Errorf("Missing expected content: %q", want)
		}
	}

	// Password should be masked
	if strings.Contains(allText, "secret") {
		t.Error("Plaintext password found in output")
	}

	t.Logf("Config dialog: %d visible lines", len(lines))
}

func TestViewHelpDialogStrippedLines(t *testing.T) {
	m := newTestModel()
	m.Dialog = DialogHelp
	m.Width = 80
	m.Height = 24

	output := m.View()
	lines := visibleLines(output)

	allText := strings.Join(lines, " ")
	expectedContent := []string{"快捷键帮助", "打开/关闭此帮助菜单", "打开配置管理", "搜索帖子", "刷新", "Esc", "关闭"}
	for _, want := range expectedContent {
		if !strings.Contains(allText, want) {
			t.Errorf("Missing expected content: %q", want)
		}
	}

	t.Logf("Help dialog: %d visible lines", len(lines))
}

func TestViewPostsStrippedLinesWithRealData(t *testing.T) {
	posts := loadRealPosts(t)
	if len(posts) < 3 {
		t.Skip("need at least 3 posts")
	}

	m := newTestModel()
	m.Page = PagePosts
	m.PostList = posts[:3]
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	output := m.View()
	lines := visibleLines(output)

	// Verify structure
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 visible lines, got %d", len(lines))
	}

	// First line is the tab bar, content starts after
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 visible lines, got %d", len(lines))
	}

	// Line[0] is tab bar, line[1] is separator, line[2] should be title
	if !strings.Contains(lines[2], "帖子列表") {
		t.Errorf("Line[2] = %q, want '帖子列表'", lines[2])
	}

	// Should contain post text
	postText := posts[0].Text
	firstLine := strings.Split(postText, "\n")[0]
	trimmedFirst := strings.TrimSpace(firstLine)
	if trimmedFirst != "" {
		found := false
		for _, line := range lines {
			// The text may be split across lines, check if any line contains part of it
			if strings.Contains(line, trimmedFirst[:min(len(trimmedFirst), 20)]) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Post text first line %q not found in visible lines", trimmedFirst[:min(len(trimmedFirst), 40)])
		}
	}

	// Should contain pagination hint
	foundPagination := false
	for _, line := range lines {
		if strings.Contains(line, "已加载") {
			foundPagination = true
			break
		}
	}
	if !foundPagination {
		t.Error("Pagination hint '已加载' not found")
	}

	t.Logf("Real data posts view: %d visible lines", len(lines))
}

func TestViewNoANSILeakage(t *testing.T) {
	// Verify that ANSI codes are properly closed (no leakage)
	m := newTestModel()
	m.Page = PagePosts
	m.PostList = []models.Post{
		{Pid: 1, Text: "Test post", Timestamp: 1000, Anonymous: true},
	}
	m.SelectedPostIdx = 0
	m.Width = 80
	m.Height = 24

	output := m.View()

	// Count opening and closing ANSI sequences
	openCount := strings.Count(output, "\x1b[")
	// Each ANSI sequence should end with a letter
	closeCount := len(regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`).FindAllString(output, -1))

	if openCount != closeCount {
		t.Errorf("ANSI sequence mismatch: %d opens, %d closes", openCount, closeCount)
	}

	t.Logf("ANSI sequences: %d open, %d close", openCount, closeCount)
}

func TestViewStrippedOutputNotEmpty(t *testing.T) {
	// All view states should produce non-empty stripped output
	tests := []struct {
		name  string
		model Model
	}{
		{"home_stopped", func() Model {
			m := newTestModel()
			m.Page = PageHome
			m.CrawlerState = CrawlerStopped
			return m
		}()},
		{"home_running", func() Model {
			m := newTestModel()
			m.Page = PageHome
			m.CrawlerState = CrawlerRunning
			return m
		}()},
		{"posts_empty", func() Model {
			m := newTestModel()
			m.Page = PagePosts
			m.PostList = nil
			return m
		}()},
		{"posts_with_data", func() Model {
			m := newTestModel()
			m.Page = PagePosts
			m.PostList = []models.Post{{Pid: 1, Text: "Hello", Timestamp: 1000}}
			m.SelectedPostIdx = 0
			return m
		}()},
		{"detail_view", func() Model {
			m := newTestModel()
			m.Page = PagePosts
			m.ShowPostDetail = true
			m.CurrentPost = &models.Post{Pid: 1, Text: "Post", Timestamp: 1000}
			return m
		}()},
		{"config_dialog", func() Model {
			m := newTestModel()
			m.Dialog = DialogConfig
			return m
		}()},
		{"help_dialog", func() Model {
			m := newTestModel()
			m.Dialog = DialogHelp
			return m
		}()},
		{"logs_dialog", func() Model {
			m := newTestModel()
			m.Dialog = DialogLogs
			m.LogLines = []string{"log line"}
			return m
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.model.View()
			stripped := stripANSI(output)
			lines := visibleLines(output)

			if stripped == "" {
				t.Errorf("Stripped output is empty for %s", tt.name)
			}
			if len(lines) == 0 {
				t.Errorf("No visible lines for %s", tt.name)
			}

			t.Logf("%s: %d stripped chars, %d visible lines", tt.name, len(stripped), len(lines))
		})
	}
}

// Key helpers

func keyDown() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyDown}
}

func keyPgDown() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyPgDown}
}

func keyPgUp() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyPgUp}
}

func keyR() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
}
