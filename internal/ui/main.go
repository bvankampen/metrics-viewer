package ui

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

func NewAppUI(ctx *cli.Context) *UI {
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(false).
		SetFixed(0, 0)

	pages := tview.NewPages()

	return &UI{
		app:        app,
		table:      table,
		pages:      pages,
		sortAsc:    true,
		ctx:        ctx,
		sortColumn: 0,
	}
}

func (ui *UI) Run(observeChan <-chan interface{}) {
	go func() {
		for data := range observeChan {
			ui.updateTable(data)
			ui.app.Draw()
		}
	}()

	ui.app.SetInputCapture(ui.handleKeyEvents)
	ui.pages.AddPage("main", ui.appPage(), true, true)

	if err := ui.app.SetRoot(ui.pages, true).Run(); err != nil {
		panic(err)
	}
}

func (ui *UI) updateTable(data interface{}) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		log.Println("Invalid data format: expected map[string]interface{}.")
		return
	}

	uiData, ok := dataMap["uiData"].([]TableRow)
	if !ok {
		log.Println("Invalid data format for uiData: expected []TableRow.")
		return
	}

	ui.table.Clear()
	rowIndex := 0

	var currentMetric string

	headerStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue)

	for _, row := range uiData {
		if row.MetricName != currentMetric {
			ui.table.SetCell(rowIndex, 0, tview.NewTableCell(fmt.Sprintf("%s [gray](%s)", row.MetricName, row.Type)).
				SetStyle(headerStyle).
				SetSelectable(false).
				SetTextColor(tcell.ColorWhite).
				SetBackgroundColor(tcell.ColorDarkBlue).
				SetExpansion(2).
				SetAlign(tview.AlignLeft))
			ui.table.SetCell(rowIndex, 1, tview.NewTableCell("").
				SetStyle(headerStyle).
				SetSelectable(false).
				SetTextColor(tcell.ColorWhite).
				SetBackgroundColor(tcell.ColorDarkBlue).
				SetAlign(tview.AlignLeft))

			rowIndex++
			currentMetric = row.MetricName
		}

		labelString := labelsToString(row.Labels)

		ui.table.SetCell(rowIndex, 0, tview.NewTableCell(labelString))

		ui.table.SetCell(rowIndex, 1, tview.NewTableCell(formatValue(row.Value)))
		rowIndex++
	}
	ui.updateLastUpdate()
}

func formatValue(value string) string {
	newValue := value
	if strings.Contains(value, ".") {
		f, _ := strconv.ParseFloat(value, 32)
		if strings.Contains(value, "e+") {
			newValue = fmt.Sprintf("%.0f", f)
		} else {
			newValue = fmt.Sprintf("%.2f", f)
		}
	}
	return newValue
}

func labelsToString(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	labelString := ""

	for _, key := range keys {
		if labels[key] != "" {
			labelString = fmt.Sprintf("%s [yellow]%s: [white]%s", labelString, key, labels[key])
		}
	}

	return labelString
}

func (ui *UI) SetFilterHandler(handler func(string)) {
	ui.filterHandler = handler
}

func (ui *UI) SetSortHandler(handler func(column int, ascending bool)) {
	ui.sortHandler = handler
}

func (ui *UI) handleKeyEvents(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q':
		ui.app.Stop()
	case '1':
		ui.ToggleSort(0)
	case '2':
		ui.ToggleSort(1)
	case '3':
		ui.ToggleSort(2)
	case '/':
		ui.openFilterInput()
		return nil
	}
	return event
}

func (ui *UI) openFilterInput() {
	inputField := tview.NewInputField()
	inputField.
		SetLabel("Filter (regex): ").
		SetText(ui.filterText).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				text := inputField.GetText()
				ui.filterText = text
				if ui.filterHandler != nil {
					ui.filterHandler(text)
				}
			}
			ui.updateFilterFlex()
			ui.app.SetRoot(ui.pages, true).SetFocus(ui.table)
		})
	s := tcell.Style.Background(tcell.Style{}, tcell.ColorDarkCyan)
	inputField.SetLabelStyle(s)
	ui.filterFlex.Clear()
	ui.filterFlex.SetBackgroundColor(tcell.ColorDarkCyan)
	ui.filterFlex.AddItem(inputField, 0, 1, true)
	ui.app.SetRoot(ui.pages, true).SetFocus(inputField)
}

func (ui *UI) ToggleSort(column int) {
	ui.sortAsc = !ui.sortAsc
	if ui.sortHandler != nil {
		ui.sortHandler(column, ui.sortAsc)
	}
}
