//go:build linux && !cli

package main

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (g *linuxGUI) buildCLI() fyne.CanvasObject {
	command := widget.NewEntry()
	command.SetText("muksmatt-cli --help")
	output := widget.NewMultiLineEntry()
	output.Wrapping = fyne.TextWrapWord
	var cancel context.CancelFunc
	var run *widget.Button
	stop := widget.NewButton("Cancel", func() {
		if cancel != nil {
			cancel()
		}
	})
	stop.Disable()
	run = widget.NewButton("Run", func() {
		line := command.Text
		ctx, c := context.WithCancel(context.Background())
		cancel = c
		command.Disable()
		run.Disable()
		stop.Enable()
		output.SetText("Running…")
		go func() {
			text, err := runInAppCLI(ctx, line)
			c()
			fyne.Do(func() {
				if err != nil {
					text += "\n" + err.Error()
				}
				output.SetText(text)
				command.Enable()
				run.Enable()
				stop.Disable()
				cancel = nil
			})
		}()
	})
	examples := blurayCLIExamples()
	presets := container.NewHBox(
		widget.NewButton("Blu-ray Scan", func() { command.SetText(examples[0]) }),
		widget.NewButton("Blu-ray Tracks", func() { command.SetText(examples[1]) }),
		widget.NewButton("Blu-ray REMUX", func() { command.SetText(examples[2]) }),
		widget.NewButton("Blu-ray DEMUX", func() { command.SetText(examples[3]) }),
	)
	help := widget.NewLabel("DVD and Blu-ray CLI — choose an example, replace SOURCE/OUTPUT_DIR, then Run. Scan first; explicitly select the correct Blu-ray playlist (e.g. 00800). Use --help for all options.")
	help.Wrapping = fyne.TextWrapWord
	return container.NewBorder(container.NewVBox(help, presets, command, container.NewHBox(run, stop)), nil, nil, nil, output)
}
