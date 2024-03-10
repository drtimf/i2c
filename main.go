package main

import (
	"context"
	"fmt"
	"math"
	"time"

	// "i2c/go-piicodev.local"
	"github.com/2tvenom/golifx"
	"github.com/drtimf/go-piicodev"
)

func PrintState(device string, status bool) {
	if status {
		fmt.Printf("Enabled: %s\n", device)
	} else {
		fmt.Printf("Disabled: %s\n", device)
	}
}

// HSV2RGB calculates the red, green and blue (RGB) values from hue, saturation and value (HSV) values.
// The hue is 0-360 and the saturation and value are 0-1.
// The red, green and blue values are 0-1.
func HSV2RGB(hue, saturation, value float64) (red, green, blue float64) {
	h := hue
	if h >= 360.0 {
		h = 0.0
	} else {
		h /= 60.0
	}

	fract := h - math.Floor(h)

	p := value * (1.0 - saturation)
	q := value * (1.0 - saturation*fract)
	t := value * (1.0 - saturation*(1.0-fract))

	if 0.0 <= h && h < 1.0 {
		red = value
		green = t
		blue = p
	} else if 1.0 <= h && h < 2.0 {
		red = q
		green = value
		blue = p
	} else if 2.0 <= h && h < 3.0 {
		red = p
		green = value
		blue = t
	} else if 3.0 <= h && h < 4.0 {
		red = p
		green = q
		blue = value
	} else if 4.0 <= h && h < 5.0 {
		red = t
		green = p
		blue = value
	} else if 5.0 <= h && h < 6.0 {
		red = value
		green = p
		blue = q
	} else {
		red = 0
		green = 0
		blue = 0
	}

	return
}

