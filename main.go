package main

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type cell int64

const MIN_SIZE = 1
const MAX_SIZE = 20

const (
	wall cell = iota
	resource
	pit
	nothing
	robot
)

var (
	wallColor     = color.NRGBA{A: 0}
	resourceColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	pitColor      = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	nothingColor  = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	robotColor    = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
)

type AppState struct {
	cellularField []cell
	rows          int
	columns       int
	inputField    []*widget.Select
}

func NewAppState() *AppState {
	return &AppState{
		cellularField: make([]cell, 0),
		rows:          0,
		columns:       0,
		inputField:    make([]*widget.Select, 0),
	}
}

func StartApp() {
	state := NewAppState()

	app := app.New()
	icon, err := fyne.LoadResourceFromPath("assets/icon.png")
	if err != nil {
		panic(err)
	}

	app.SetIcon(icon)

	mainWindow := app.NewWindow("Main programm")
	startWindow := app.NewWindow("Start programm")

	mainWindow.SetMaster()
	mainWindow.SetFixedSize(true)
	mainWindow.SetCloseIntercept(func() {
		app.Quit()
	})

	startWindow.SetFixedSize(true)
	startWindow.SetCloseIntercept(func() {
		app.Quit()
	})

	startWindow.SetCloseIntercept(func() {
		app.Quit()
	})

	startRender(&app, state, startWindow, mainWindow)

	startWindow.Resize(fyne.NewSize(640, 420))
	startWindow.Show()
	mainWindow.Hide()
	app.Run()
}

func getColor(value int) color.Color {
	res := wallColor

	switch value {
	case 1:
		res = resourceColor
	case 2:
		res = pitColor
	case 3:
		res = nothingColor
	case 4:
		res = robotColor
	}

	return res
}

func fieldRender(grid *fyne.Container, state *AppState) {
	grid.RemoveAll()
	for i := 0; i < len(state.cellularField); i++ {
		fill := getColor(int(state.cellularField[i]))

		box := canvas.NewRectangle(fill)
		box.SetMinSize(fyne.NewSize(62, 62))

		ceil := container.New(layout.NewStackLayout(), box)
		grid.Add(ceil)
	}

}

func mainRender(app *fyne.App, state *AppState, window fyne.Window, prev fyne.Window) {

	title := canvas.NewText("TRANSLATOR", color.White)
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 32
	title.TextStyle.Bold = true

	informWindow := (*app).NewWindow("About")
	informWindow.SetCloseIntercept(func() {
		informWindow.Hide()
	})
	informWindow.Hide()
	informWindow.Resize(fyne.NewSize(640, 480))

	informWindow.SetContent(widget.NewLabel("Test"))

	informationBtn := widget.NewButton(
		"?",
		func() {
			informWindow.Show()
		},
	)

	backBtn := widget.NewButton(
		"Back",
		func() {
			window.Hide()
			informWindow.Hide()
			prev.Show()
		},
	)

	header := container.New(
		layout.NewGridLayout(3),
		backBtn,
		title,
		informationBtn,
	)

	inputCommands := widget.NewEntry()
	inputCommands.Resize(fyne.NewSize(1000, 500))

	field := container.New(
		layout.NewGridLayoutWithColumns(state.columns),
	)

	fieldRender(field, state)

	body := container.NewHBox(
		inputCommands,
		field,
	)

	main_container := container.New(
		layout.NewVBoxLayout(),
		header,
		body,
	)

	window.Resize(fyne.NewSize(640, 480))
	window.SetContent(main_container)
}

func checkField(state *AppState) bool {
	wasRobot := false
	state.cellularField = make([]cell, 0)
	for _, sel := range state.inputField {

		if sel.Selected == "ROBOT" && wasRobot {
			return false
		}

		switch sel.Selected {
		case "NONE":
			state.cellularField = append(state.cellularField, nothing)
		case "WALL":
			state.cellularField = append(state.cellularField, wall)
		case "RESOURCE":
			state.cellularField = append(state.cellularField, resource)
		case "PIT":
			state.cellularField = append(state.cellularField, pit)
		case "ROBOT":
			state.cellularField = append(state.cellularField, robot)
			wasRobot = true
		default:
			return false
		}
	}

	return wasRobot
}

func inputGridRender(state *AppState) *fyne.Container {
	grid := container.New(layout.NewGridLayoutWithColumns(state.columns))
	for i := 0; i < state.columns*state.rows; i++ {
		sel := widget.NewSelect([]string{"NONE", "WALL", "RESOURCE", "PIT", "ROBOT"}, nil)
		state.inputField = append(state.inputField, sel)
		grid.Add(sel)
	}
	return grid
}

