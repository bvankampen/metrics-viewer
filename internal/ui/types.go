package ui

import (
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

type TableRow struct {
	MetricName string
	Type       string
	Labels     map[string]string // Universal labels as key-value pairs
	Value      string            // Main value for the row
}

type UI struct {
	app            *tview.Application
	table          *tview.Table
	pages          *tview.Pages
	filterFlex     *tview.Flex
	filterHandler  func(string)
	filterText     string
	lastUpdateFlex *tview.Flex
	sortHandler    func(column int, ascending bool)
	sortAsc        bool
	sortColumn     int
	ctx            *cli.Context
}
