package ui

import (
	"log"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TableRow struct {
	MetricName string
	Labels     map[string]string // Universal labels as key-value pairs
	Value      string            // Main value for the row
}

type VirtualTableView struct {
	app           *tview.Application
	table         *tview.Table
	filterHandler func(string)
	sortHandler   func(column int, ascending bool)
	sortAsc       bool
	sortColumn    int
}

func NewVirtualTableView() *VirtualTableView {
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(true).
		SetFixed(1, 0)

	return &VirtualTableView{
		app:        app,
		table:      table,
		sortAsc:    true,
		sortColumn: 0,
	}
}

func (vt *VirtualTableView) Run(observeChan <-chan interface{}) {
	go func() {
		for data := range observeChan {
			vt.updateTable(data)
			vt.app.Draw()
		}
	}()

	vt.app.SetInputCapture(vt.handleKeyEvents)
	if err := vt.app.SetRoot(vt.table, true).Run(); err != nil {
		panic(err)
	}
}

func (vt *VirtualTableView) updateTable(data interface{}) {
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

	vt.table.Clear()
	rowIndex := 0

	var currentMetric string
	labelKeys := getUniqueLabelKeysFromUIData(uiData)

	for _, row := range uiData {
		if row.MetricName != currentMetric {
			// Add a header row for the metric
			vt.table.SetCell(rowIndex, 0, tview.NewTableCell(row.MetricName).
				SetSelectable(false).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter))
			rowIndex++

			// Add column headers
			colIndex := 0
			for _, key := range labelKeys {
				vt.table.SetCell(rowIndex, colIndex, tview.NewTableCell(key).SetSelectable(false))
				colIndex++
			}
			vt.table.SetCell(rowIndex, colIndex, tview.NewTableCell("Value").SetSelectable(false))
			rowIndex++
			currentMetric = row.MetricName
		}

		// Add data rows
		colIndex := 0
		for _, key := range labelKeys {
			value := row.Labels[key] // Match value with its header
			vt.table.SetCell(rowIndex, colIndex, tview.NewTableCell(value))
			colIndex++
		}
		vt.table.SetCell(rowIndex, colIndex, tview.NewTableCell(row.Value))
		rowIndex++
	}
}

// Helper function: Extract unique keys from TableRow data
func getUniqueLabelKeysFromUIData(uiData []TableRow) []string {
	labelSet := make(map[string]struct{})
	for _, row := range uiData {
		for key := range row.Labels {
			labelSet[key] = struct{}{}
		}
	}

	labelKeys := make([]string, 0, len(labelSet))
	for key := range labelSet {
		labelKeys = append(labelKeys, key)
	}
	sort.Strings(labelKeys) // Ensure consistent order
	return labelKeys
}

func (vt *VirtualTableView) SetFilterHandler(handler func(string)) {
	vt.filterHandler = handler
}

func (vt *VirtualTableView) SetSortHandler(handler func(column int, ascending bool)) {
	vt.sortHandler = handler
}

func (vt *VirtualTableView) handleKeyEvents(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q':
		vt.app.Stop()
	case '1':
		vt.ToggleSort(0)
	case '2':
		vt.ToggleSort(1)
	case '3':
		vt.ToggleSort(2)
	case '/':
		vt.openFilterInput()
	}
	return event
}

func (vt *VirtualTableView) openFilterInput() {
	inputField := tview.NewInputField()
	inputField.
		SetLabel("Filter (regex): ").
		SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				text := inputField.GetText()
				if vt.filterHandler != nil {
					vt.filterHandler(text)
				}
			}
			vt.app.SetRoot(vt.table, true).SetFocus(vt.table)
		})
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(vt.table, 0, 1, false).
		AddItem(inputField, 1, 0, true)
	vt.app.SetRoot(layout, true).SetFocus(inputField)
}

func (vt *VirtualTableView) ToggleSort(column int) {
	vt.sortAsc = !vt.sortAsc
	if vt.sortHandler != nil {
		vt.sortHandler(column, vt.sortAsc)
	}
}
