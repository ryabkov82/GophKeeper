package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ryabkov82/gophkeeper/internal/tui"
)

func main() {

	/*
	   cfg, err := config.Load()
	   if err != nil {
	       log.Fatalf("Failed to load config: %v", err)
	   }
	*/

	if len(os.Args) > 1 && os.Args[1] == "--tui" {
		//exec.Command("cmd", "/c", "chcp 65001 > nul").Run()

		p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			panic(err)
		}
	}
}
