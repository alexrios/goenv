package tui

import tea "charm.land/bubbletea/v2"

// HelpScreenModel displays keyboard shortcuts and usage information.
type HelpScreenModel struct{}

// NewHelpScreenModel creates a new help screen model.
func NewHelpScreenModel() HelpScreenModel {
	return HelpScreenModel{}
}

// Init initializes the help screen. Returns nil as no initial command is needed.
func (m HelpScreenModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help screen.
func (m *HelpScreenModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q", "?":
			return sendMsg(changeScreenMsg(AppScreenList))
		}
	}
	return nil
}

// View renders the help screen with all keyboard shortcuts organized by category.
func (m HelpScreenModel) View() string {
	var content string
	content += helpTitleStyle.Render(helpScreenTitle) + "\n\n"

	content += helpSectionStyle.Render("Navigation") + "\n"
	content += helpKeyStyle.Render("j/k") + helpDescStyle.Render("Move down/up") + "\n"
	content += helpKeyStyle.Render("Ctrl+D") + helpDescStyle.Render("Half-page down") + "\n"
	content += helpKeyStyle.Render("Ctrl+U") + helpDescStyle.Render("Half-page up") + "\n"
	content += helpKeyStyle.Render("g/G") + helpDescStyle.Render("Go to top/bottom") + "\n"
	content += helpKeyStyle.Render("/") + helpDescStyle.Render("Search/filter variables") + "\n"
	content += helpKeyStyle.Render("Enter") + helpDescStyle.Render("Edit variable (view details if [RO])") + "\n"
	content += helpKeyStyle.Render("q") + helpDescStyle.Render("Quit") + "\n"

	content += "\n" + helpSectionStyle.Render("Actions") + "\n"
	content += helpKeyStyle.Render("y") + helpDescStyle.Render("Copy value to clipboard") + "\n"
	content += helpKeyStyle.Render("Y") + helpDescStyle.Render("Copy KEY=VALUE to clipboard") + "\n"
	content += helpKeyStyle.Render("f") + helpDescStyle.Render("Toggle favorite (pin to top)") + "\n"
	content += helpKeyStyle.Render("s") + helpDescStyle.Render("Toggle sort mode") + "\n"
	content += helpKeyStyle.Render("C") + helpDescStyle.Render("Cycle category filter") + "\n"
	content += helpKeyStyle.Render("r") + helpDescStyle.Render("Reload environment") + "\n"
	content += helpKeyStyle.Render("t") + helpDescStyle.Render("Cycle color theme") + "\n"
	content += helpKeyStyle.Render("w") + helpDescStyle.Render("Toggle watch mode (auto-reload)") + "\n"
	content += helpKeyStyle.Render(":") + helpDescStyle.Render("Open command palette") + "\n"
	content += helpKeyStyle.Render("Ctrl+Z") + helpDescStyle.Render("Undo last change") + "\n"
	content += helpKeyStyle.Render("Ctrl+Y") + helpDescStyle.Render("Redo last change") + "\n"

	content += "\n" + helpSectionStyle.Render("Snapshots & Presets") + "\n"
	content += helpKeyStyle.Render("e") + helpDescStyle.Render("Export snapshot") + "\n"
	content += helpKeyStyle.Render("i") + helpDescStyle.Render("Import snapshot") + "\n"
	content += helpKeyStyle.Render("p") + helpDescStyle.Render("Manage presets") + "\n"
	content += helpKeyStyle.Render("c") + helpDescStyle.Render("Compare with snapshot/preset") + "\n"
	content += helpKeyStyle.Render("S") + helpDescStyle.Render("Show statistics") + "\n"

	content += "\n" + helpSectionStyle.Render("Reset & Export") + "\n"
	content += helpKeyStyle.Render("u") + helpDescStyle.Render("Reset variable to default") + "\n"
	content += helpKeyStyle.Render("U") + helpDescStyle.Render("Batch reset to defaults") + "\n"
	content += helpKeyStyle.Render("x") + helpDescStyle.Render("Export as shell commands") + "\n"

	content += "\n" + helpSectionStyle.Render("Edit Mode") + "\n"
	content += helpKeyStyle.Render("Enter") + helpDescStyle.Render("Save changes") + "\n"
	content += helpKeyStyle.Render("Esc") + helpDescStyle.Render("Cancel editing") + "\n"
	content += helpKeyStyle.Render("Ctrl+R") + helpDescStyle.Render("Reset to original value") + "\n"
	content += helpKeyStyle.Render("Tab") + helpDescStyle.Render("Next suggestion") + "\n"
	content += helpKeyStyle.Render("Shift+Tab") + helpDescStyle.Render("Previous suggestion") + "\n"

	content += "\n" + helpSectionStyle.Render("Read-Only Variables") + "\n"
	content += helpDescStyle.Render("Variables marked [RO] are computed by Go and cannot") + "\n"
	content += helpDescStyle.Render("be changed. Press Enter to view details, y/Y to copy.") + "\n"

	content += "\n" + helpFooterStyle.Render(helpScreenFooter)

	return helpBoxStyle.Render(content)
}
