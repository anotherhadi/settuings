package input

import "testing"

func f(v float64) *float64 { return &v }

func TestParse(t *testing.T) {
	raw := []rawDevice{
		{Name: "AT Translated Set 2 keyboard", Type: "keyboard", LayoutName: []string{"French"}},
		{Name: "Razer Ornata", Type: "keyboard", LayoutName: []string{"French"}},
		{Name: "Razer Ornata", Type: "pointer", Libinput: rawDevice{}.Libinput},
		{Name: "Generic USB Audio", Type: "keyboard", LayoutName: []string{"French"}},
		{Name: "Office (SONOS) (AVRCP)", Type: "keyboard", LayoutName: []string{"French"}},
		{Name: "Power Button", Type: "keyboard", LayoutName: []string{"French"}},
		{Name: "Logitech Mouse", Type: "pointer"},
		{Name: "Touchpad", Type: "touchpad"},
	}
	raw[2].Libinput.AccelSpeed = f(0.3)
	raw[6].Libinput.AccelSpeed = f(-0.2)
	raw[7].Libinput.AccelSpeed = f(0.5)
	st := parse(raw)
	if st.Keyboard == nil || len(st.Keyboard.Devices) != 2 || st.Keyboard.Layout() != "French" {
		t.Fatalf("keyboard = %+v", st.Keyboard)
	}
	if st.Mouse == nil || len(st.Mouse.Devices) != 1 || st.Mouse.Devices[0] != "Logitech Mouse" || st.Mouse.Speed != -0.2 {
		t.Fatalf("mouse = %+v", st.Mouse)
	}
	if st.Touchpad == nil || st.Touchpad.Speed != 0.5 {
		t.Fatalf("touchpad = %+v", st.Touchpad)
	}
}

func TestValidLayouts(t *testing.T) {
	for s, ok := range map[string]bool{"fr": true, "fr,us": true, "fr(oss)": true, "": false, "fr;exec x": false, `fr"`: false} {
		if (ValidLayouts(s) == nil) != ok {
			t.Errorf("ValidLayouts(%q) = %v", s, ValidLayouts(s))
		}
	}
}
