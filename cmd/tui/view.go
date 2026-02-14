package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 {
		return m.spinner.View() + " Initialising…"
	}

	// Header (full width)
	header := m.renderHeader()

	// Sidebar (left pane)
	sidebar := m.renderSidebar()

	// Chat pane (right pane)
	chatPane := m.renderChatPane()

	// Join sidebar and chat horizontally
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, chatPane)

	// Status bar (full width)
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, header, mainArea, statusBar)
}

// mainHeight returns the height available for the sidebar/chat area
func (m Model) mainHeight() int {
	headerH := 1
	statusH := 1
	if m.showHelp {
		statusH = 4
	}
	h := m.height - headerH - statusH
	if h < 4 {
		h = 4
	}
	return h
}

func (m Model) renderHeader() string {
	title := "Webex Teams Chat"
	if m.currentRoomTitle != "" {
		title = fmt.Sprintf("Webex Teams Chat — %s", m.currentRoomTitle)
	}

	statusIcon := disconnectedStyle.Render("●")
	if m.listening {
		statusIcon = connectedStyle.Render("●")
	}

	left := headerStyle.Render(title)
	right := headerStyle.Render(statusIcon)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m Model) renderSidebar() string {
	h := m.mainHeight()
	sw := sidebarWidth

	// Search box
	searchStyle := sidebarSearchStyle.Width(sw - 4)
	search := searchStyle.Render(m.sidebarInput.View())

	// Type filter tabs
	filterLine := m.renderFilterTabs()

	// Room list
	listHeight := h - 4 // search + filter + borders
	if listHeight < 1 {
		listHeight = 1
	}
	roomListStr := m.renderRoomList(listHeight)

	content := lipgloss.JoinVertical(lipgloss.Left, search, filterLine, roomListStr)

	return sidebarStyle.
		Width(sw).
		Height(h).
		Render(content)
}

func (m Model) renderFilterTabs() string {
	filters := []struct {
		key   string
		label string
		value string
	}{
		{"a", "All", "all"},
		{"g", "Group", "group"},
		{"d", "Direct", "direct"},
	}

	var parts []string
	for _, f := range filters {
		label := fmt.Sprintf("[%s]%s", f.key, f.label)
		if m.roomTypeFilter == f.value {
			parts = append(parts, sidebarFilterActiveStyle.Render(label))
		} else {
			parts = append(parts, sidebarFilterStyle.Render(label))
		}
	}
	return strings.Join(parts, " ")
}