func startRender(app *fyne.App, state *AppState, window fyne.Window, next fyne.Window) {
	startContainer := container.New(layout.NewVBoxLayout())

	rowsEntry := widget.NewEntry()
	rowsEntry.PlaceHolder = "..."
	rowsLabel := widget.NewLabel("Rows")

	columnsEntry := widget.NewEntry()
	columnsEntry.PlaceHolder = "..."
	columnsLabel := widget.NewLabel("Columns")

	sizeContainer := container.New(
		layout.NewGridLayoutWithColumns(2),
		rowsLabel,
		columnsLabel,
		rowsEntry,
		columnsEntry,
	)

	filepath := widget.NewLabel("Empty")

	openFileBtn := widget.NewButton(
		"Choose file",
		func() {
			dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
				if uc != nil {
					filepath.SetText(uc.URI().Path())
				}
			},
				window,
			)
		},
	)

	inputFieldGrid := container.New(
		layout.NewGridLayoutWithColumns(state.columns),
	)

	inputFieldGrid.Hide()

	checkSizeBtn := widget.NewButton(
		"Check size",
		func() {
			rows, _ := strconv.ParseInt(rowsEntry.Text, 10, 64)
			columns, _ := strconv.ParseInt(columnsEntry.Text, 10, 64)
			if (rows < MIN_SIZE) || (rows > MAX_SIZE) || (columns < MIN_SIZE) || (columns > MAX_SIZE) {
				rowsEntry.SetText("")
				columnsEntry.SetText("")
				dialog.ShowError(
					fmt.Errorf("ERROR: %s", "Incorrect input size\n 0 < Size < 21"),
					window,
				)
			} else {
				state.columns = int(columns)
				state.rows = int(rows)
				inputFieldGrid.RemoveAll()
				inputFieldGrid.Add(inputGridRender(state))
				inputFieldGrid.Show()
			}
		},
	)

	inputField := container.New(layout.NewVBoxLayout(), sizeContainer, checkSizeBtn, inputFieldGrid)

	openFile := container.New(layout.NewVBoxLayout(), filepath, openFileBtn)

	inputTypeRadio := widget.NewRadioGroup([]string{"File", "Input"}, func(s string) {
		if s == "File" {
			openFile.Show()
			inputField.Hide()
		} else {
			openFile.Hide()
			inputField.Show()
		}
	})
	inputTypeRadio.Selected = "File"
	inputField.Hide()

	confirmBtn := widget.NewButton("Confirm size", func() {
		if inputTypeRadio.Selected == "Input" {
			rows, _ := strconv.ParseInt(rowsEntry.Text, 10, 64)
			columns, _ := strconv.ParseInt(columnsEntry.Text, 10, 64)
			if (rows < MIN_SIZE) || (rows > MAX_SIZE) || (columns < MIN_SIZE) || (columns > MAX_SIZE) {
				dialog.ShowError(
					fmt.Errorf("ERROR: %s", "Incorrect input size\n 0 < Size < 21"),
					window,
				)
				rowsEntry.SetText("")
				columnsEntry.SetText("")
			} else if !checkField(state) {
				dialog.ShowError(
					fmt.Errorf("ERROR: %s", "Incorrect field. Must be 1 robot and selected all fields"),
					window,
				)
			} else {
				state.columns = int(columns)
				state.rows = int(rows)
				window.Hide()
				mainRender(app, state, next, window)
				next.Show()
			}
		} else {
			if _, err := os.Stat(filepath.Text); err != nil {
				dialog.ShowError(
					fmt.Errorf("ERROR: %w", err),
					window,
				)
			} else if !readFieldFromFile(state, filepath.Text) {
				dialog.ShowError(
					fmt.Errorf("ERROR: %s", "Bad format for preset"),
					window,
				)
			} else {
				window.Hide()
				mainRender(app, state, next, window)
				next.Show()
			}
		}

	})

	startContainer.Add(inputTypeRadio)
	startContainer.Add(inputField)
	startContainer.Add(openFile)
	startContainer.Add(confirmBtn)

	window.SetContent(startContainer)
}

func readFieldFromFile(state *AppState, path string) bool {

	file, err := os.Open(path)
	if err != nil {
		return false
	}

	defer file.Close()
	state.cellularField = make([]cell, 0)

	sc := bufio.NewScanner(file)
	buffer := make([]string, 0)
	for sc.Scan() {
		buffer = append(buffer, sc.Text())
	}

	tmp := strings.Fields(buffer[0])
	tmpColumns, err := strconv.ParseInt(tmp[0], 10, 0)
	if err != nil {
		return false
	}
	tmpRows, err := strconv.ParseInt(tmp[1], 10, 0)
	if err != nil {
		return false
	}

	if tmpColumns > MAX_SIZE || tmpColumns < MIN_SIZE || tmpRows > MAX_SIZE || tmpRows < MIN_SIZE {
		return false
	}

	state.columns = int(tmpColumns)
	state.rows = int(tmpRows)

	for rows := 1; rows < len(buffer); rows++ {
		columns := strings.Fields(buffer[rows])
		for col := 0; col < state.columns; col++ {
			tmp, err := strconv.ParseInt(columns[col], 10, 64)
			if err != nil || tmp < 0 || tmp > 4 {
				return false
			}
			state.cellularField = append(state.cellularField, cell(tmp))
		}
	}

	return true
}

func main() {
	StartApp()
}
