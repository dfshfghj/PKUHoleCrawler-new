package tui

import (
	"fmt"
	"strings"
	"time"

	"treehole/internal/models"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type PostsPageModel struct {
	PostList        []models.Post
	PostListTotal   int
	PostListHasMore bool
	PostListCursor  int
	PostListLoading bool
	PostListError   string
	PostPerPage     int
	PostViewport    *viewport.Model
	postContent     string
	CursorLine      int
	SelectedPostIdx int

	ShowPostDetail     bool
	CurrentPost        *models.Post
	PostBodyViewport   *viewport.Model
	postBodyContent    string
	CommentList        []models.Comment
	CommentListHasMore bool
	CommentListLoading bool
	CommentListCursor  int32
	CommentListError   string
	CommentSortAsc     bool
	CommentViewport    *viewport.Model
	commentContent     string
	DetailFocus        DetailFocus

	PostsMode    PostsMode
	Searching    bool
	SearchInput  string
	SearchActive bool
}

func NewPostsPageModel() PostsPageModel {
	pv := viewport.New(0, 0)
	bv := viewport.New(0, 0)
	cv := viewport.New(0, 0)
	return PostsPageModel{
		PostPerPage:      20,
		PostViewport:     &pv,
		PostBodyViewport: &bv,
		CommentViewport:  &cv,
		CommentSortAsc:   true,
		PostsMode:        PostsModeList,
		DetailFocus:      DetailFocusComments,
	}
}

func (p *PostsPageModel) ensureInitialized() {
	if p.PostViewport == nil || p.PostBodyViewport == nil || p.CommentViewport == nil {
		*p = NewPostsPageModel()
	}
}

func (p *PostsPageModel) syncViewports(width, height int) {
	p.ensureInitialized()
	if width < 1 {
		width = 80
	}
	if height < 1 {
		height = 24
	}

	contentWidth := width - 8
	if contentWidth < 20 {
		contentWidth = 20
	}
	postHeight := p.calcPostViewportHeight(height)

	if !p.ShowPostDetail && len(p.PostList) > 0 {
		p.syncCursorToSelection()
		newContent := p.buildPostListContent(contentWidth)
		if p.postContent != newContent || p.PostViewport.Width != contentWidth || p.PostViewport.Height != postHeight {
			p.PostViewport.Width = contentWidth
			p.PostViewport.Height = postHeight
			p.PostViewport.SetContent(newContent)
			p.postContent = newContent
		}
	}

	if p.ShowPostDetail && p.CurrentPost != nil {
		bodyHeight, commentHeight := p.calcDetailViewportHeights(height)
		bodyContent := p.buildDetailBodyContent(contentWidth)
		if p.postBodyContent != bodyContent || p.PostBodyViewport.Width != contentWidth || p.PostBodyViewport.Height != bodyHeight {
			p.PostBodyViewport.Width = contentWidth
			p.PostBodyViewport.Height = bodyHeight
			p.PostBodyViewport.SetContent(bodyContent)
			p.postBodyContent = bodyContent
		}

		commentContent := p.buildCommentContent(contentWidth)
		if p.commentContent != commentContent || p.CommentViewport.Width != contentWidth || p.CommentViewport.Height != commentHeight {
			p.CommentViewport.Width = contentWidth
			p.CommentViewport.Height = commentHeight
			p.CommentViewport.SetContent(commentContent)
			p.commentContent = commentContent
		}
	}
}

func (p PostsPageModel) View(width, height int) string {
	p.ensureInitialized()
	if p.ShowPostDetail {
		return p.renderPostDetail(width, height)
	}
	return p.renderPosts(width, height)
}

func (p PostsPageModel) renderPosts(width, height int) string {
	var b strings.Builder

	if p.SearchActive {
		b.WriteString(vTitleStyle.Render(fmt.Sprintf("搜索结果: %s", p.SearchInput)))
	} else {
		b.WriteString(vTitleStyle.Render("帖子列表"))
	}
	b.WriteString("\n")

	pageWidth := maxInt(20, width-8)
	searchLabel := "按 / 搜索"
	searchStyle := vSearchInput.Width(maxInt(1, pageWidth-vSearchInput.GetHorizontalFrameSize()))
	searchFocusedStyle := vSearchInputFocused.Width(maxInt(1, pageWidth-vSearchInputFocused.GetHorizontalFrameSize()))
	if p.Searching {
		searchLabel = "输入关键词: " + p.SearchInput
		b.WriteString(searchFocusedStyle.Render(searchLabel))
	} else {
		b.WriteString(searchStyle.Render(searchLabel))
	}
	b.WriteString("\n")

	if p.PostListLoading && len(p.PostList) == 0 {
		b.WriteString(vLoadingStyle.Render("加载中..."))
		return b.String()
	}

	if p.PostListError != "" {
		b.WriteString(vErrorStyle.Render("错误: " + p.PostListError))
		b.WriteString("\n")
	}

	if len(p.PostList) == 0 {
		b.WriteString(vEmptyStyle.Render("暂无数据"))
		return b.String()
	}

	contentWidth := pageWidth
	vp := viewport.New(contentWidth, p.calcPostViewportHeight(height))
	vp.SetContent(p.buildPostListContent(contentWidth))
	if p.PostViewport != nil {
		vp.SetYOffset(p.PostViewport.YOffset)
	}
	b.WriteString(vp.View())
	b.WriteString("\n")
	status := fmt.Sprintf("↑↓: 选择 | Enter: 查看 | /: 搜索 | r: 刷新 | PgUp/PgDn: 快滚 | 已加载 %d", len(p.PostList))
	if p.PostListLoading {
		status += " | 正在加载更多..."
	}
	b.WriteString(vPaginationStyle.Render(status))
	return b.String()
}

func (p PostsPageModel) renderPostDetail(width, height int) string {
	var b strings.Builder

	if p.CurrentPost == nil {
		return "无帖子数据"
	}
	ts := time.Unix(int64(p.CurrentPost.Timestamp), 0).In(shanghaiLocation).Format("2006-01-02 15:04")
	b.WriteString(vPostPidStyle.Render(fmt.Sprintf("#%d", p.CurrentPost.Pid)))
	b.WriteString("  ")
	b.WriteString(vPostTimeStyle.Render(ts))
	b.WriteString("  ")
	b.WriteString(vPostReplyStyle.Render(fmt.Sprintf("回复: %d", p.CurrentPost.Reply)))
	b.WriteString("  ")
	b.WriteString(vPostLikeStyle.Render(fmt.Sprintf("点赞: %d", p.CurrentPost.Likenum)))
	b.WriteString("\n")

	dividerWidth := width - 8
	if dividerWidth < 20 {
		dividerWidth = 20
	}
	sortLabel := "正序"
	if !p.CommentSortAsc {
		sortLabel = "逆序"
	}
	commentsTitle := fmt.Sprintf("评论 %d  %s", len(p.CommentList), sortLabel)
	if p.CommentListLoading {
		commentsTitle += ", 加载中"
	}
	commentsTitleStyle := vSectionTitleStyle
	bodySectionStyle := vDetailSection
	commentsSectionStyle := vDetailSection
	if p.DetailFocus == DetailFocusPost {
		bodySectionStyle = vDetailSectionFocused
	} else {
		commentsTitleStyle = vSectionTitleFocused
		commentsSectionStyle = vDetailSectionFocused
	}

	contentWidth := width - 8
	if contentWidth < 20 {
		contentWidth = 20
	}
	bodyHeight, commentHeight := p.calcDetailViewportHeights(height)
	bodyViewport := viewport.New(contentWidth, bodyHeight)
	bodyViewport.SetContent(p.buildDetailBodyContent(contentWidth))
	if p.PostBodyViewport != nil {
		bodyViewport.SetYOffset(p.PostBodyViewport.YOffset)
	}
	b.WriteString(bodySectionStyle.Render(bodyViewport.View()))
	b.WriteString("\n")

	b.WriteString(vDividerStyle.Render(strings.Repeat("─", dividerWidth)))
	b.WriteString("\n")

	b.WriteString(commentsTitleStyle.Render(commentsTitle))
	b.WriteString("\n")

	if len(p.CommentList) == 0 {
		b.WriteString(commentsSectionStyle.Render(vEmptyStyle.Render("暂无评论")))
	} else {
		vp := viewport.New(contentWidth, commentHeight)
		vp.SetContent(p.buildCommentContent(contentWidth))
		if p.CommentViewport != nil {
			vp.SetYOffset(p.CommentViewport.YOffset)
		}
		b.WriteString(commentsSectionStyle.Render(vp.View()))
	}

	b.WriteString("\n")
	b.WriteString(vPaginationStyle.Render("Tab: 切换正文/评论 | s: 正序/逆序 | Esc: 返回列表 | ↑↓/PgUp/PgDn: 滚动当前区域"))
	return b.String()
}

func (p PostsPageModel) buildDetailBodyContent(contentWidth int) string {
	if p.CurrentPost == nil {
		return ""
	}
	textWidth := p.detailBodyTextWidth(contentWidth)
	return vPostTextStyle.Render(strings.Join(
		p.wrapPlainTextLines(p.postDisplayText(*p.CurrentPost), textWidth),
		"\n",
	))
}

func (p PostsPageModel) buildPostListContent(contentWidth int) string {
	selStyle := lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		Padding(0, 0, 0, 1).
		Background(colorBg).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorAccent).
		Render

	var content strings.Builder
	lineNo := 0
	for i, post := range p.PostList {
		if i > 0 {
			content.WriteString(p.linePrefix(lineNo) + "\n")
			lineNo++
		}

		selected := i == p.SelectedPostIdx
		lineWidth := p.listLineTextWidth(contentWidth, selected)
		headerLines := p.postHeaderLines(post, lineWidth)
		textLines := p.wrapPlainTextLines(p.postDisplayText(post), lineWidth)

		if selected {
			for _, line := range headerLines {
				content.WriteString(selStyle(p.linePrefix(lineNo)+line) + "\n")
				lineNo++
			}
			for _, line := range textLines {
				content.WriteString(selStyle(p.linePrefix(lineNo)+line) + "\n")
				lineNo++
			}
		} else {
			for _, line := range headerLines {
				content.WriteString(p.linePrefix(lineNo) + line + "\n")
				lineNo++
			}
			for _, line := range textLines {
				content.WriteString(p.linePrefix(lineNo) + line + "\n")
				lineNo++
			}
		}
	}
	return content.String()
}