func (m Model) renderRoomList(maxVisible int) string {
	if len(m.filteredRooms) == 0 {
		if len(m.allRooms) == 0 {
			return m.spinner.View() + " Loading…"
		}
		return lipgloss.NewStyle().Foreground(mutedColor).Render("  No matching rooms")
	}

	visibleStart := 0
	if m.roomCursor >= maxVisible {
		visibleStart = m.roomCursor - maxVisible + 1
	}

	var b strings.Builder
	for i := visibleStart; i < len(m.filteredRooms) && i < visibleStart+maxVisible; i++ {
		room := m.filteredRooms[i]

		// Type badge
		badge := "[G]"
		if room.Type == "direct" {
			badge = "[D]"
		}

		// Truncate title
		maxTitleLen := sidebarWidth - 8
		title := room.Title
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen-1] + "…"
		}

		label := fmt.Sprintf("%s %s", badge, title)

		if i == m.roomCursor {
			// Highlight if this is the cursor position
			if room.ID == m.currentRoomID {
				b.WriteString(sidebarItemActiveStyle.Render("▸ " + label))
			} else {
				b.WriteString(roomItemSelectedStyle.Render("▸ " + label))
			}
		} else if room.ID == m.currentRoomID {
			b.WriteString(sidebarItemActiveStyle.Render("  " + label))
		} else {
			b.WriteString(roomItemStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderChatPane() string {
	h := m.mainHeight()
	chatW := m.width - sidebarWidth - 1
	if chatW < 10 {
		chatW = 10
	}

	if m.currentRoomID == "" {
		placeholder := lipgloss.NewStyle().
			Foreground(mutedColor).
			Width(chatW).
			Height(h).
			Align(lipgloss.Center, lipgloss.Center).
			Render("Select a room to start chatting")
		return placeholder
	}

	// Room title bar
	titleBar := chatHeaderStyle.Width(chatW).Render(m.currentRoomTitle)

	// Reply bar (if replying)
	replyBar := ""
	if m.replyToMsg != nil {
		truncText := m.replyToMsg.text
		if len(truncText) > 40 {
			truncText = truncText[:37] + "..."
		}
		replyBar = replyBarStyle.Width(chatW).Render(
			fmt.Sprintf("Replying to %s: %s  [Esc to cancel]", m.replyToMsg.senderEmail, truncText),
		)
	}

	// Chat viewport
	extraH := 4 // title + input
	if replyBar != "" {
		extraH += 1
	}
	vpHeight := h - extraH
	if vpHeight < 1 {
		vpHeight = 1
	}
	m.viewport.Width = chatW
	m.viewport.Height = vpHeight
	chatView := m.viewport.View()

	// Input
	inputStyle := inputBorderStyle.Width(chatW - 4)
	input := inputStyle.Render(m.textInput.View())

	// Image viewer overlay
	if m.focus == focusImageViewer && m.showingImage {
		imgView := m.renderImageViewer(chatW, vpHeight)
		return lipgloss.JoinVertical(lipgloss.Left, titleBar, imgView, input)
	}

	// File picker overlay
	if m.focus == focusFilePicker {
		picker := m.renderFilePicker(chatW, vpHeight)
		if replyBar != "" {
			return lipgloss.JoinVertical(lipgloss.Left, titleBar, replyBar, picker, input)
		}
		return lipgloss.JoinVertical(lipgloss.Left, titleBar, picker, input)
	}

	if replyBar != "" {
		return lipgloss.JoinVertical(lipgloss.Left, titleBar, replyBar, chatView, input)
	}
	return lipgloss.JoinVertical(lipgloss.Left, titleBar, chatView, input)
}

func (m Model) renderImageViewer(width, height int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(" 🖼 %s\n", m.imageFileName))
	b.WriteString(" [Esc] Close  [s] Save to disk\n\n")
	b.WriteString(m.imageText)

	return filePickerStyle.Width(width - 4).Height(height).Render(b.String())
}

func (m Model) renderFilePicker(width, height int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(" 📁 %s\n", m.filePickerDir))
	if m.filePickerFn == filePickerModeSave {
		b.WriteString(fmt.Sprintf(" Save: %s\n", m.pendingFileName))
		b.WriteString(" [Enter] Open dir  [s] Save here  [Backspace] Up  [Esc] Cancel\n\n")
	} else {
		b.WriteString(" [Backspace] Go up  [Enter] Select  [Esc] Cancel\n\n")
	}

	maxVisible := height - 4
	if maxVisible < 3 {
		maxVisible = 3
	}

	if len(m.filePickerEntries) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("  (empty directory)"))
	} else {
		visibleStart := 0
		if m.filePickerCursor >= maxVisible {
			visibleStart = m.filePickerCursor - maxVisible + 1
		}

		for i := visibleStart; i < len(m.filePickerEntries) && i < visibleStart+maxVisible; i++ {
			entry := m.filePickerEntries[i]
			name := entry.Name()
			if entry.IsDir() {
				name = name + "/"
			}

			if i == m.filePickerCursor {
				if entry.IsDir() {
					b.WriteString(filePickerSelectedStyle.Render("▸ " + filePickerDirStyle.Render(name)))
				} else {
					b.WriteString(filePickerSelectedStyle.Render("▸ " + name))
				}
			} else {
				if entry.IsDir() {
					b.WriteString("  " + filePickerDirStyle.Render(name))
				} else {
					b.WriteString("  " + name)
				}
			}
			b.WriteString("\n")
		}
	}

	return filePickerStyle.Width(width - 4).Height(height).Render(b.String())
}

func (m Model) renderStatusBar() string {
	if m.errMsg != "" {
		return errorStyle.Width(m.width).Render("Error: " + m.errMsg)
	}

	if m.showHelp {
		return m.renderFullHelp()
	}

	return m.renderShortHelp()
}

func (m Model) renderShortHelp() string {
	var parts []string
	for _, kb := range m.keys.ShortHelp() {
		k := helpKeyStyle.Render(kb.Help().Key)
		d := helpDescStyle.Render(kb.Help().Desc)
		parts = append(parts, fmt.Sprintf("%s %s", k, d))
	}
	return helpBarStyle.Width(m.width).Render(strings.Join(parts, "  "))
}

