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

	agentConn *dbus.Conn
)

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
