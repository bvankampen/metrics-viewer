package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createHeader(version string) *tview.TextView {
	header := tview.NewTextView()
	header.SetBackgroundColor(tcell.ColorDarkCyan)
	header.SetTextColor(tcell.ColorYellow)
	header.SetDynamicColors(true)
	header.SetText(fmt.Sprintf("metrics-viewer %s [white]https://github.com/bvankampen/metrics-viewer", version))
	return header
}

func createFooter() *tview.TextView {
	footer := tview.NewTextView()
	footer.SetDynamicColors(true)
	footer.SetBackgroundColor(tcell.ColorDarkCyan)
	footerText := "[yellow]q:[white] Quit " +
		"[yellow]/:[white] Filter "
	footer.SetText(footerText)
	return footer
}

func (ui *UI) updateFilterFlex() {
	text := tview.NewTextView()
	filter := "none"
	text.SetDynamicColors(true)
	text.SetBackgroundColor(tcell.ColorDarkCyan)
	if ui.filterText != "" {
		filter = ui.filterText
	}
	text.SetText(fmt.Sprintf("[yellow]Filter: [lightblue]%s", filter))
	ui.filterFlex.Clear()
	ui.filterFlex.AddItem(text, 0, 1, false)
}

func (ui *UI) updateLastUpdate() {
	ui.lastUpdateFlex.Clear()
	ui.lastUpdateFlex.SetBackgroundColor(tcell.ColorDarkCyan)
	current_time := time.Now()

	updateText := fmt.Sprintf(current_time.Format("2006-01-02 15:04:05"))

	text := tview.NewTextView()
	text.SetDynamicColors(true)
	text.SetBackgroundColor(tcell.ColorDarkCyan)

	text.SetText(fmt.Sprintf("[yellow]Last Update: [lightblue] %s", updateText))

	ui.lastUpdateFlex.AddItem(text, 0, 1, false)
}

func (ui *UI) appPage() *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	headerflex := tview.NewFlex()
	headerflex.SetBackgroundColor(tcell.ColorDarkCyan)
	bottomflex := tview.NewFlex()
	bottomflex.SetBackgroundColor(tcell.ColorDarkCyan)

	ui.filterFlex = tview.NewFlex().SetDirection(tview.FlexColumn)
	ui.lastUpdateFlex = tview.NewFlex()

	flex.AddItem(headerflex, 1, 1, false)
	flex.AddItem(ui.table, 0, 1, true)
	flex.AddItem(bottomflex, 1, 1, false)

	headerflex.AddItem(createHeader(ui.ctx.App.Version), 0, 3, false)
	headerflex.AddItem(ui.lastUpdateFlex, 0, 1, false)
	bottomflex.AddItem(createFooter(), 0, 2, false)
	bottomflex.AddItem(ui.filterFlex, 0, 1, false)

	ui.filterText = ""
	ui.updateFilterFlex()
	ui.updateLastUpdate()

	return flex
}
