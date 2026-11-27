package power

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const powerSupplyPath = "/sys/class/power_supply"

type Battery struct {
	Name         string
	Status       string
	Capacity     int
	Health       int
	TimeToEmpty  string
	TimeToFull   string
	Technology   string
	CycleCount   int
	Manufacturer string
	Model        string
}

type ACStatus struct {
	Present bool
	Online  bool
}

func ReadBatteries() ([]Battery, error) {
	entries, err := os.ReadDir(powerSupplyPath)
	if err != nil {
		return nil, err
	}

	var batteries []Battery
	for _, e := range entries {
		dir := filepath.Join(powerSupplyPath, e.Name())
		if readAttr(dir, "type") != "Battery" {
			continue
		}
		if readAttr(dir, "scope") == "Device" {
			continue
		}
		if readAttr(dir, "present") == "0" {
			continue
		}
		batteries = append(batteries, readBattery(dir, e.Name()))
	}
	sort.Slice(batteries, func(i, j int) bool { return batteries[i].Name < batteries[j].Name })
	return batteries, nil
}

func readBattery(dir, name string) Battery {
	status := readAttrDefault(dir, "status", "Unknown")

	full := readAttrInt(dir, "energy_full", 0)
	design := readAttrInt(dir, "energy_full_design", 0)
	now := readAttrInt(dir, "energy_now", 0)
	rate := readAttrInt(dir, "power_now", 0)
	if full == 0 {
		full = readAttrInt(dir, "charge_full", 0)
		design = readAttrInt(dir, "charge_full_design", 0)
		now = readAttrInt(dir, "charge_now", 0)
		rate = readAttrInt(dir, "current_now", 0)
	}

	capacity := readAttrInt(dir, "capacity", -1)
	if capacity < 0 && full > 0 {
		capacity = now * 100 / full
	}

	health := -1
	if full > 0 && design > 0 {
		health = full * 100 / design
	}

	b := Battery{
		Name:         name,
		Status:       status,
		Capacity:     capacity,
		Health:       health,
		Technology:   readAttr(dir, "technology"),
		CycleCount:   readAttrInt(dir, "cycle_count", -1),
		Manufacturer: readAttr(dir, "manufacturer"),
		Model:        readAttr(dir, "model_name"),
	}

	if rate > 0 {
		switch status {
		case "Discharging":
			b.TimeToEmpty = formatHours(float64(now) / float64(rate))
		case "Charging":
			if full > now {
				b.TimeToFull = formatHours(float64(full-now) / float64(rate))
			}
		}
	}

	return b
}

func ReadAC() (ACStatus, error) {
	entries, err := os.ReadDir(powerSupplyPath)
	if err != nil {
		return ACStatus{}, err
	}

	var status ACStatus
	for _, e := range entries {
		dir := filepath.Join(powerSupplyPath, e.Name())
		t := readAttr(dir, "type")
		if t != "Mains" && t != "USB" {
			continue
		}
		online := readAttr(dir, "online")
		if online == "" {
			continue
		}
		status.Present = true
		if online == "1" {
			status.Online = true
		}
	}
	return status, nil
}

func formatHours(hours float64) string {
	if hours <= 0 {
		return ""
	}
	totalMinutes := int(hours*60 + 0.5)
	return fmt.Sprintf("%dh%02dm", totalMinutes/60, totalMinutes%60)
}

func readAttr(dir, name string) string {
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readAttrDefault(dir, name, def string) string {
	if v := readAttr(dir, name); v != "" {
		return v
	}
	return def
}

func readAttrInt(dir, name string, def int) int {
	v := readAttr(dir, name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
