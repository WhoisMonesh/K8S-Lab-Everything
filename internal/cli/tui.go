package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/labs"
	"github.com/WhoisMonesh/K8S-Lab-Everything/internal/progress"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(2)

	normalItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("249"))

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(lipgloss.Color("229")).
				Bold(true)

	selectedItemBarStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("57")).
				Foreground(lipgloss.Color("229")).
				Bold(true).
				PaddingLeft(1).
				PaddingRight(1)

	diffEasyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	diffMediumStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")).
				Bold(true)

	diffHardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	catStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	checkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

type labItem struct {
	lab      labs.Lab
	info     labs.Info
	completed bool
}

type labSelectorModel struct {
	labs        []labItem
	cursor      int
	filtered    []labItem
	filter      string
	selected    int
	quitting    bool
	width       int
	height      int
}

func newLabSelectorModel(labList []labs.Lab, showProgress bool) labSelectorModel {
	items := make([]labItem, len(labList))
	for i, l := range labList {
		info := labs.GetInfo(l)
		items[i] = labItem{
			lab:       l,
			info:      info,
			completed: showProgress && progress.IsCompleted(info.ID),
		}
	}

	return labSelectorModel{
		labs:     items,
		filtered: items,
		cursor:   0,
	}
}

func (m labSelectorModel) Init() tea.Cmd {
	return nil
}

func (m labSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case "pgup":
			m.cursor -= 10
			if m.cursor < 0 {
				m.cursor = 0
			}

		case "pgdown":
			m.cursor += 10
			if m.cursor >= len(m.filtered) {
				m.cursor = len(m.filtered) - 1
			}

		case "home", "g":
			m.cursor = 0

		case "end", "G":
			m.cursor = len(m.filtered) - 1

		case "enter":
			m.selected = m.cursor
			return m, tea.Quit

		case "/":
			m.filter = ""
			return m, nil

		case "backspace":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.applyFilter()
			}

		case "escape":
			if m.filter != "" {
				m.filter = ""
				m.applyFilter()
			}
		}

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			// Calculate which item was clicked based on row
			clickRow := msg.Y - 8 // Offset for header
			if clickRow >= 0 && clickRow < len(m.filtered) {
				m.cursor = clickRow
				m.selected = m.cursor
				return m, tea.Quit
			}
		}
		if msg.Type == tea.MouseWheelUp {
			if m.cursor > 0 {
				m.cursor--
			}
		}
		if msg.Type == tea.MouseWheelDown {
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

func (m *labSelectorModel) applyFilter() {
	if m.filter == "" {
		m.filtered = m.labs
	} else {
		var result []labItem
		for _, item := range m.labs {
			if strings.Contains(strings.ToLower(item.info.ID), strings.ToLower(m.filter)) ||
				strings.Contains(strings.ToLower(item.info.Title), strings.ToLower(m.filter)) ||
				strings.Contains(strings.ToLower(string(item.info.Category)), strings.ToLower(m.filter)) {
				result = append(result, item)
			}
		}
		m.filtered = result
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func diffStyle(d string) lipgloss.Style {
	switch strings.ToLower(d) {
	case "easy":
		return diffEasyStyle
	case "medium":
		return diffMediumStyle
	case "hard":
		return diffHardStyle
	default:
		return normalItemStyle
	}
}

func (m labSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  K8S-Lab-Everything"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("  Select a lab to practice with"))
	b.WriteString("\n")

	// Filter input
	if m.filter != "" {
		b.WriteString(fmt.Sprintf("  Filter: %s\n\n", m.filter))
	} else {
		b.WriteString("\n")
	}

	// Lab list
	visibleCount := 0
	for i, item := range m.filtered {
		isSelected := i == m.cursor

		diffTag := diffStyle(string(item.info.Difficulty)).Render(fmt.Sprintf(" %s ", strings.ToUpper(string(item.info.Difficulty))))

		check := "  "
		if item.completed {
			check = checkStyle.Render("✔ ")
		}

		title := item.info.Title
		if len(title) > 38 {
			title = title[:35] + "..."
		}

		id := item.info.ID
		if len(id) > 28 {
			id = id[:25] + "..."
		}

		if isSelected {
			row := selectedItemBarStyle.Render(fmt.Sprintf(" ▸ %-28s  %-38s  %s", id, title, diffTag))
			b.WriteString(fmt.Sprintf("  %s%s\n", check, row))
		} else {
			row := normalItemStyle.Render(fmt.Sprintf("  %-28s  %-38s  %s", id, title, diffTag))
			b.WriteString(fmt.Sprintf("  %s%s\n", check, row))
		}
		visibleCount++
		_ = visibleCount
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render(strings.Repeat("─", 70)))
	b.WriteString("\n")

	// Stats
	easy, medium, hard := 0, 0, 0
	for _, item := range m.filtered {
		switch strings.ToLower(string(item.info.Difficulty)) {
		case "easy":
			easy++
		case "medium":
			medium++
		case "hard":
			hard++
		}
	}
	b.WriteString(fmt.Sprintf("  %s%d lab(s)%s  %s  %s  %s",
		lipgloss.NewStyle().Bold(true).Render(""),
		len(m.filtered),
		"",
		diffEasyStyle.Render(fmt.Sprintf("● %d easy", easy)),
		diffMediumStyle.Render(fmt.Sprintf("● %d medium", medium)),
		diffHardStyle.Render(fmt.Sprintf("● %d hard", hard)),
	))

	b.WriteString("\n")

	// Help
	b.WriteString(helpStyle.Render("  ↑↓ navigate  ↵ select  / filter  q quit"))

	b.WriteString("\n")

	return b.String()
}

func RunInteractiveLabSelector(labList []labs.Lab) (labs.Lab, error) {
	model := newLabSelectorModel(labList, true)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("interactive selector failed: %w", err)
	}

	result := finalModel.(labSelectorModel)
	if result.quitting || result.selected >= len(result.filtered) {
		return nil, fmt.Errorf("no lab selected")
	}

	return result.filtered[result.selected].lab, nil
}
