package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/service"
)

type LightSensor struct {
	A           *accessory.A
	LightSensor *service.LightSensor
}

func NewLightSensor(info accessory.Info, id uint64, lux, min, max, steps float64) (ls *LightSensor) {
	ls = &LightSensor{}
	ls.A = accessory.New(info, accessory.TypeSensor)
	ls.A.Id = id
	ls.LightSensor = service.NewLightSensor()
	ls.LightSensor.CurrentAmbientLightLevel.SetValue(lux)
	ls.LightSensor.CurrentAmbientLightLevel.SetMinValue(min)
	ls.LightSensor.CurrentAmbientLightLevel.SetMaxValue(max)
	ls.LightSensor.CurrentAmbientLightLevel.SetStepValue(steps)
	ls.A.AddS(ls.LightSensor.S)
	return
}

type OccupancySensor struct {
	A               *accessory.A
	OccupancySensor *service.OccupancySensor
}

func NewOccupancySensor(info accessory.Info, id uint64) (oc *OccupancySensor) {
	oc = &OccupancySensor{}
	oc.A = accessory.New(info, accessory.TypeSensor)
	oc.A.Id = id
	oc.OccupancySensor = service.NewOccupancySensor()
	oc.A.AddS(oc.OccupancySensor.S)
	return
}

type HomeKitBridge struct {
	bridge      *accessory.Bridge
	Lamp        *accessory.ColoredLightbulb
	temperature *accessory.Thermometer
	lightLevel  *LightSensor
	occupancy   *OccupancySensor

	setLamp func(hue, saturation float64, brightness int) (err error)
}

func HomeKitBridgeStart(deviceID string, devicePin uint32, bridgeName string, enableTemperature bool, enableLight bool, enableOccupancy bool, enableLED bool) (b *HomeKitBridge, err error) {
	b = &HomeKitBridge{}
	accessories := make([]*accessory.A, 0)

	b.bridge = accessory.NewBridge(accessory.Info{Name: bridgeName})
	b.bridge.Id = 1

	if enableTemperature {
		b.temperature = accessory.NewTemperatureSensor(accessory.Info{Name: "Temperature", Manufacturer: "Tim"})
		b.temperature.Id = 2
		accessories = append(accessories, b.temperature.A)
	}

	if enableLight {
		b.lightLevel = NewLightSensor(accessory.Info{Name: "Light", Manufacturer: "Tim"}, 3, 0, 0, 100, 0.1)
		accessories = append(accessories, b.lightLevel.A)
	}

	if enableOccupancy {
		b.occupancy = NewOccupancySensor(accessory.Info{Name: "Occupancy", Manufacturer: "Tim"}, 4)
		accessories = append(accessories, b.occupancy.A)
	}

	if enableLED {
		b.Lamp = accessory.NewColoredLightbulb(accessory.Info{Name: "Lamp"})
		b.Lamp.Id = 5
		accessories = append(accessories, b.Lamp.A)

		var hue, saturation float64
		var brightness int

		b.Lamp.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
			if on == true {
				fmt.Println("Lamp on")
				if err := b.updateLamp(hue, saturation, brightness); err == nil {
					b.Lamp.Lightbulb.On.SetValue(true)
				}

			} else {
				fmt.Println("Lamp off")
				if err := b.lampOff(); err == nil {
					b.Lamp.Lightbulb.On.SetValue(false)
				}
			}
		})

		b.Lamp.Lightbulb.Hue.OnValueRemoteUpdate(func(v float64) {
			hue = v
			b.updateLamp(hue, saturation, brightness)
		})

		b.Lamp.Lightbulb.Saturation.OnValueRemoteUpdate(func(v float64) {
			saturation = v
			b.updateLamp(hue, saturation, brightness)
		})

		b.Lamp.Lightbulb.Brightness.OnValueRemoteUpdate(func(v int) {
			brightness = v
			b.updateLamp(hue, saturation, brightness)
		})
	}

	fs := hap.NewFsStore("config/bridge")
	server, err := hap.NewServer(fs, b.bridge.A, accessories...)
	if err != nil {
		return
	}

	server.Pin = fmt.Sprintf("%d", devicePin)
	server.SetupId = deviceID

	xhm := CreateXHMUrl(accessory.TypeBridge, HAP_TYPE_IP, devicePin, deviceID)
	qr, _ := GenCLIQRCode(xhm)
	fmt.Println(qr)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		// Stop ivering signals.
		signal.Stop(c)
		// Cance the context to stop the server.
		cancel()
	}()

	go func() {
		server.ListenAndServe(ctx)
	}()

	return
}

func (b *HomeKitBridge) SetTemperature(temp float64) {
	if b.temperature != nil {
		b.temperature.TempSensor.CurrentTemperature.SetValue(temp)
	}
}

func (b *HomeKitBridge) SetLightLevel(lightLevel float64) {
	if b.lightLevel != nil {
		b.lightLevel.LightSensor.CurrentAmbientLightLevel.SetValue(lightLevel)
	}
}

func (b *HomeKitBridge) SetRangeSensor(dist uint16) {
	if b.occupancy != nil {
		if dist < 1000 {
			b.occupancy.OccupancySensor.OccupancyDetected.SetValue(1)
		} else {
			b.occupancy.OccupancySensor.OccupancyDetected.SetValue(0)
		}
	}
}

func (b *HomeKitBridge) SetMovement(m int) {
	if b.occupancy != nil {
		b.occupancy.OccupancySensor.OccupancyDetected.SetValue(m)
	}
}

func (b *HomeKitBridge) OnLampChange(fn func(hue, saturation float64, brightness int) (err error)) {
	b.setLamp = fn
}

func (b *HomeKitBridge) updateLamp(hue, saturation float64, brightness int) (err error) {
	if b.setLamp != nil {
		if err = b.setLamp(hue, saturation, brightness); err == nil {
			b.Lamp.Lightbulb.Hue.SetValue(hue)
			b.Lamp.Lightbulb.Saturation.SetValue(saturation)
			b.Lamp.Lightbulb.Brightness.SetValue(brightness)
		}
	}

	return
}

func (b *HomeKitBridge) lampOff() (err error) {
	if b.setLamp != nil {
		err = b.setLamp(0, 0, 0)
	}

	return
}
