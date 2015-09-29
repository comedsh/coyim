package gui

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/twstrike/coyim/config"
	"github.com/twstrike/coyim/session"
	"github.com/twstrike/coyim/xmpp"

	"github.com/twstrike/go-gtk/gdk"
	"github.com/twstrike/go-gtk/glib"
	"github.com/twstrike/go-gtk/gtk"
	"github.com/twstrike/otr3"
)

type gtkUI struct {
	roster *Roster
	window *gtk.Window

	configFileManager *config.ConfigFileManager
	multiConfig       *config.MultiAccountConfig
	accounts          []Account
}

func NewGTK() *gtkUI {
	return &gtkUI{}
}

func (ui *gtkUI) LoadConfig(configFile string) {
	var err error
	if ui.configFileManager, err = config.NewConfigFileManager(configFile); err != nil {
		ui.Alert(err.Error())
		return
	}

	if err := ui.configFileManager.ParseConfigFile(); err != nil {
		ui.Alert(err.Error())
		ui.enroll()
		return
	}

	//TODO: REMOVE this
	ui.multiConfig = ui.configFileManager.MultiAccountConfig

	ui.accounts = BuildAccountsFrom(ui.multiConfig)
}

//TODO: Should be per session
func (*gtkUI) Disconnected() {
	//TODO: remove everybody from the roster
	fmt.Println("TODO: Should disconnect the account")
}

func (*gtkUI) RegisterCallback(title, instructions string, fields []interface{}) error {

	//TODO: should open a registration window
	fmt.Println("TODO")
	return nil
}

func (u *gtkUI) MessageReceived(s *session.Session, from, timestamp string, encrypted bool, message []byte) {
	u.roster.MessageReceived(s, from, timestamp, encrypted, message)
}

func (u *gtkUI) NewOTRKeys(uid string, conversation *otr3.Conversation) {
	u.Info(fmt.Sprintf("TODO: notify new keys from %s", uid))
}

func (u *gtkUI) OTREnded(uid string) {
	//TODO: conversation ended
}

func (u *gtkUI) Info(m string) {
	fmt.Println(">>> INFO", m)
}

func (u *gtkUI) Warn(m string) {
	fmt.Println(">>> WARN", m)
}

func (u *gtkUI) Alert(m string) {
	fmt.Println(">>> ALERT", m)
}

func (u *gtkUI) Loop() {
	gtk.Init(&os.Args)
	gdk.ThreadsInit()

	gdk.ThreadsEnter()
	u.mainWindow()
	gtk.Main()
	gdk.ThreadsLeave()
}

func (u *gtkUI) onReceiveSignal(s *glib.Signal, f func()) {
	u.window.Connect(s.Name(), f)
}

func (u *gtkUI) initRoster() {
	u.roster = NewRoster()

	//TODO: Should redraw the roster when any account is disconnected
	//u.onReceiveSignal(DISCONNECTED_SIG, u.roster.Clear)
}

func (u *gtkUI) mainWindow() {
	u.window = gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	u.initRoster()

	menubar := initMenuBar(u)
	vbox := gtk.NewVBox(false, 1)
	vbox.PackStart(menubar, false, false, 0)
	vbox.Add(u.roster.Window)
	u.window.Add(vbox)

	u.window.SetTitle("Coy")
	u.window.Connect("destroy", gtk.MainQuit)
	u.window.SetSizeRequest(200, 600)
	u.window.ShowAll()
}

//TODO: REMOVE ME
func (*gtkUI) AskForPassword(c *config.Config) (string, error) {
	return "", nil
}

func (*gtkUI) askForPassword(connect func(string)) {
	glib.IdleAdd(func() bool {
		dialog := gtk.NewDialog()
		dialog.SetTitle("Password")
		dialog.SetPosition(gtk.WIN_POS_CENTER)
		vbox := dialog.GetVBox()

		vbox.Add(gtk.NewLabel("Password"))
		passwordInput := gtk.NewEntry()
		passwordInput.SetEditable(true)
		passwordInput.SetVisibility(false)
		vbox.Add(passwordInput)

		button := gtk.NewButtonWithLabel("Send")
		button.Connect("clicked", func() {
			go connect(passwordInput.GetText())
			dialog.Destroy()
		})
		vbox.Add(button)

		dialog.ShowAll()
		return false
	})
}

