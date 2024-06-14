package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type SensorConfiguration struct {
	SensorType string
	Name       string
	I2CAddress uint8
	Server     string
	DeviceType string
	Serial     string
	Password   string
}

type I2cConfiguration struct {
	HomeKitDeviceID   string
	HomeKitDevicePin  uint32
	HomeKitBridgeName string
	SampleTime        uint32
	EnableLifx        bool
	LifxMAC           string
	EnableOLED        bool
	EnableLED         bool
	EnableHDPrice     bool
	DebugOutput       bool

	Sensors []SensorConfiguration
}

func LoadConfiguration(fileName string) (cfg *I2cConfiguration, err error) {
	var f *os.File
	if f, err = os.Open(fileName); err != nil {
		return
	}
	defer f.Close()

	cfg = &I2cConfiguration{}
	err = yaml.NewDecoder(f).Decode(&cfg)
	return
}
