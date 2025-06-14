package aigit

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type spinnerModel struct {
	spinner  spinner.Model
	message  string
	quitting bool
}

func initialSpinnerModel(message string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{
		spinner:  s,
		message:  message,
		quitting: false,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m spinnerModel) View() string {
	if m.quitting {
		return ""
	}
	return fmt.Sprintf("\r%s %s", m.spinner.View(), m.message)
}

// WithSpinner runs the provided function while showing a spinner with the given message.
// The spinner will be displayed until the function completes or an error occurs.
func WithSpinner(message string, fn func() error) error {
	p := tea.NewProgram(initialSpinnerModel(message))

	// Start the spinner in a goroutine
	go func() {
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running spinner: %v\n", err)
		}
	}()

	// Run the provided function
	err := fn()

	// Stop the spinner
	p.Quit()

	// Add a small delay to ensure the spinner is cleared
	time.Sleep(100 * time.Millisecond)

	return err
}
