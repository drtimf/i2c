package main

import (
	"fmt"

	"github.com/drtimf/go-piicodev"
)

type SensorSwitch struct {
	name             string
	sw               *piicodev.Switch
	wasPressed       bool
	wasDoublePressed bool
}

func NewSensorSwitch(name string, i2cAddress uint8) (s *SensorSwitch, err error) {
	s = &SensorSwitch{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.SwitchAddress
	}

	if s.sw, err = piicodev.NewSwitch(i2cAddress, 1); err != nil {
		return
	}

	s.sw.SetLED(false)

	return
}

func (s *SensorSwitch) Update() {
	var err error

	if s.wasPressed, err = s.sw.WasPressed(); err != nil {
		fmt.Printf("ERROR: Failed to read pressed status from Switch \"%s\": %v\n", s.name, err)
	}

	if s.wasDoublePressed, err = s.sw.WasDoublePressed(); err != nil {
		fmt.Printf("ERROR: Failed to read double-pressed status from Switch \"%s\": %v\n", s.name, err)
	}
}

func (s *SensorSwitch) Summary() string {
	return fmt.Sprintf("%s: %t,%t", s.name, s.wasPressed, s.wasDoublePressed)
}

func (s *SensorSwitch) Details() string {
	return fmt.Sprintf("%s - Switch: %t,%t", s.name, s.wasPressed, s.wasDoublePressed)
}
