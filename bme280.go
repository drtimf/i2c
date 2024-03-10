package main

import (
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

type BME280 struct {
	bus i2c.BusCloser
	dev *bmxx80.Dev
}

func NewBME280(address uint8) (d *BME280, err error) {
	d = &BME280{}

	if _, err = host.Init(); err != nil {
		return
	}

	if d.bus, err = i2creg.Open(""); err != nil {
		return
	}

	if d.dev, err = bmxx80.NewI2C(d.bus, uint16(address), &bmxx80.DefaultOpts); err != nil {
		return
	}

	return
}

func (d *BME280) Read() (temperature float64, pressure float64, humidity float64, err error) {
	e := physic.Env{}
	if err = d.dev.Sense(&e); err != nil {
		return
	}

	temperature = e.Temperature.Celsius()
	pressure = float64(e.Pressure) / float64(physic.Pascal*100)
	humidity = float64(e.Humidity) / float64(physic.PercentRH)
	return
}
