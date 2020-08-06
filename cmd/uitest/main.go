package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kubism/smorgasbord/assets/icon"

	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
	"github.com/getlantern/systray"
)

// TODO: well next... maybe os.Process is nicer?

var children = []int{}

var mainwin *ui.Window

func main() {
	pid := os.Getpid()
	ppid := os.Getppid()
	fmt.Printf("pid: %d, ppid: %d, args: %s\n", pid, ppid, os.Args)
	// handle exit for every process and check if child is dead
	go func() {
		sig := make(chan os.Signal)
		signal.Notify(sig)
		for s := range sig {
			if s == syscall.SIGCHLD {
				children = []int{}
			}
			if s == syscall.SIGQUIT || s == syscall.SIGTERM {
				fmt.Printf("[%d] exit\n", pid)
				// make sure that parent can send signals to the children
				for _, child := range children {
					fmt.Printf("parent send %s to %d\n", s, child)
					syscall.Kill(child, s.(syscall.Signal))
				}
				ui.QueueMain(func() {
					ui.Quit()
				})
				syscall.Exit(0)
			}
		}
	}()
	// only the parent process can do
	if _, isChild := os.LookupEnv("CHILD_ID"); !isChild {
		systray.Run(onReady, onExit)
		for _, child := range children {
			fmt.Printf("parent send %s to %d\n", syscall.SIGQUIT, child)
			syscall.Kill(child, syscall.SIGQUIT)
		}
	} else {
		// otherwise let's show gui
		ui.Main(setup)
	}
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Smorgasbord")
	systray.SetTooltip("Smorgasbord")
	quit := systray.AddMenuItem("Quit", "Quit the whole app")
	open := systray.AddMenuItem("Login", "Login thing")
	checked := systray.AddMenuItem("Unchecked", "Check Me")
	go func() {
		for {
			select {
			case <-checked.ClickedCh:
				if checked.Checked() {
					checked.Uncheck()
					checked.SetTitle("Unchecked")
				} else {
					checked.Check()
					checked.SetTitle("Checked")
				}
			case <-open.ClickedCh:
				fmt.Println("tray.open")
				if len(children) > 0 {
					fmt.Println("tray.checkchildren")
					if syscall.Kill(children[0], syscall.SIGHUP) != nil {
						fmt.Println("tray.childdead")
						children = []int{}
					}
				}
				if len(children) == 0 { // well yes requires work...
					args := append(os.Args, fmt.Sprintf("#child_%d_of_%d", 1, os.Getpid()))
					childENV := []string{
						fmt.Sprintf("CHILD_ID=%d", 1),
					}
					pwd, err := os.Getwd()
					if err != nil {
						fmt.Printf("getwd err: %v\n", err)
						os.Exit(1)
					}
					childPID, _ := syscall.ForkExec(args[0], args, &syscall.ProcAttr{
						Dir: pwd,
						Env: append(os.Environ(), childENV...),
						Sys: &syscall.SysProcAttr{
							Setsid: true,
						},
						Files: []uintptr{0, 1, 2}, // print message to the same pty
					})
					fmt.Printf("parent %d fork %d\n", os.Getpid(), childPID)
					if childPID != 0 {
						children = append(children, childPID)
					}
				}
			case <-quit.ClickedCh:
				fmt.Println("tray.quit")
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
}

func makeBasicControlsPage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)

	hbox.Append(ui.NewButton("Button"), false)
	hbox.Append(ui.NewCheckbox("Checkbox"), false)

	vbox.Append(ui.NewLabel("This is a label. Right now, labels can only span one line."), false)

	vbox.Append(ui.NewHorizontalSeparator(), false)

	group := ui.NewGroup("Entries")
	group.SetMargined(true)
	vbox.Append(group, true)

	group.SetChild(ui.NewNonWrappingMultilineEntry())

	entryForm := ui.NewForm()
	entryForm.SetPadded(true)
	group.SetChild(entryForm)

	entryForm.Append("Entry", ui.NewEntry(), false)
	entryForm.Append("Password Entry", ui.NewPasswordEntry(), false)
	entryForm.Append("Search Entry", ui.NewSearchEntry(), false)
	entryForm.Append("Multiline Entry", ui.NewMultilineEntry(), true)
	entryForm.Append("Multiline Entry No Wrap", ui.NewNonWrappingMultilineEntry(), true)

	return vbox
}

func makeNumbersPage() ui.Control {
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)

	group := ui.NewGroup("Numbers")
	group.SetMargined(true)
	hbox.Append(group, true)

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	group.SetChild(vbox)

	spinbox := ui.NewSpinbox(0, 100)
	slider := ui.NewSlider(0, 100)
	pbar := ui.NewProgressBar()
	spinbox.OnChanged(func(*ui.Spinbox) {
		slider.SetValue(spinbox.Value())
		pbar.SetValue(spinbox.Value())
	})
	slider.OnChanged(func(*ui.Slider) {
		spinbox.SetValue(slider.Value())
		pbar.SetValue(slider.Value())
	})
	vbox.Append(spinbox, false)
	vbox.Append(slider, false)
	vbox.Append(pbar, false)

	ip := ui.NewProgressBar()
	ip.SetValue(-1)
	vbox.Append(ip, false)

	group = ui.NewGroup("Lists")
	group.SetMargined(true)
	hbox.Append(group, true)

	vbox = ui.NewVerticalBox()
	vbox.SetPadded(true)
	group.SetChild(vbox)

	cbox := ui.NewCombobox()
	cbox.Append("Combobox Item 1")
	cbox.Append("Combobox Item 2")
	cbox.Append("Combobox Item 3")
	vbox.Append(cbox, false)

	ecbox := ui.NewEditableCombobox()
	ecbox.Append("Editable Item 1")
	ecbox.Append("Editable Item 2")
	ecbox.Append("Editable Item 3")
	vbox.Append(ecbox, false)

	rb := ui.NewRadioButtons()
	rb.Append("Radio Button 1")
	rb.Append("Radio Button 2")
	rb.Append("Radio Button 3")
	vbox.Append(rb, false)

	return hbox
}

