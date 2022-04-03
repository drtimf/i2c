package main

import (
	"fmt"
	"math"
	"os"
	"time"

	// "i2c/go-piicodev.local"
	"github.com/2tvenom/golifx"
	"github.com/drtimf/go-piicodev"
	"gopkg.in/yaml.v2"
)

type I2cConfiguration struct {
	HomeKitDeviceID   string
	HomeKitDevicePin  uint32
	HomeKitBridgeName string
	SampleTime        uint32
	EnableLifx        bool
	LifxMAC           string
	EnableTMP117      bool
	EnableVEML6030    bool
	EnableVL53L1X     bool
	EnableMS5637      bool
	EnableBME280      bool
	EnableCAP1203     bool
	EnableOLED        bool
	EnableLED         bool
	EnablePIR         bool
}

func LoadI2cConfiguration(fileName string) (cfg *I2cConfiguration, err error) {
	var f *os.File
	if f, err = os.Open(fileName); err != nil {
		return
	}
	defer f.Close()

	cfg = &I2cConfiguration{}
	err = yaml.NewDecoder(f).Decode(&cfg)
	return
}

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

	var config *I2cConfiguration
	if config, err = LoadI2cConfiguration("config/config.yaml"); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("HomeKit bridge device ID:", config.HomeKitDeviceID)
	var hkb *HomeKitBridge
	if hkb, err = HomeKitBridgeStart(config.HomeKitDeviceID, config.HomeKitDevicePin, config.HomeKitBridgeName,
		config.EnableTMP117 || config.EnableBME280, config.EnableVEML6030, config.EnableVL53L1X || config.EnablePIR, config.EnableLED); err != nil {
		fmt.Println(err)
		return
	}

	prom := PrometheusStart()

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
				}
			}

			fmt.Println(bulb.String())
		}
	}

	var tmp117 *piicodev.TMP117 = nil
	PrintState("TMP117 temperature sensor", config.EnableTMP117)
	if config.EnableTMP117 {
		if tmp117, err = piicodev.NewTMP117(piicodev.TMP117Address, 1); err != nil {
			fmt.Println(err)
			return
		}
	}

	var veml6030 *piicodev.VEML6030 = nil
	PrintState("VEML6030 light sensor", config.EnableVEML6030)
	if config.EnableVEML6030 {
		if veml6030, err = piicodev.NewVEML6030(piicodev.VEML6030Address, 1); err != nil {
			fmt.Println(err)
			return
		}
	}

	var vl53l1x *piicodev.VL53L1X = nil
	PrintState("VL53L1X distance sensor", config.EnableVL53L1X)
	if config.EnableVL53L1X {
		if vl53l1x, err = piicodev.NewVL53L1X(piicodev.VL53L1XAddress, 1); err != nil {
			fmt.Println(err)
			return
		}
	}

	var ms5637 *piicodev.MS5637 = nil
	PrintState("MS5637 pressure sensor", config.EnableMS5637)
	if config.EnableMS5637 {
		if ms5637, err = piicodev.NewMS5637(piicodev.MS5637Address, 1); err != nil {
			fmt.Println(err)
			return
		}
	}

	var bme280 *BME280 = nil
	PrintState("BME280 pressure,temperature and humidity sensor", config.EnableBME280)
	if config.EnableBME280 {
		if bme280, err = NewBME280(); err != nil {
			fmt.Println(err)
			return
		}
	}

	var cap1203 *piicodev.CAP1203 = nil
	PrintState("CAP1203 pressure sensor", config.EnableCAP1203)
	if config.EnableCAP1203 {
		if cap1203, err = piicodev.NewCAP1203(piicodev.CAP1203Address, 1); err != nil {
			fmt.Println(err)
			return
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

	var pir *piicodev.QwiicPIR = nil
	PrintState("PIR", config.EnablePIR)
	if config.EnablePIR {
		if pir, err = piicodev.NewQwiicPIR(piicodev.QwiicPIRAddress, 1); err != nil {
			fmt.Println(err)
			return
		}
	}

	var pubTemp, pubPressure, pubHumidity, pubLightLevel float64

	var bme280temp, bme280pressure, bme280humidity float64
	var tmp117temp float64
	var ms5637pressure, ms5637temperature float64
	var veml6030light float64
	var vl53l1xrange uint16
	var pirMovement int

	for {
		if bme280 != nil {
			if t, p, h, err := bme280.Read(); err != nil {
				fmt.Println("ERROR: atmospheric", err)
			} else {
				// REVISIT: apply weird offsets relative to other sensors
				bme280temp = t + 1.0
				bme280pressure = p - 1.4
				bme280humidity = h

				pubTemp = bme280temp
				pubPressure = bme280pressure
				pubHumidity = bme280humidity
			}
		}

		if tmp117 != nil {
			if t, err := tmp117.ReadTempC(); err != nil {
				fmt.Println("ERROR: temperature", err)
			} else {
				tmp117temp = t
				pubTemp = tmp117temp
			}
		}

		if ms5637 != nil {
			if p, t, err := ms5637.Read(); err != nil {
				fmt.Println("ERROR: pressure", err)
			} else {
				ms5637pressure = p
				ms5637temperature = t
				pubPressure = ms5637pressure
			}
		}

		if veml6030 != nil {
			if l, err := veml6030.Read(); err != nil {
				fmt.Println("ERROR: light level", err)
			} else {
				veml6030light = l
				pubLightLevel = veml6030light
			}
		}

		if vl53l1x != nil {
			if r, err := vl53l1x.Read(); err != nil {
				fmt.Println("ERROR: distance", err)
			} else {
				vl53l1xrange = r
			}
		}

		var capStatus1, capStatus2, capStatus3 bool
		if cap1203 != nil {
			if capStatus1, capStatus2, capStatus3, err = cap1203.Read(); err != nil {
				fmt.Println("ERROR: capacitive touch", err)
			}
		}

		if pir != nil {
			/*
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
			*/

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

		prom.SetHumidity(pubHumidity)

		hkb.SetTemperature(pubTemp)
		hkb.SetLightLevel(pubLightLevel)
		if config.EnableVL53L1X {
			hkb.SetRangeSensor(vl53l1xrange)
		}
		if config.EnablePIR {
			hkb.SetMovement(pirMovement)
		}

		prom.SetTemperature(tmp117temp)
		prom.SetTemperatureBME(bme280temp)
		prom.SetTemperatureMS(ms5637temperature)
		prom.SetLightLevel(veml6030light)
		prom.SetRangeSensor(vl53l1xrange)
		prom.SetPressure(ms5637pressure)
		prom.SetPressureBME(bme280pressure)
		if oled != nil {
			// oled.WriteOLED(fmt.Sprintf("%.2f C\n%.2f hPa\n%.2f lux", tempC, pressure, light))
			if config.EnableVEML6030 {
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

				if capStatus1 {
					if bulbState.Power {
						fmt.Println("Turn off light")
						bulb.SetPowerState(false)
					} else {
						fmt.Println("Turn on light")
						bulb.SetPowerState(true)
					}
				}

				if capStatus2 {
					fmt.Println("Low light")
					bulb.SetColorState(&golifx.HSBK{
						Hue:        5461,
						Saturation: 0,
						Brightness: 47185,
						Kelvin:     2000,
					}, 0)
				}

				if capStatus3 {
					fmt.Println("Full light")
					bulb.SetColorState(&golifx.HSBK{
						Hue:        5461,
						Saturation: 0,
						Brightness: 65535,
						Kelvin:     4000,
					}, 0)
				}
			}
		}

		fmt.Printf("%.2f hPa, %.2f C | %.2f C | %.2f lux | %d mm | %.2f C, %.2f hPa, %.2f rH | %t,%t,%t | %d | bulb: %s\n",
			ms5637pressure, ms5637temperature, tmp117temp, veml6030light, vl53l1xrange, bme280temp, bme280pressure, bme280humidity,
			capStatus1, capStatus2, capStatus3, pirMovement, powerState)
		if config.SampleTime > 0 {
			time.Sleep(time.Duration(config.SampleTime) * time.Second)
		} else {
			time.Sleep(5 * time.Second)
		}
	}
}