func (p PostsPageModel) buildCommentContent(contentWidth int) string {
	if len(p.CommentList) == 0 {
		if p.CommentListLoading {
			return vLoadingStyle.Render("加载评论中...")
		}
		if p.CommentListError != "" {
			return vErrorStyle.Render("错误: " + p.CommentListError)
		}
		return vEmptyStyle.Render("暂无评论")
	}
	textWidth := p.commentBodyTextWidth(contentWidth)
	var content strings.Builder
	comments := p.orderedComments()
	for i, c := range comments {
		if i > 0 {
			content.WriteString("\n")
		}

		cName := c.NameTag
		if cName == "" {
			cName = "匿名"
		}
		cTs := time.Unix(int64(c.Timestamp), 0).In(shanghaiLocation).Format("2006-01-02 15:04")
		content.WriteString(cTs)
		content.WriteString("\n")
		if quotePreview := p.commentQuotePreview(c, textWidth); quotePreview != "" {
			content.WriteString("  " + vCommentQuoteStyle.Width(textWidth).Render(quotePreview) + "\n")
		}
		commentLines := p.wrapPlainTextLines(p.commentDisplayText(c, cName), textWidth)
		for j, line := range commentLines {
			if j > 0 {
				content.WriteString("\n")
			}
			content.WriteString("  " + line)
		}
	}
	if p.CommentListError != "" {
		content.WriteString("\n\n")
		content.WriteString(vErrorStyle.Render("错误: " + p.CommentListError))
	} else if p.CommentListLoading {
		content.WriteString("\n\n")
		content.WriteString(vLoadingStyle.Render("加载更多评论中..."))
	}
	return content.String()
}

