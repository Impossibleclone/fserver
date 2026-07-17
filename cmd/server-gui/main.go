package main

import (
	"context"
	"crypto/tls"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/impossibleclone/fserver/internal/auth"
	"github.com/impossibleclone/fserver/internal/config"
	"github.com/impossibleclone/fserver/internal/server"
)

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	a := app.New()
	w := a.NewWindow("Zoho SETU - Server Administrator")
	w.Resize(fyne.NewSize(600, 400)) // Smaller size since no client browser!

	cfg := config.DefaultConfig()
	
	authenticator := auth.NewMemoryAuth()
	authenticator.AddUser(cfg.Username, cfg.Password)
	
	var srv *server.FileServer

	// ==========================================
	// TAB 1: SERVER CONFIGURATION
	// ==========================================
	portEntry := widget.NewEntry(); portEntry.SetText(cfg.Port)
	dirEntry := widget.NewEntry(); dirEntry.SetText(cfg.StorageDir)
	statusLabel := widget.NewLabel("Server Status: Stopped")

	startButton := widget.NewButtonWithIcon("Start Server", theme.MediaPlayIcon(), nil)
	stopButton := widget.NewButtonWithIcon("Stop Server", theme.MediaStopIcon(), nil)
	stopButton.Disable()

	startButton.OnTapped = func() {
		cfg.Port = portEntry.Text
		cfg.StorageDir = dirEntry.Text
		srv = server.NewFileServer(cfg, authenticator)
		go func() {
			if err := srv.Start(); err != nil && err != http.ErrServerClosed {
				statusLabel.SetText("Error: " + err.Error())
			}
		}()
		statusLabel.SetText("Server Status: Running securely on port " + cfg.Port)
		startButton.Disable()
		stopButton.Enable()
	}

	stopButton.OnTapped = func() {
		if srv != nil {
			srv.Stop(context.Background())
		}
		statusLabel.SetText("Server Status: Stopped")
		startButton.Enable()
		stopButton.Disable()
	}

	serverTab := container.NewVBox(
		widget.NewLabelWithStyle("Server Configuration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Start the backend server. Clients connect via Web Browser."),
		widget.NewSeparator(),
		container.NewGridWithColumns(2, widget.NewLabel("Port:"), portEntry),
		container.NewGridWithColumns(2, widget.NewLabel("Storage Dir:"), dirEntry),
		statusLabel,
		container.NewHBox(startButton, stopButton),
	)

	// ==========================================
	// TAB 2: LOCAL USER MANAGEMENT
	// ==========================================
	newUserEntry := widget.NewEntry()
	newPassEntry := widget.NewPasswordEntry()
	
	var currentUsers []string
	var selectedUser string

	userListWidget := widget.NewList(
		func() int { return len(currentUsers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			u := currentUsers[i]
			info := authenticator.GetUserInfo(u)
			status := " [Offline]"
			if info.IsActive {
				status = " [ACTIVE]"
			}
			o.(*widget.Label).SetText(u + status)
		},
	)
	
	userListWidget.OnSelected = func(id widget.ListItemID) {
		selectedUser = currentUsers[id]
	}

	refreshUsers := func() {
		currentUsers = authenticator.GetUsers()
		userListWidget.Refresh()
	}
	
	addUserBtn := widget.NewButtonWithIcon("Add Secure User", theme.ContentAddIcon(), func() {
		if newUserEntry.Text != "" && newPassEntry.Text != "" {
			err := authenticator.AddUser(newUserEntry.Text, newPassEntry.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			newUserEntry.SetText("")
			newPassEntry.SetText("")
			refreshUsers()
			dialog.ShowInformation("Success", "User added (password securely hashed with bcrypt)", w)
		}
	})

	deleteUserBtn := widget.NewButtonWithIcon("Delete Selected User", theme.DeleteIcon(), func() {
		if selectedUser == "" {
			dialog.ShowInformation("Notice", "Please select a user to delete.", w)
			return
		}
		if selectedUser == cfg.Username {
			dialog.ShowInformation("Notice", "Cannot delete the default admin account from here.", w)
			return
		}
		dialog.ShowConfirm("Confirm", "Delete user "+selectedUser+"?", func(b bool) {
			if b {
				authenticator.RemoveUser(selectedUser)
				refreshUsers()
			}
		}, w)
	})

	refreshUsersBtn := widget.NewButtonWithIcon("Refresh Status", theme.ViewRefreshIcon(), refreshUsers)

	usersTab := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Add New Authenticated User", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewGridWithColumns(2, widget.NewLabel("Username:"), newUserEntry),
			container.NewGridWithColumns(2, widget.NewLabel("Password:"), newPassEntry),
			addUserBtn,
			widget.NewSeparator(),
			container.NewHBox(widget.NewLabel("Account List (Select to Manage):"), refreshUsersBtn),
		),
		container.NewHBox(deleteUserBtn),
		nil, nil,
		userListWidget,
	)

	// ==========================================
	// APP TABS
	// ==========================================
	// Notice: The Network Client has been completely removed! 
	// This makes it a true dedicated server application.
	tabs := container.NewAppTabs(
		container.NewTabItem("Server Admin", serverTab),
		container.NewTabItem("User Management", usersTab),
	)
	
	refreshUsers()

	w.SetContent(tabs)
	w.ShowAndRun()
}