func (m Model) renderFullHelp() string {
	var lines []string
	for _, row := range m.keys.FullHelp() {
		var parts []string
		for _, kb := range row {
			k := helpKeyStyle.Render(kb.Help().Key)
			d := helpDescStyle.Render(kb.Help().Desc)
			parts = append(parts, fmt.Sprintf("%s %s", k, d))
		}
		lines = append(lines, strings.Join(parts, "  "))
	}
	return helpBarStyle.Width(m.width).Render(strings.Join(lines, "\n"))
}

// renderChatContent builds the full chat viewport content from chatEntries,
// grouping messages into threads and highlighting the chat cursor.
func (m Model) renderChatContent() string {
	if len(m.chatEntries) == 0 {
		return lipgloss.NewStyle().Foreground(mutedColor).Render("\n  No messages yet. Start typing!\n")
	}

	chatW := m.width - sidebarWidth - 10
	if chatW < 20 {
		chatW = 20
	}

	// Build a set of thread root IDs (messages that have children)
	threadRoots := make(map[string]bool)
	for _, entry := range m.chatEntries {
		if entry.parentID != "" {
			threadRoots[entry.parentID] = true
		}
	}

	// Track which thread replies we've already rendered (grouped under parent)
	rendered := make(map[int]bool)

	var lines []string
	for i, entry := range m.chatEntries {
		if rendered[i] {
			continue
		}

		isCursorHere := (m.focus == focusChat && i == m.chatCursor)

		// Render this message
		msgBlock := m.renderSingleMessage(entry, chatW, isCursorHere)

		// If this message is a thread root, collect and render its replies indented below
		if threadRoots[entry.messageID] {
			var threadBlock []string
			threadBlock = append(threadBlock, msgBlock)

			for j := i + 1; j < len(m.chatEntries); j++ {
				if m.chatEntries[j].parentID == entry.messageID {
					rendered[j] = true
					isReplyCursor := (m.focus == focusChat && j == m.chatCursor)
					replyBlock := m.renderSingleMessage(m.chatEntries[j], chatW-4, isReplyCursor)
					threadBlock = append(threadBlock, threadLineStyle.Render(replyBlock))
				}
			}
			lines = append(lines, strings.Join(threadBlock, "\n"))
		} else if entry.parentID != "" {
			// This is a reply whose parent was already rendered above (or not found).
			// If not already rendered as part of a thread group, render inline with indent.
			replyBlock := m.renderSingleMessage(entry, chatW-4, isCursorHere)
			lines = append(lines, threadLineStyle.Render(replyBlock))
		} else {
			lines = append(lines, msgBlock)
		}
	}

	return strings.Join(lines, "\n\n")
}

// renderSingleMessage renders one chat entry as a styled bubble
func (m Model) renderSingleMessage(entry chatEntry, maxW int, highlighted bool) string {
	var nameStyle, bubbleStyle lipgloss.Style
	if entry.isSelf {
		nameStyle = selfNameStyle
		bubbleStyle = selfBubbleStyle.Width(maxW)
	} else {
		nameStyle = otherNameStyle
		bubbleStyle = otherBubbleStyle.Width(maxW)
	}

	// Sender + optional timestamp
	senderLine := nameStyle.Render(entry.senderEmail)
	if entry.created != nil {
		ts := timestampStyle.Render(entry.created.Format("15:04"))
		senderLine = senderLine + " " + ts
	}

	// Message text
	text := wordwrap.String(entry.text, maxW-4)
	content := text

	// File attachments with names
	if entry.hasFiles {
		var fileParts []string
		for idx, fileURL := range entry.fileURLs {
			name := "attachment"
			if idx < len(entry.fileNames) {
				name = entry.fileNames[idx]
			}
			_ = fileURL
			fileParts = append(fileParts, fileBadgeStyle.Render(fmt.Sprintf("📎 [%d] %s", idx+1, name)))
		}
		fileBlock := strings.Join(fileParts, "  ")
		if content != "" {
			content = fileBlock + "\n" + content
		} else {
			content = fileBlock
		}
	}

	// Thread reply indicator for inline context
	if entry.parentID != "" {
		content = lipgloss.NewStyle().Foreground(mutedColor).Render("↳ thread reply") + "\n" + content
	}

	bubble := bubbleStyle.Render(content)
	result := senderLine + "\n" + bubble

	// Highlight if this is the chat cursor position
	if highlighted {
		result = chatCursorStyle.Render(result)
	}

	return result
}