func (p PostsPageModel) orderedComments() []models.Comment {
	return p.CommentList
}

func (p PostsPageModel) commentQuotePreview(c models.Comment, width int) string {
	if c.Quote == nil {
		return ""
	}
	quoteName := c.Quote.NameTag
	if quoteName == "" {
		quoteName = "匿名"
	}
	preview := fmt.Sprintf("%s: %s", quoteName, strings.ReplaceAll(c.Quote.Text, "\n", " "))
	return truncateVisibleLine(preview, width, "...")
}

func (p *PostsPageModel) scrollToSelectedPost() {
	if len(p.PostList) == 0 {
		return
	}
	startLine, _ := p.selectedPostLineRange()
	p.CursorLine = startLine
	p.scrollCursorIntoView()
}

func (p *PostsPageModel) selectedPostLineRange() (startLine, endLine int) {
	line := 0
	for i := 0; i < len(p.PostList); i++ {
		if i > 0 {
			line++
		}
		postLines := p.postRenderedLinesAt(i)
		if i == p.SelectedPostIdx {
			return line, line + postLines - 1
		}
		line += postLines
	}
	return 0, 0
}

func (p *PostsPageModel) adjustSelectedToViewport() {
	if len(p.PostList) == 0 {
		return
	}
	yOffset := p.PostViewport.YOffset
	visibleLines := p.PostViewport.VisibleLineCount()
	lineIdx := 0
	for i := 0; i < len(p.PostList); i++ {
		if i > 0 {
			if lineIdx == yOffset {
				p.SelectedPostIdx = i - 1
				return
			}
			lineIdx++
		}
		postLines := p.postRenderedLinesAt(i)
		if lineIdx+postLines > yOffset && lineIdx < yOffset+visibleLines {
			p.SelectedPostIdx = i
			return
		}
		lineIdx += postLines
	}
}

