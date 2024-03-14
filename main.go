package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type fruit struct {
	name  string
	color string
}

func (f fruit) Name() string                            { return f.name }
func (f fruit) Color() string                           { return f.color }
func (f fruit) Height() int                             { return 1 }
func (f fruit) Spacing() int                            { return 0 }
func (f fruit) FilterValue() string                     { return f.name }
func (f fruit) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (f fruit) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(fruit)
	if !ok {
		return
	}

	itemStyle := lipgloss.NewStyle()
	if m.Index() == index {
		itemStyle = lipgloss.NewStyle().Reverse(true)
	}

	str := fmt.Sprintf("%s - %s", i.Name(), i.Color())

	fmt.Fprint(w, itemStyle.Render(str))
}

var fruits = []fruit{
	{"Apple", "Green"},
	{"Orange", "Orange"},
	{"Banana", "Yellow"},
	{"Grapes", "Blue"},
    {"Lychee", "White"},
}

type fruitList struct {
	list *list.Model
}

func initialFruitList() *fruitList {
	fl := new(fruitList)
	newList := list.New(nil, fruit{}, 0, 10)
	fl.list = &newList

	fl.list.SetShowHelp(false)
	fl.list.SetShowFilter(false)
	fl.list.SetShowStatusBar(false)
	fl.list.SetShowTitle(false)

	fl.LoadFruits()

	return fl
}

func (fl fruitList) Init() tea.Cmd {
	return nil
}

func (fl fruitList) Update(msg tea.Msg) (fruitList, tea.Cmd) {
	updatedList, cmd := fl.list.Update(msg)
	fl.list = &updatedList

	return fl, cmd
}

func (fl fruitList) View() string {
	return fl.list.View()
}

func (fl *fruitList) LoadFruits() tea.Cmd {
	items := []list.Item{}
	for _, f := range fruits {
		items = append(items, fruit{f.name, f.color})
	}
	return fl.list.SetItems(items)
}

type fruitForm struct {
	form  *huh.Form
	name  *huh.Input
	color *huh.Input
}

func initialFruitForm() *fruitForm {
	ff := new(fruitForm)
	ff.name = huh.NewInput().Title("Name").Key("name")
	ff.color = huh.NewInput().Title("Color").Key("color")
	ff.form = huh.NewForm(
		huh.NewGroup(
			ff.name,
			ff.color)).
		WithShowHelp(false)
	return ff
}

func (ff fruitForm) Init() tea.Cmd {
	return ff.form.Init()
}

type FruitAddedMsg string

func (ff fruitForm) AddFruit() tea.Msg {
	fruits = append(fruits, fruit{
		name:  ff.form.GetString("name"),
		color: ff.form.GetString("color"),
	})

	return FruitAddedMsg("Fruit Added")
}

func (ff fruitForm) Update(msg tea.Msg) (fruitForm, tea.Cmd) {
	updatedForm, cmd := ff.form.Update(msg)
	if f, ok := updatedForm.(*huh.Form); ok {
		ff.form = f
	}
	if ff.form.State == huh.StateCompleted {
		ff.form.State = huh.StateNormal
		return ff, ff.AddFruit
	}
	return ff, cmd
}

func (ff fruitForm) View() string {
	return ff.form.View()
}

func (ff *fruitForm) NewFruit() {
	newFruit := fruit{}
    ff.name.Focus() 
	ff.name.Value(&newFruit.name)
	ff.color.Value(&newFruit.color)
}

func (ff fruitForm) SetValues(f fruit) {
	ff.name.Value(&f.name)
	ff.color.Value(&f.color)
}

type activeComponent int

const (
	listComponent activeComponent = iota
	formComponent
)

type UI struct {
	fruitList *fruitList
	fruitForm *fruitForm
	active    activeComponent
}

func initialUI() (ui UI) {
	ui.fruitList = initialFruitList()
	ui.fruitForm = initialFruitForm()

	return ui
}

func (ui UI) Init() tea.Cmd {
	return ui.fruitForm.Init()
}

func (ui UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		key := msg.String()

		switch key {
		case "n":
			ui.fruitForm.NewFruit()
			tempFruitForm, _ := ui.fruitForm.Update(nil)
			ui.fruitForm = &tempFruitForm
			ui.active = formComponent
			return ui, nil
		}

	case FruitAddedMsg:
		ui.NewForm()
		ui.active = listComponent
		cmd = ui.fruitList.LoadFruits()
		cmds = append(cmds, cmd)
	}

	switch ui.active {

	case listComponent:
		tempFruitList, cmd := ui.fruitList.Update(msg)
		ui.fruitList = &tempFruitList
		cmds = append(cmds, cmd)
		ui.fruitForm.SetValues(fruits[ui.fruitList.list.Index()])

	case formComponent:
		tempFruitForm, cmd := ui.fruitForm.Update(msg)
		ui.fruitForm = &tempFruitForm
		cmds = append(cmds, cmd)
	}

	return ui, tea.Batch(cmds...)
}

func (ui *UI) NewForm() {
	ui.fruitForm = initialFruitForm()
}

func (ui UI) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		ui.fruitList.View(),
		ui.fruitForm.View(),
	)
}

func main() {
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Printf("failed. Aborting.\nError: %v", err)
		return
	}
	log.SetOutput(f)
	log.Println(" ----------------------------------------------------------------")
	ui := initialUI()

	if _, err := tea.NewProgram(ui).Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