//TODO: Remove?
func (*gtkUI) Enroll(c *config.Config) bool {
	return false
}

func authors() []string {
	if b, err := exec.Command("git", "log").Output(); err == nil {
		lines := strings.Split(string(b), "\n")

		var a []string
		r := regexp.MustCompile(`^Author:\s*([^ <]+).*$`)
		for _, e := range lines {
			ms := r.FindStringSubmatch(e)
			if ms == nil {
				continue
			}
			a = append(a, ms[1])
		}
		sort.Strings(a)
		var p string
		lines = []string{}
		for _, e := range a {
			if p == e {
				continue
			}
			lines = append(lines, e)
			p = e
		}
		lines = append(lines, "STRIKE Team <strike-public(AT)thoughtworks.com>")
		return lines
	}
	return []string{"STRIKE Team <strike-public@thoughtworks.com>"}
}

func aboutDialog() {
	dialog := gtk.NewAboutDialog()
	dialog.SetName("Coy IM!")
	dialog.SetProgramName("Coyim")
	dialog.SetAuthors(authors())
	// dir, _ := path.Split(os.Args[0])
	// imagefile := path.Join(dir, "../../data/coyim-logo.png")
	// pixbuf, _ := gdkpixbuf.NewFromFile(imagefile)
	// dialog.SetLogo(pixbuf)
	dialog.SetLicense(`Copyright (c) 2012 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.`)
	dialog.SetWrapLicense(true)
	dialog.Run()
	dialog.Destroy()
}

func accountDialog(account Account, saveFunction func() error) {
	dialog := gtk.NewDialog()
	dialog.SetTitle("Account Details")
	dialog.SetPosition(gtk.WIN_POS_CENTER)
	vbox := dialog.GetVBox()

	accountLabel := gtk.NewLabel("Account")
	vbox.Add(accountLabel)

	accountInput := gtk.NewEntry()
	accountInput.SetText(account.Account)
	accountInput.SetEditable(true)
	vbox.Add(accountInput)

	vbox.Add(gtk.NewLabel("Password"))
	passwordInput := gtk.NewEntry()
	passwordInput.SetText(account.Password)
	passwordInput.SetEditable(true)
	passwordInput.SetVisibility(false)
	vbox.Add(passwordInput)

	vbox.Add(gtk.NewLabel("Server"))
	serverInput := gtk.NewEntry()
	serverInput.SetText(account.Server)
	serverInput.SetEditable(true)
	vbox.Add(serverInput)

	vbox.Add(gtk.NewLabel("Port"))
	portInput := gtk.NewEntry()
	portInput.SetText(strconv.Itoa(account.Port))
	portInput.SetEditable(true)
	vbox.Add(portInput)

	vbox.Add(gtk.NewLabel("Tor Proxy"))
	proxyInput := gtk.NewEntry()
	if len(account.Proxies) > 0 {
		proxyInput.SetText(account.Proxies[0])
	}
	proxyInput.SetEditable(true)
	vbox.Add(proxyInput)

	alwaysEncrypt := gtk.NewCheckButtonWithLabel("Always Encrypt")
	alwaysEncrypt.SetActive(account.AlwaysEncrypt)
	vbox.Add(alwaysEncrypt)

	button := gtk.NewButtonWithLabel("Save")
	button.Connect("clicked", func() {
		account.Account = accountInput.GetText()
		account.Password = passwordInput.GetText()
		account.Server = serverInput.GetText()

		v, err := strconv.Atoi(portInput.GetText())
		if err == nil {
			account.Port = v
		}

		if len(account.Proxies) == 0 {
			account.Proxies = append(account.Proxies, "")
		}
		account.Proxies[0] = proxyInput.GetText()

		account.AlwaysEncrypt = alwaysEncrypt.GetActive()

		//TODO: handle errors
		saveFunction()
		dialog.Destroy()
	})
	vbox.Add(button)

	dialog.ShowAll()
}

