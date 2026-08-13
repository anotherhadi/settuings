package app

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/config"
	aboutUI "github.com/anotherhadi/settuings/internal/ui/about"
	audioUI "github.com/anotherhadi/settuings/internal/ui/audio"
	bluetoothUI "github.com/anotherhadi/settuings/internal/ui/bluetooth"
	notificationsUI "github.com/anotherhadi/settuings/internal/ui/components/notifications"
	inputUI "github.com/anotherhadi/settuings/internal/ui/input"
	networkUI "github.com/anotherhadi/settuings/internal/ui/network"
	powerUI "github.com/anotherhadi/settuings/internal/ui/power"
	"github.com/sirupsen/logrus"
)

// buildPageShortcuts numbers the visible pages 1..N in sidebar order, so
// hiding a page (or a missing CLI tool) never leaves a gap in the
// shortcuts.
func buildPageShortcuts(visible []pageEntry) map[string]page {
	m := make(map[string]page, len(visible))
	for i, e := range visible {
		m[strconv.Itoa(i+1)] = e.id
	}
	return m
}

type Model struct {
	page          page
	pageShortcuts map[string]page
	visiblePages  []pageEntry
	logPath       string
	fatalErr      error
	logFileErr    error

	width         int
	height        int
	sidebarState  sidebarState
	about         aboutUI.Model
	network       networkUI.Model
	bluetooth     bluetoothUI.Model
	audio         audioUI.Model
	power         powerUI.Model
	input         inputUI.Model
	notifications notificationsUI.Model
}

const logPath = "/tmp/settuings.log"

// New creates the app's root model. initialPage selects the page shown on
// launch (matched case-insensitively against PageNames); an empty string
// falls back to the default page. It returns an error if initialPage is
// non-empty and does not match a known page.
func New(initialPage string) (Model, error) {
	startPage := pageAbout
	if initialPage != "" {
		p, ok := lookupPage(initialPage)
		if !ok {
			return Model{}, fmt.Errorf("unknown page %q (valid: %s)", initialPage, strings.Join(PageNames(), ", "))
		}
		startPage = p
	}

	cfg := config.Global
	visible := visiblePages()

	m := Model{
		page:          startPage,
		pageShortcuts: buildPageShortcuts(visible),
		visiblePages:  visible,
		about:         aboutUI.New(),
		network:       networkUI.New(),
		bluetooth:     bluetoothUI.New(),
		audio:         audioUI.New(),
		power:         powerUI.New(),
		input:         inputUI.New(),
		notifications: notificationsUI.New(),
		sidebarState:  sidebarState(cfg.TUI.DefaultSidebarState),
		logPath:       logPath,
	}

	if lf, err := os.OpenFile(m.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); err == nil {
		log.SetOutput(lf)
		logrus.SetOutput(lf)
	} else {
		m.logFileErr = err
	}

	return m, nil
}

func (m Model) FatalErr() error { return m.fatalErr }

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.about.Init(), m.network.Init(), m.bluetooth.Init(), m.audio.Init(), m.power.Init(), m.input.Init()}
	if m.page != pageAbout {
		cmds = append(cmds, m.activatePage(m.page))
	}
	if m.logFileErr != nil {
		err := m.logFileErr
		cmds = append(cmds, func() tea.Msg {
			return notificationsUI.NotificationMsg{
				Title: "Warning",
				Body:  "Could not open log file: " + err.Error(),
				Kind:  notificationsUI.KindWarning,
			}
		})
	}
	return tea.Batch(cmds...)
}
