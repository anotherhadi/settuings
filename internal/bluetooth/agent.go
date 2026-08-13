package bluetooth

import (
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	agentPath       = dbus.ObjectPath("/org/anotherhadi/settuings/agent")
	agentCapability = "DisplayYesNo"
)

// agent implements the BlueZ org.bluez.Agent1 interface, auto-accepting
// every pairing prompt. `bluetoothctl pair` run non-interactively never
// registers an agent, so BlueZ has nobody to ask for PIN/passkey
// confirmation and any device that isn't "Just Works" fails with
// AuthenticationFailed. Registering this agent as the system default
// fixes that even though pairing itself still happens in a separate
// bluetoothctl process: BlueZ routes auth requests to whichever agent is
// registered as default, not to whoever initiated the pairing.
type agent struct{}

func (agent) Release() *dbus.Error { return nil }

func (agent) RequestPinCode(_ dbus.ObjectPath) (string, *dbus.Error) {
	return "", dbus.NewError("org.bluez.Error.Rejected", nil)
}

func (agent) DisplayPinCode(_ dbus.ObjectPath, _ string) *dbus.Error { return nil }

func (agent) RequestPasskey(_ dbus.ObjectPath) (uint32, *dbus.Error) {
	return 0, dbus.NewError("org.bluez.Error.Rejected", nil)
}

func (agent) DisplayPasskey(_ dbus.ObjectPath, _ uint32, _ uint16) *dbus.Error { return nil }

func (agent) RequestConfirmation(_ dbus.ObjectPath, _ uint32) *dbus.Error { return nil }

func (agent) RequestAuthorization(_ dbus.ObjectPath) *dbus.Error { return nil }

func (agent) AuthorizeService(_ dbus.ObjectPath, _ string) *dbus.Error { return nil }

func (agent) Cancel() *dbus.Error { return nil }

var (
	agentOnce sync.Once
	agentErr  error
	// agentConn is kept alive for the process lifetime: BlueZ calls back
	// into our agent over this connection while pairing is in progress.
	agentConn *dbus.Conn
)

// ensureAgent registers this process as BlueZ's default pairing agent, once
// per run. Call before any Pair.
func ensureAgent() error {
	agentOnce.Do(func() {
		conn, err := dbus.ConnectSystemBus()
		if err != nil {
			agentErr = fmt.Errorf("connect to system bus: %w", err)
			return
		}

		if err := conn.Export(agent{}, agentPath, "org.bluez.Agent1"); err != nil {
			agentErr = fmt.Errorf("export agent: %w", err)
			return
		}

		manager := conn.Object("org.bluez", dbus.ObjectPath("/org/bluez"))
		if call := manager.Call("org.bluez.AgentManager1.RegisterAgent", 0, agentPath, agentCapability); call.Err != nil {
			agentErr = fmt.Errorf("register agent: %w", call.Err)
			return
		}
		if call := manager.Call("org.bluez.AgentManager1.RequestDefaultAgent", 0, agentPath); call.Err != nil {
			agentErr = fmt.Errorf("request default agent: %w", call.Err)
			return
		}

		agentConn = conn
	})
	return agentErr
}