func makeDataChoosersPage() ui.Control {
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	hbox.Append(vbox, false)

	vbox.Append(ui.NewDatePicker(), false)
	vbox.Append(ui.NewTimePicker(), false)
	vbox.Append(ui.NewDateTimePicker(), false)
	vbox.Append(ui.NewFontButton(), false)
	vbox.Append(ui.NewColorButton(), false)

	hbox.Append(ui.NewVerticalSeparator(), false)

	vbox = ui.NewVerticalBox()
	vbox.SetPadded(true)
	hbox.Append(vbox, true)

	grid := ui.NewGrid()
	grid.SetPadded(true)
	vbox.Append(grid, false)

	button := ui.NewButton("Open File")
	entry := ui.NewEntry()
	entry.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		filename := ui.OpenFile(mainwin)
		if filename == "" {
			filename = "(cancelled)"
		}
		entry.SetText(filename)
	})
	grid.Append(button,
		0, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(entry,
		1, 0, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)

	button = ui.NewButton("Save File")
	entry2 := ui.NewEntry()
	entry2.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		filename := ui.SaveFile(mainwin)
		if filename == "" {
			filename = "(cancelled)"
		}
		entry2.SetText(filename)
	})
	grid.Append(button,
		0, 1, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(entry2,
		1, 1, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)

	msggrid := ui.NewGrid()
	msggrid.SetPadded(true)
	grid.Append(msggrid,
		0, 2, 2, 1,
		false, ui.AlignCenter, false, ui.AlignStart)

	button = ui.NewButton("Message Box")
	button.OnClicked(func(*ui.Button) {
		ui.MsgBox(mainwin,
			"This is a normal message box.",
			"More detailed information can be shown here.")
	})
	msggrid.Append(button,
		0, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	button = ui.NewButton("Error Box")
	button.OnClicked(func(*ui.Button) {
		ui.MsgBoxError(mainwin,
			"This message box describes an error.",
			"More detailed information can be shown here.")
	})
	msggrid.Append(button,
		1, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)

	return hbox
}

func setup() {
	mainwin = ui.NewWindow("libui Control Gallery", 640, 480, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		fmt.Println("onclosing")
		syscall.Kill(os.Getppid(), syscall.SIGCHLD)
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		fmt.Println("onshouldquit")
		mainwin.Destroy()
		return true
	})

	tab := ui.NewTab()
	mainwin.SetChild(tab)
	mainwin.SetMargined(true)

	tab.Append("Basic Controls", makeBasicControlsPage())
	tab.SetMargined(0, true)

	tab.Append("Numbers and Lists", makeNumbersPage())
	tab.SetMargined(1, true)

	tab.Append("Data Choosers", makeDataChoosersPage())
	tab.SetMargined(2, true)

	mainwin.Show()
}
