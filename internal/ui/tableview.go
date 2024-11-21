package ui

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TableRow struct {
	MetricName string
	Label      string
	Value      string
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

// func (vt *VirtualTableView) Run(observeChan <-chan interface{}) {
// 	go func() {
// 		for newData := range observeChan {
// 			vt.updateTable(newData) // Update table on data change
// 			vt.app.Draw()           // Force redraw
// 		}
// 	}()

// 	// Set initial focus on the table
// 	vt.app.SetRoot(vt.table, true).SetFocus(vt.table)

// 	// Refresh the table with an initial redraw
// 	vt.app.Draw()

// 	// Start the application
// 	if err := vt.app.Run(); err != nil {
// 		panic(err)
// 	}
// }

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
	vt.table.SetCell(0, 0, tview.NewTableCell("Metric").SetSelectable(false))
	vt.table.SetCell(0, 1, tview.NewTableCell("Label").SetSelectable(false))
	vt.table.SetCell(0, 2, tview.NewTableCell("Value").SetSelectable(false))

	for i, row := range uiData {
		vt.table.SetCell(i+1, 0, tview.NewTableCell(row.MetricName))
		vt.table.SetCell(i+1, 1, tview.NewTableCell(row.Label))
		vt.table.SetCell(i+1, 2, tview.NewTableCell(row.Value))
	}
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