func (p *PostsPageModel) moveCursor(delta int) {
	if len(p.PostList) == 0 {
		return
	}
	totalLines := p.totalPostLines()
	if totalLines <= 0 {
		return
	}
	p.CursorLine = clampInt(p.CursorLine+delta, 0, totalLines-1)
	p.SelectedPostIdx = p.postIndexAtLine(p.CursorLine)
	p.scrollCursorIntoView()
}

func (p *PostsPageModel) pageMove(direction int) {
	if len(p.PostList) == 0 || direction == 0 {
		return
	}
	totalLines := p.totalPostLines()
	if totalLines <= 0 {
		return
	}

	step := p.pageStep()
	delta := step
	if direction < 0 {
		delta = -step
	}

	p.CursorLine = clampInt(p.CursorLine+delta, 0, totalLines-1)
	p.SelectedPostIdx = p.postIndexAtLine(p.CursorLine)

	maxOffset := maxInt(0, totalLines-p.PostViewport.VisibleLineCount())
	p.PostViewport.SetYOffset(clampInt(p.PostViewport.YOffset+delta, 0, maxOffset))
	p.scrollCursorIntoView()
}

func (p *PostsPageModel) pageStep() int {
	visibleLines := p.PostViewport.VisibleLineCount()
	if visibleLines <= 1 {
		return 1
	}
	step := visibleLines - 2
	if step < 1 {
		step = 1
	}
	return step
}

func (p *PostsPageModel) scrollCursorIntoView() {
	visibleLines := p.PostViewport.VisibleLineCount()
	if visibleLines <= 0 {
		return
	}

	topMargin := 2
	bottomMargin := 15
	if maxTop := visibleLines / 4; maxTop < topMargin {
		topMargin = maxTop
	}
	if maxBottom := visibleLines - 2; maxBottom < bottomMargin {
		bottomMargin = maxBottom
	}
	if topMargin < 1 {
		topMargin = 1
	}
	if bottomMargin < 1 {
		bottomMargin = 1
	}

	topThreshold := p.PostViewport.YOffset + topMargin
	bottomThreshold := p.PostViewport.YOffset + visibleLines - bottomMargin - 1

	if p.CursorLine < topThreshold {
		newOffset := p.CursorLine - topMargin
		if newOffset < 0 {
			newOffset = 0
		}
		p.PostViewport.SetYOffset(newOffset)
		return
	}
	if p.CursorLine > bottomThreshold {
		newOffset := p.CursorLine - visibleLines + bottomMargin + 1
		if newOffset < 0 {
			newOffset = 0
		}
		p.PostViewport.SetYOffset(newOffset)
	}
}

