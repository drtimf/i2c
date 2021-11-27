package main

import (
	"fmt"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type LightSensor struct {
	Accessory   *accessory.Accessory
	LightSensor *service.LightSensor
}

func NewLightSensor(info accessory.Info, lux, min, max, steps float64) (ls *LightSensor) {
	ls = &LightSensor{}
	ls.Accessory = accessory.New(info, accessory.TypeSensor)
	ls.LightSensor = service.NewLightSensor()
	ls.LightSensor.CurrentAmbientLightLevel.SetValue(lux)
	ls.LightSensor.CurrentAmbientLightLevel.SetMinValue(min)
	ls.LightSensor.CurrentAmbientLightLevel.SetMaxValue(max)
	ls.LightSensor.CurrentAmbientLightLevel.SetStepValue(steps)
	ls.Accessory.AddService(ls.LightSensor.Service)
	return
}

type OccupancySensor struct {
	Accessory       *accessory.Accessory
	OccupancySensor *service.OccupancySensor
}

func NewOccupancySensor(info accessory.Info) (oc *OccupancySensor) {
	oc = &OccupancySensor{}
	oc.Accessory = accessory.New(info, accessory.TypeSensor)
	oc.OccupancySensor = service.NewOccupancySensor()
	oc.Accessory.AddService(oc.OccupancySensor.Service)
	return
}

type HomeKitBridge struct {
	bridge      *accessory.Bridge
	temperature *accessory.Thermometer
	lightLevel  *LightSensor
	occupancy   *OccupancySensor
}

func HomeKitBridgeStart(deviceID string, devicePin uint32) (b *HomeKitBridge, err error) {
	b = &HomeKitBridge{}

	b.bridge = accessory.NewBridge(accessory.Info{Name: "Bridge", ID: 1})
	b.temperature = accessory.NewTemperatureSensor(accessory.Info{Name: "Temperature", Manufacturer: "Tim", ID: 2}, 0, -5, 50, 0.1)
	b.lightLevel = NewLightSensor(accessory.Info{Name: "Light", Manufacturer: "Tim", ID: 3}, 0, 0, 100, 0.1)
	b.occupancy = NewOccupancySensor(accessory.Info{Name: "Occupancy", Manufacturer: "Tim", ID: 4})

	hkt, terr := hc.NewIPTransport(hc.Config{StoragePath: "config/bridge", Pin: fmt.Sprintf("%d", devicePin), SetupId: deviceID}, b.bridge.Accessory, b.temperature.Accessory, b.lightLevel.Accessory, b.occupancy.Accessory)
	if terr != nil {
		err = terr
		return
	}

	hc.OnTermination(func() {
		<-hkt.Stop()
	})

	go func() {
		hkt.Start()
	}()

	xhm := CreateXHMUrl(accessory.TypeBridge, HAP_TYPE_IP, devicePin, deviceID)
	qr, _ := GenCLIQRCode(xhm)
	fmt.Println(qr)

	return
}

func (b *HomeKitBridge) SetTemperature(temp float64) {
	b.temperature.TempSensor.CurrentTemperature.SetValue(temp)
}

func (b *HomeKitBridge) SetLightLevel(lightLevel float64) {
	b.lightLevel.LightSensor.CurrentAmbientLightLevel.SetValue(lightLevel)
}

func (b *HomeKitBridge) SetRangeSensor(dist uint16) {
	if dist < 1000 {
		b.occupancy.OccupancySensor.OccupancyDetected.SetValue(1)
	} else {
		b.occupancy.OccupancySensor.OccupancyDetected.SetValue(0)
	}
}
