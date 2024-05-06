package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"Go_emu/src/cpu"
	"strings"
	"time"
)

type Appmodel struct {
	emulator *cpu.Cpu
}

func initialModel() Appmodel {

	model := Appmodel{emulator: &cpu.Cpu{}}
	model.emulator.LoadRom("../cpu/test_roms/PrintDigits_rom")
	return model
}

type stepMsg time.Time

func stepAnimation() tea.Cmd {
	return tea.Tick(time.Second/90, func(t time.Time) tea.Msg {
		return stepMsg(t)
	})
}
func (m Appmodel) Init() tea.Cmd {

	go func() {
		for {
			m.emulator.ClockCycle()
			time.Sleep(time.Microsecond * 1)
		}

	}()
	return stepAnimation()
}
func (m Appmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		}
	case stepMsg:

		return m, stepAnimation()
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}
func (m Appmodel) View() string {

	view := strings.Builder{}
	counter := 0
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if m.emulator.Ram.GetLine(uint32(30032+counter)) == 1 {
				s := lipgloss.NewStyle().SetString("***").Background(lipgloss.Color("#FAFAFA"))
				view.WriteString(s.String())
			} else {

				s := lipgloss.NewStyle().SetString("***").Background(lipgloss.Color("#7D56F4"))
				view.WriteString(s.String())
			}
			counter += 4
		}
		view.WriteRune('\n')

	}
	return view.String()
}

func main() {

	p := tea.NewProgram(initialModel())
	p.Run()

}