func (p *PostsPageModel) syncCursorToSelection() {
	if len(p.PostList) == 0 {
		p.CursorLine = 0
		p.SelectedPostIdx = 0
		return
	}
	if p.SelectedPostIdx < 0 {
		p.SelectedPostIdx = 0
	}
	if p.SelectedPostIdx >= len(p.PostList) {
		p.SelectedPostIdx = len(p.PostList) - 1
	}
	startLine, endLine := p.selectedPostLineRange()
	separatorAfter := endLine
	if p.SelectedPostIdx < len(p.PostList)-1 {
		separatorAfter = endLine + 1
	}
	if p.CursorLine < startLine || p.CursorLine > separatorAfter {
		p.CursorLine = startLine
	}
}

func (p *PostsPageModel) postIndexAtLine(target int) int {
	line := 0
	for i := 0; i < len(p.PostList); i++ {
		if i > 0 {
			if target == line {
				return i - 1
			}
			line++
		}
		postLines := p.postRenderedLinesAt(i)
		if target < line+postLines {
			return i
		}
		line += postLines
	}
	return maxInt(0, len(p.PostList)-1)
}

func (p *PostsPageModel) totalPostLines() int {
	total := 0
	for i := 0; i < len(p.PostList); i++ {
		if i > 0 {
			total++
		}
		total += p.postRenderedLinesAt(i)
	}
	return total
}

func (p *PostsPageModel) postRenderedLinesAt(index int) int {
	if index < 0 || index >= len(p.PostList) {
		return 0
	}
	post := p.PostList[index]
	selected := index == p.SelectedPostIdx
	lineWidth := p.listLineTextWidth(p.currentListContentWidth(), selected)
	headerLines := len(p.postHeaderLines(post, lineWidth))
	textLines := len(p.wrapPlainTextLines(p.postDisplayText(post), lineWidth))
	return headerLines + textLines
}

func (p *PostsPageModel) atLastContentLine() bool {
	total := p.totalPostLines()
	return total > 0 && p.CursorLine >= total-1
}

func (p *PostsPageModel) shouldPrefetchMore() bool {
	if p.PostListLoading || !p.PostListHasMore || len(p.PostList) == 0 {
		return false
	}
	totalLines := p.totalPostLines()
	remainingLines := totalLines - p.CursorLine - 1
	return remainingLines <= 10
}

func (p PostsPageModel) linePrefix(lineNo int) string {
	if lineNo == p.CursorLine {
		if p.isSeparatorLine(lineNo) {
			return lipgloss.NewStyle().Foreground(colorMuted).Render("· ")
		}
		return "▸ "
	}
	return "  "
}

func (p PostsPageModel) isSeparatorLine(target int) bool {
	line := 0
	for i := 0; i < len(p.PostList); i++ {
		if i > 0 {
			if line == target {
				return true
			}
			line++
		}
		line += p.postRenderedLinesAt(i)
	}
	return false
}

func (p *PostsPageModel) resetList() {
	p.PostList = nil
	p.PostListTotal = 0
	p.PostListHasMore = false
	p.PostListCursor = 0
	p.CursorLine = 0
	p.SelectedPostIdx = 0
	p.postContent = ""
	p.PostViewport.GotoTop()
	p.PostsMode = PostsModeList
}

func (p *PostsPageModel) resetComments() {
	p.CommentList = nil
	p.CommentListHasMore = false
	p.CommentListLoading = false
	p.CommentListCursor = 0
	p.CommentListError = ""
	p.CommentSortAsc = true
	p.commentContent = ""
	p.CommentViewport.GotoTop()
}

func (p *PostsPageModel) calcPostViewportHeight(height int) int {
	titleLines := 1
	searchLines := 3
	paginationLines := 2
	avail := height - titleLines - searchLines - paginationLines
	if avail < 3 {
		avail = 3
	}
	return avail
}

