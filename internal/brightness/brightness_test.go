package brightness

import "testing"

func TestParse(t *testing.T) {
	devs := parse("amdgpu_bl1,backlight,32768,50%,65535\nintel,backlight,10,5%,200\nkbd,leds,1,100%,1\n\nbad\n")
	if len(devs) != 2 || devs[0] != (Device{"amdgpu_bl1", 50}) || devs[1] != (Device{"intel", 5}) {
		t.Fatalf("devs = %+v", devs)
	}
}