func main() {
	var err error
	ctx := context.Background()

	var config *I2cConfiguration
	if config, err = LoadConfiguration("config/config.yaml"); err != nil {
		fmt.Println(err)
		return
	}

	prom := PrometheusStart(ctx)

	sensorManagement := NewSensorManagement()

	for _, s := range config.Sensors {
		var err error
		var newSensor Sensor

		switch s.SensorType {
		case "tmp117":
			newSensor, err = NewSensorTMP117(s.Name, s.I2CAddress)
		case "ms5637":
			newSensor, err = NewSensorMS5637(s.Name, s.I2CAddress)
		case "aht10":
			newSensor, err = NewSensorAHT10(s.Name, s.I2CAddress)
		case "veml6030":
			newSensor, err = NewSensorVEML6030(s.Name, s.I2CAddress)
		case "bme280":
			newSensor, err = NewSensorBME280(s.Name, s.I2CAddress)
		case "cap1203":
			newSensor, err = NewSensorCAP1203(s.Name, s.I2CAddress)
		case "vl53l1x":
			newSensor, err = NewSensorVL53L1X(s.Name, s.I2CAddress)
		case "ens160":
			newSensor, err = NewSensorENS160(s.Name, s.I2CAddress)
		case "switch":
			newSensor, err = NewSensorSwitch(s.Name, s.I2CAddress)
		case "potentiometer":
			newSensor, err = NewSensorPotentiometer(s.Name, s.I2CAddress)
		case "pir":
			newSensor, err = NewSensorPIR(s.Name, s.I2CAddress)
		case "bom":
			newSensor, err = NewSensorBOM(s.Name, config.DebugOutput)
		default:
			fmt.Printf("ERROR: Unknown sensor type \"%s\"\n", s.SensorType)
			return
		}

		if err == nil {
			sensorManagement.AddSensor(newSensor)
		} else {
			fmt.Printf("ERROR: Failed to initialize sensor of type \"%s\": %v", s.SensorType, err)
		}
	}

	var haveTemperature, haveLightLevel, haveOccupancy bool
	_, haveTemperature = sensorManagement.GetTemperature()
	_, haveLightLevel = sensorManagement.GetLightLevel()
	_, haveOccupancy = sensorManagement.GetOccupancy()

	fmt.Println("HomeKit bridge device ID:", config.HomeKitDeviceID)
	var hkb *HomeKitBridge
	if hkb, err = HomeKitBridgeStart(ctx, config.HomeKitDeviceID, config.HomeKitDevicePin, config.HomeKitBridgeName,
		haveTemperature, haveLightLevel, haveOccupancy, config.EnableLED); err != nil {
		fmt.Println(err)
		return
	}

	var bulb *golifx.Bulb = nil

	PrintState("Lifx bulb", config.EnableLifx)
	if config.EnableLifx {
		var bulbs []*golifx.Bulb
		if bulbs, err = golifx.LookupBulbs(); err == nil && len(bulbs) > 0 {
			bulb = bulbs[0]

			for _, b := range bulbs {
				fmt.Println(b)
				if b.MacAddress() == config.LifxMAC {
					bulb = b
					fmt.Println(">>> Found", config.LifxMAC)
				}
			}

			fmt.Println(">>> Using", config.LifxMAC)
			fmt.Println(bulb.String())
		}
	}

	var oled *OLEDDisplay = nil
	PrintState("OLED display", config.EnableOLED)
	if config.EnableOLED {
		if oled, err = NewOLEDDisplay(); err != nil {
			fmt.Println(err)
		}
	}

	var led *piicodev.RGBLED = nil
	PrintState("RGB LED", config.EnableLED)
	if config.EnableLED {
		if led, err = piicodev.NewRGBLED(piicodev.RGBLEDAddress, 1); err != nil {
			fmt.Println(err)
			return
		}

		led.SetBrightness(255)
		led.EnablePowerLED(false)

		hkb.OnLampChange(func(hue, saturation float64, brightness int) (err error) {
			red, green, blue := HSV2RGB(hue, float64(saturation)/100.0, float64(brightness)/100.0)
			fmt.Printf("Set lamp (%f,%f,%d) -> (%d,%d,%d)\n", hue, saturation, brightness, byte(red*255.0), byte(green*255.0), byte(blue*255.0))

			led.FillPixels(byte(red*255.0), byte(green*255.0), byte(blue*255.0))
			err = led.Show()
			return
		})
	}

	var hdPriceScanner *HDPriceScanner = nil
	PrintState("HD Price", config.EnableHDPrice)
	if config.EnableHDPrice {
		if hdPriceScanner, err = NewHDPriceScanner(); err != nil {
			fmt.Println(err)
		}
	}

	for {
		/*
			if pir != nil {
					var pirDetected, pirRemoved, pirAvailable bool
					if pirAvailable, pirDetected, pirRemoved, err = pir.GetDebounceEvents(); err != nil {
						fmt.Println("ERROR: pir", err)
					}

					if pirAvailable {
						if pirDetected {
							pirMovement = 1
						}

						if pirRemoved {
							pirMovement = 0
						}
					}

					var detected bool
					if detected, err = pir.GetRawReading(); err != nil {
						fmt.Println("ERROR: pir", err)
					}

					if detected {
						pirMovement = 1
					} else {
						pirMovement = 0
					}
			}
		*/

		sensorManagement.UpdateSensors()

		var pubTemp, pubPressure, pubHumidity, pubLightLevel float64
		var pubOccupancy bool

		pubTemp, _ = sensorManagement.GetTemperature()
		pubLightLevel, _ = sensorManagement.GetLightLevel()
		pubPressure, _ = sensorManagement.GetPressure()
		pubHumidity, _ = sensorManagement.GetHumidity()
		pubOccupancy, _ = sensorManagement.GetOccupancy()

		hkb.SetTemperature(pubTemp)
		hkb.SetLightLevel(pubLightLevel)
		hkb.SetOccupancy(pubOccupancy)

		if hdPriceScanner != nil {
			prom.SetWesternDigitalHDPrice(hdPriceScanner.wd)
		}

		if oled != nil {
			// oled.WriteOLED(fmt.Sprintf("%.2f C\n%.2f hPa\n%.2f lux", tempC, pressure, light))
			if haveLightLevel {
				oled.DisplayTemperature(pubTemp, fmt.Sprintf("%.2f hPa\n%.2f lux", pubPressure, pubLightLevel))
			} else {
				oled.DisplayTemperature(pubTemp, fmt.Sprintf("%.2f hPa\n%.2f rH", pubPressure, pubHumidity))
			}
		}

		powerState := "n/a"
		if bulb != nil {
			var bulbState *golifx.BulbState
			if bulbState, err = bulb.GetColorState(); err == nil {
				if bulbState.Power {
					powerState = "on "
				} else {
					powerState = "off "
				}

				powerState += fmt.Sprintf("(H: %d, S: %d B: %d, K: %d)", bulbState.Color.Hue, bulbState.Color.Saturation, bulbState.Color.Brightness, bulbState.Color.Kelvin)

				if pressType, ok := sensorManagement.GetSwitchPress(); ok {
					if pressType == 1 {
						fmt.Println("Light on")
						bulb.SetColorState(&golifx.HSBK{
							Hue:        5461,
							Saturation: 0,
							Brightness: 65535,
							Kelvin:     4000,
						}, 0)
						bulb.SetPowerState(true)
					}

					if pressType == 2 {
						fmt.Println("Light off")
						bulb.SetPowerState(false)
					}
				}

				if changed, value, ok := sensorManagement.GetPotentiometer(); ok {
					if changed {
						fmt.Println("Light ", value)
						bulb.SetColorState(&golifx.HSBK{
							Hue:        5461,
							Saturation: 0,
							Brightness: value << 6,
							Kelvin:     4000,
						}, 0)
					}
				}

				if status, ok := sensorManagement.GetCapSensorStatus(); ok {
					if status[0] {
						fmt.Println("Light off")
						bulb.SetPowerState(false)
					}

					if status[1] {
						fmt.Println("Low light")
						bulb.SetColorState(&golifx.HSBK{
							Hue:        5461,
							Saturation: 0,
							Brightness: 47185,
							Kelvin:     2000,
						}, 0)
						bulb.SetPowerState(true)
					}

					if status[2] {
						fmt.Println("Full light")
						bulb.SetColorState(&golifx.HSBK{
							Hue:        5461,
							Saturation: 0,
							Brightness: 65535,
							Kelvin:     4000,
						}, 0)
						bulb.SetPowerState(true)
					}
				}
			}
		}

		if config.DebugOutput {
			fmt.Println(sensorManagement.Summary() + fmt.Sprintf(" | bulb: %s", powerState))
		}

		if config.SampleTime > 0 {
			time.Sleep(time.Duration(config.SampleTime) * time.Second)
		} else {
			time.Sleep(5 * time.Second)
		}
	}
}