func (p *PostsPageModel) calcDetailViewportHeights(height int) (int, int) {
	available := p.calcPostViewportHeight(height) - 3
	if available < 8 {
		return 4, 3
	}

	minBodyHeight := 2
	minCommentHeight := 3
	maxBodyHeight := available / 2
	if maxBodyHeight < minBodyHeight {
		maxBodyHeight = minBodyHeight
	}
	if maxAllowed := available - minCommentHeight; maxBodyHeight > maxAllowed {
		maxBodyHeight = maxAllowed
	}

	bodyLines := p.detailBodyLineCount()
	commentLines := p.commentLineCount()

	bodyHeight := minInt(maxInt(bodyLines, minBodyHeight), maxBodyHeight)
	commentHeight := available - bodyHeight

	if commentHeight < minCommentHeight {
		commentHeight = minCommentHeight
		bodyHeight = available - commentHeight
	}

	extra := available - bodyHeight - commentHeight
	if extra > 0 {
		bodyNeed := maxInt(0, bodyLines-bodyHeight)
		commentNeed := maxInt(0, commentLines-commentHeight)
		switch {
		case commentNeed > bodyNeed:
			add := minInt(extra, commentNeed)
			commentHeight += add
			extra -= add
		case bodyNeed > 0:
			add := minInt(extra, minInt(bodyNeed, maxBodyHeight-bodyHeight))
			bodyHeight += add
			extra -= add
		}
		if extra > 0 {
			commentHeight += extra
		}
	}

	return bodyHeight, commentHeight
}