func buildAccountSubmenu(u *gtkUI, account Account) *gtk.MenuItem {
	menuitem := gtk.NewMenuItemWithMnemonic(account.Account)

	accountSubMenu := gtk.NewMenu()
	menuitem.SetSubmenu(accountSubMenu)

	connectItem := gtk.NewMenuItemWithMnemonic("_Connect")
	accountSubMenu.Append(connectItem)

	disconnectItem := gtk.NewMenuItemWithMnemonic("_Disconnect")
	disconnectItem.SetSensitive(false)
	accountSubMenu.Append(disconnectItem)

	connectItem.Connect("activate", func() {
		connectItem.SetSensitive(false)
		u.connect(account)
	})

	disconnectItem.Connect("activate", func() {
		u.disconnect(account)
	})

	connToggle := func() {
		s := account.Session
		connected := s.ConnStatus == session.CONNECTED
		connectItem.SetSensitive(!connected)
		disconnectItem.SetSensitive(connected)
	}

	u.window.Connect(account.Connected.Name(), connToggle)
	u.window.Connect(account.Disconnected.Name(), connToggle)

	editItem := gtk.NewMenuItemWithMnemonic("_Edit")
	editItem.Connect("activate", func() {
		accountDialog(account, u.configFileManager.Save)
	})
	accountSubMenu.Append(editItem)

	return menuitem
}

func initMenuBar(u *gtkUI) *gtk.MenuBar {
	menubar := gtk.NewMenuBar()

	//Config -> Account
	cascademenu := gtk.NewMenuItemWithMnemonic("_Accounts")
	menubar.Append(cascademenu)
	submenu := gtk.NewMenu()
	cascademenu.SetSubmenu(submenu)

	for _, account := range u.accounts {
		submenu.Append(buildAccountSubmenu(u, account))
	}

	//Help -> About
	cascademenu = gtk.NewMenuItemWithMnemonic("_Help")
	menubar.Append(cascademenu)
	submenu = gtk.NewMenu()
	cascademenu.SetSubmenu(submenu)
	menuitem := gtk.NewMenuItemWithMnemonic("_About")
	menuitem.Connect("activate", aboutDialog)
	submenu.Append(menuitem)
	return menubar
}

func (u *gtkUI) SubscriptionRequest(from string) {
}

func (u *gtkUI) ProcessPresence(stanza *xmpp.ClientPresence, gone bool) {

	//TODO: Notify via UI
	jid := xmpp.RemoveResourceFromJid(stanza.From)
	fmt.Println(jid, "is", stanza.Show)
}

func (u *gtkUI) IQReceived(string) {
	//TODO
}

//TODO: we should update periodically (like Pidgin does) if we include the status (online/offline/away) on the label
func (u *gtkUI) RosterReceived(s *session.Session, roster []xmpp.RosterEntry) {
	glib.IdleAdd(func() bool {
		u.roster.Update(s, roster)
		u.roster.Redraw()
		return false
	})
}

func (u *gtkUI) enroll() {
	//TODO: extract to function when implementing "add account"
	u.configFileManager.MultiAccountConfig = &config.MultiAccountConfig{
		Accounts: []config.Config{
			config.Config{},
		},
	}

	//TODO: Remove this
	u.multiConfig = u.configFileManager.MultiAccountConfig

	u.accounts = BuildAccountsFrom(u.multiConfig)

	glib.IdleAdd(func() bool {
		accountDialog(u.accounts[0], u.configFileManager.Save)
		return false
	})
}

func (u *gtkUI) disconnect(account Account) error {
	account.Session.Close()
	u.window.Emit(account.Disconnected.Name())
	return nil
}

func (u *gtkUI) connect(account Account) {
	//TODO find a better place to initialize the eventHandler
	s := account.Session
	s.SessionEventHandler = guiSessionEventHandler{u}

	var registerCallback xmpp.FormCallback
	if *config.CreateAccount {
		registerCallback = u.RegisterCallback
	}

	connectFn := func(password string) {
		err := s.Connect(password, registerCallback)
		if err != nil {
			u.window.Emit(account.Disconnected.Name())
			return
		}

		u.window.Emit(account.Connected.Name())
	}

	//TODO We do not support empty passwords
	if len(account.Password) == 0 {
		u.askForPassword(connectFn)
		return
	}

	go connectFn(account.Password)
}
