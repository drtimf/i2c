package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"strings"

	"github.com/golang/freetype/truetype"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/devices/v3/ssd1306/image1bit"
	"periph.io/x/host/v3"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

// OLED at 128x64 pixels
const OLEDDisplayDPI = 172.0

type OLEDFont struct {
	ttf *truetype.Font
}

func LoadFont(fontFile string) (font *OLEDFont, err error) {
	font = &OLEDFont{}

	var ttfFont []byte
	if ttfFont, err = ioutil.ReadFile(fontFile); err != nil {
		return
	}

	if font.ttf, err = truetype.Parse(ttfFont); err != nil {
		return
	}

	return
}

func (f *OLEDFont) GetFace(size float64) (face font.Face) {
	return truetype.NewFace(f.ttf, &truetype.Options{
		Size:    size,
		DPI:     OLEDDisplayDPI,
		Hinting: font.HintingNone,
	})
}

type OLEDDisplay struct {
	bus                 i2c.BusCloser
	dev                 *ssd1306.Dev
	font                *OLEDFont
	numberFace          font.Face
	symbolFace          font.Face
	extraFace           font.Face
	temperatureDisplayY int
}

func NewOLEDDisplay() (d *OLEDDisplay, err error) {
	d = &OLEDDisplay{}

	if _, err = host.Init(); err != nil {
		return
	}

	// Use i2creg I²C bus registry to find the first available I²C bus.
	if d.bus, err = i2creg.Open(""); err != nil {
		return
	}

	if d.dev, err = ssd1306.NewI2C(d.bus, &ssd1306.DefaultOpts); err != nil {
		return
	}

	if d.font, err = LoadFont("Orbitron-Medium.ttf"); err != nil {
		return
	}

	d.numberFace = d.font.GetFace(15.0)
	d.symbolFace = d.font.GetFace(8.0)
	d.extraFace = inconsolata.Bold8x16 // d.font.GetFace(5.0)

	d.temperatureDisplayY = d.numberFace.Metrics().Height.Ceil() - 9
	return
}

func (d *OLEDDisplay) DisplayTemperature(t float64, s string) (err error) {
	img := image1bit.NewVerticalLSB(d.dev.Bounds())

	temperatureDrawer := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: d.numberFace,
		Dot:  fixed.P(0, d.temperatureDisplayY),
	}

	temperatureDrawer.DrawString(fmt.Sprintf("%.1f", t))

	temperatureDrawer.Face = d.symbolFace
	temperatureDrawer.DrawString(" °c")

	strs := strings.Split(s, "\n")
	for i, ds := range strs {
		extraDrawer := font.Drawer{
			Dst:  img,
			Src:  &image.Uniform{image1bit.On},
			Face: d.extraFace,
			Dot:  fixed.P(0, d.temperatureDisplayY+(i+1)*(d.extraFace.Metrics().Height.Ceil()+2)),
		}
		extraDrawer.DrawString(ds)
	}

	return d.dev.Draw(d.dev.Bounds(), img, image.Point{})
}

/*
func (d *OLEDDisplay) WriteOLED(s string) {
	img := image1bit.NewVerticalLSB(d.dev.Bounds())
	f := inconsolata.Bold8x16
	strs := strings.Split(s, "\n")
	for i, ds := range strs {
		drawer := font.Drawer{
			Dst:  img,
			Src:  &image.Uniform{image1bit.On},
			Face: f,
			Dot:  fixed.P(0, (i+1) * (f.Height+2)),
		}
		drawer.DrawString(ds)
	}

	if err := d.dev.Draw(d.dev.Bounds(), img, image.Point{}); err != nil {
		log.Fatal(err)
	}
}
*/