func clampInt(value, minValue, maxValue int) int {
	if maxValue < minValue {
		return minValue
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func (p *PostsPageModel) detailBodyLineCount() int {
	if p.CurrentPost == nil || p.CurrentPost.Text == "" {
		if p.CurrentPost == nil {
			return 1
		}
	}
	width := 20
	if p.PostBodyViewport != nil && p.PostBodyViewport.Width > 0 {
		width = p.PostBodyViewport.Width
	}
	return len(p.wrapPlainTextLines(p.postDisplayText(*p.CurrentPost), p.detailBodyTextWidth(width)))
}

func (p *PostsPageModel) commentLineCount() int {
	width := 20
	if p.CommentViewport != nil && p.CommentViewport.Width > 0 {
		width = p.CommentViewport.Width
	}
	if len(p.CommentList) == 0 {
		if p.CommentListLoading || p.CommentListError != "" {
			return len(strings.Split(p.buildCommentContent(width), "\n"))
		}
		return 1
	}
	textWidth := p.commentBodyTextWidth(width)
	lines := 0
	for i, c := range p.orderedComments() {
		if i > 0 {
			lines++
		}
		lines++
		if p.commentQuotePreview(c, textWidth) != "" {
			lines++
		}
		cName := c.NameTag
		if cName == "" {
			cName = "匿名"
		}
		lines += len(p.wrapPlainTextLines(p.commentDisplayText(c, cName), textWidth))
	}
	if p.CommentListError != "" || p.CommentListLoading {
		lines += 2
	}
	return lines
}

func (p *PostsPageModel) shouldPrefetchCommentsMore() bool {
	if p.CommentListLoading || !p.CommentListHasMore || p.CommentViewport == nil {
		return false
	}
	totalLines := p.commentLineCount()
	bottom := p.CommentViewport.YOffset + p.CommentViewport.Height
	return totalLines-bottom <= 3
}

func (p PostsPageModel) currentListContentWidth() int {
	if p.PostViewport != nil && p.PostViewport.Width > 0 {
		return p.PostViewport.Width
	}
	return 20
}

func (p PostsPageModel) postDisplayText(post models.Post) string {
	text := post.Text
	if p.hasPostMedia(post) {
		if text == "" {
			return "[图片]"
		}
		return text + "\n[图片]"
	}
	return text
}

func (p PostsPageModel) commentDisplayText(c models.Comment, name string) string {
	text := fmt.Sprintf("%s: %s", name, c.Text)
	if p.hasCommentMedia(c) {
		return text + "\n[图片]"
	}
	return text
}

func (p PostsPageModel) hasPostMedia(post models.Post) bool {
	return post.Type == "image" || strings.TrimSpace(post.MediaIds) != ""
}

func (p PostsPageModel) hasCommentMedia(c models.Comment) bool {
	return strings.TrimSpace(c.MediaIds) != ""
}

func (p PostsPageModel) listLineTextWidth(contentWidth int, selected bool) int {
	width := contentWidth - lipgloss.Width("  ")
	if selected {
		width -= 2
	}
	if width < 1 {
		width = 1
	}
	return width
}

func (p PostsPageModel) detailBodyTextWidth(contentWidth int) int {
	width := contentWidth - vDetailSection.GetHorizontalFrameSize() - vPostTextStyle.GetHorizontalFrameSize()
	if focusedWidth := contentWidth - vDetailSectionFocused.GetHorizontalFrameSize() - vPostTextStyle.GetHorizontalFrameSize(); focusedWidth < width {
		width = focusedWidth
	}
	return maxInt(1, width)
}

func (p PostsPageModel) commentBodyTextWidth(contentWidth int) int {
	width := contentWidth - vDetailSection.GetHorizontalFrameSize() - 2
	if focusedWidth := contentWidth - vDetailSectionFocused.GetHorizontalFrameSize() - 2; focusedWidth < width {
		width = focusedWidth
	}
	return maxInt(1, width)
}

func (p PostsPageModel) postHeader(post models.Post) string {
	ts := time.Unix(int64(post.Timestamp), 0).In(shanghaiLocation).Format("2006-01-02 15:04")
	replyStr := vPostReplyStyle.Render(fmt.Sprintf("回复:%d", post.Reply))
	likeStr := vPostLikeStyle.Render(fmt.Sprintf("赞:%d", post.Likenum))
	meta := replyStr + " " + likeStr
	pidStr := vPostPidStyle.Render(fmt.Sprintf("#%-6d", post.Pid))
	tsStr := vPostTimeStyle.Render(ts)
	if !post.Anonymous {
		return pidStr + " [实名] " + tsStr + "  " + meta
	}
	return pidStr + " " + tsStr + "  " + meta
}

func (p PostsPageModel) postHeaderPlain(post models.Post) string {
	ts := time.Unix(int64(post.Timestamp), 0).In(shanghaiLocation).Format("2006-01-02 15:04")
	header := fmt.Sprintf("#%-6d %s  回复:%d 赞:%d", post.Pid, ts, post.Reply, post.Likenum)
	if !post.Anonymous {
		header = fmt.Sprintf("#%-6d [实名] %s  回复:%d 赞:%d", post.Pid, ts, post.Reply, post.Likenum)
	}
	return header
}

func (p PostsPageModel) postHeaderLines(post models.Post, width int) []string {
	styled := p.postHeader(post)
	if lipgloss.Width(styled) <= width {
		return []string{styled}
	}
	return wrapVisibleLine(p.postHeaderPlain(post), width)
}

func (p PostsPageModel) wrapPlainTextLines(text string, width int) []string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	rawLines := strings.Split(normalized, "\n")
	if len(rawLines) == 0 {
		return []string{""}
	}

	var wrapped []string
	for _, line := range rawLines {
		wrapped = append(wrapped, wrapVisibleLine(line, width)...)
	}
	if len(wrapped) == 0 {
		return []string{""}
	}
	return wrapped
}

func wrapVisibleLine(line string, width int) []string {
	if width < 1 {
		width = 1
	}
	if line == "" {
		return []string{""}
	}

	var wrapped []string
	var current []rune
	currentWidth := 0

	for _, r := range []rune(line) {
		runeWidth := lipgloss.Width(string(r))
		if runeWidth < 1 {
			runeWidth = 1
		}
		if len(current) > 0 && currentWidth+runeWidth > width {
			wrapped = append(wrapped, string(current))
			current = current[:0]
			currentWidth = 0
		}
		current = append(current, r)
		currentWidth += runeWidth
	}

	if len(current) > 0 {
		wrapped = append(wrapped, string(current))
	}
	if len(wrapped) == 0 {
		return []string{""}
	}
	return wrapped
}

func truncateVisibleLine(line string, width int, suffix string) string {
	if width < 1 {
		return ""
	}
	if suffix == "" {
		suffix = "..."
	}
	if lipgloss.Width(line) <= width {
		return line
	}

	suffixWidth := lipgloss.Width(suffix)
	if suffixWidth >= width {
		return string([]rune(suffix)[:1])
	}

	var current []rune
	currentWidth := 0
	for _, r := range []rune(line) {
		runeWidth := lipgloss.Width(string(r))
		if runeWidth < 1 {
			runeWidth = 1
		}
		if currentWidth+runeWidth+suffixWidth > width {
			break
		}
		current = append(current, r)
		currentWidth += runeWidth
	}
	return string(current) + suffix
}
