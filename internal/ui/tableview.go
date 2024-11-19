package ui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/bvankampen/metrics-viewer/internal/realtimedata"
	"golang.org/x/term"
)

type TableView struct {
	CurrentView interface{}
	mu          sync.Mutex
}

func NewTableView(initialData interface{}) *TableView {
	if initialData == nil {
		initialData = realtimedata.RealTimeData{}
	}
	return &TableView{
		CurrentView: initialData,
	}
}

func (tv *TableView) RenderTable() {
	tv.mu.Lock()
	defer tv.mu.Unlock()

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
	}

	fmt.Print("\033[H\033[2J")

	if tv.CurrentView == nil {
		fmt.Println("No data available.")
		return
	}

	switch data := tv.CurrentView.(type) {
	case realtimedata.RealTimeData:

		for _, metric := range data.Metrics {
			tv.renderMetricTable(metric, width)
			fmt.Println()
		}
	default:
		fmt.Println("Unsupported data type for rendering.")
	}
}

func (tv *TableView) stringifyLabels(labels []realtimedata.RealTimeDataMetricLabel) string {
	var parts []string
	for _, label := range labels {
		parts = append(parts, fmt.Sprintf("%s {%s}", label.Label, label.Value))
	}
	return strings.Join(parts, ", ")
}

func (tv *TableView) calculateWidths(headers []string, rows [][]string, terminalWidth int) []int {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	totalWidth := len(widths) + 3
	for _, w := range widths {
		totalWidth += w
	}

	if totalWidth > terminalWidth {
		excess := totalWidth - terminalWidth
		reduce := excess / len(widths)
		for i := range widths {
			widths[i] -= reduce
			if widths[i] < 10 {
				widths[i] = 10
			}
		}
	}

	return widths
}

func (tv *TableView) printRow(columns []string, widths []int) {
	fmt.Print("\r")
	row := "|"
	for i, col := range columns {
		cell := col
		if len(col) > widths[i] {
			cell = col[:widths[i]-3] + "..."
		}
		row += " " + fmt.Sprintf("%-*s", widths[i], cell) + " |"
	}
	fmt.Println(row)
}

func (tv *TableView) printSeparator(widths []int) {
	fmt.Print("\r")
	separator := "+"
	for _, width := range widths {
		separator += strings.Repeat("-", width+2) + "+"
	}
	fmt.Println(separator)
}

func (tv *TableView) renderMetricTable(metric realtimedata.RealTimeDataMetric, terminalWidth int) {

	fmt.Print("\r")
	fmt.Printf("<%s> %s:\n", metric.Name, metric.Description)

	rows := [][]string{}
	for _, value := range metric.Values {
		labelString := tv.stringifyLabels(value.Labels)
		rows = append(rows, []string{labelString, value.Value})
	}

	headers := []string{"Label", "Value"}
	widths := tv.calculateWidths(headers, rows, terminalWidth)

	tv.printSeparator(widths)
	tv.printRow(headers, widths)
	tv.printSeparator(widths)
	for _, row := range rows {
		tv.printRow(row, widths)
	}
	tv.printSeparator(widths)

	fmt.Println()
}

func (tv *TableView) UpdateView(newData interface{}) {
	tv.mu.Lock()
	tv.CurrentView = newData
	tv.mu.Unlock()
	tv.RenderTable()
}

func (tv *TableView) Run(observeChan <-chan interface{}) {

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-exitChan
		term.Restore(int(os.Stdin.Fd()), oldState)
		fmt.Print("\033[H\033[2J")
		os.Exit(0)
	}()

	go func() {
		for newData := range observeChan {
			tv.UpdateView(newData)
		}
		fmt.Println("Data channel closed.")
	}()

	for {
		input := make([]byte, 1)
		_, err := os.Stdin.Read(input)
		if err != nil {
			panic(err)
		}
		if input[0] == 'q' || input[0] == 'Q' {
			term.Restore(int(os.Stdin.Fd()), oldState)
			fmt.Print("\033[H\033[2J")
			return
		}
	}
}
