package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
)

const BOMURL = "http://www.bom.gov.au/fwo/IDV60901/IDV60901.95867.json"

type BOMScanner struct {
	scheduler    *gocron.Scheduler
	bomUpdateJob *gocron.Job

	AirTemp   float64
	WindSpeed float64
	WindGust  float64
	WindDir   float64

	debugOutput bool
}

var bomScannerOnce sync.Once
var bomScanner *BOMScanner

type BOMQueryResult struct {
	Observations struct {
		Notice []struct {
			Copyright     string `json:"copyright"`
			CopyrightURL  string `json:"copyright_url"`
			DisclaimerURL string `json:"disclaimer_url"`
			FeedbackURL   string `json:"feedback_url"`
		} `json:"notice"`
		Header []struct {
			RefreshMessage string `json:"refresh_message"`
			ID             string `json:"ID"`
			MainID         string `json:"main_ID"`
			Name           string `json:"name"`
			StateTimeZone  string `json:"state_time_zone"`
			TimeZone       string `json:"time_zone"`
			ProductName    string `json:"product_name"`
			State          string `json:"state"`
		} `json:"header"`
		Data []struct {
			SortOrder         int     `json:"sort_order"`
			Wmo               int     `json:"wmo"`
			Name              string  `json:"name"`
			HistoryProduct    string  `json:"history_product"`
			LocalDateTime     string  `json:"local_date_time"`
			LocalDateTimeFull string  `json:"local_date_time_full"`
			AifstimeUtc       string  `json:"aifstime_utc"`
			Lat               float64 `json:"lat"`
			Lon               float64 `json:"lon"`
			ApparentT         float64 `json:"apparent_t"`
			Cloud             string  `json:"cloud"`
			CloudBaseM        string  `json:"cloud_base_m"`
			CloudOktas        string  `json:"cloud_oktas"`
			CloudTypeID       string  `json:"cloud_type_id"`
			CloudType         string  `json:"cloud_type"`
			DeltaT            float64 `json:"delta_t"`
			GustKmh           int     `json:"gust_kmh"`
			GustKt            int     `json:"gust_kt"`
			AirTemp           float64 `json:"air_temp"`
			Dewpt             float64 `json:"dewpt"`
			Press             string  `json:"press"`
			PressQnh          string  `json:"press_qnh"`
			PressMsl          string  `json:"press_msl"`
			PressTend         string  `json:"press_tend"`
			RainTrace         string  `json:"rain_trace"`
			RelHum            int     `json:"rel_hum"`
			SeaState          string  `json:"sea_state"`
			SwellDirWorded    string  `json:"swell_dir_worded"`
			SwellHeight       string  `json:"swell_height"`
			SwellPeriod       string  `json:"swell_period"`
			VisKm             string  `json:"vis_km"`
			Weather           string  `json:"weather"`
			WindDir           string  `json:"wind_dir"`
			WindSpdKmh        int     `json:"wind_spd_kmh"`
			WindSpdKt         int     `json:"wind_spd_kt"`
		} `json:"data"`
	} `json:"observations"`
}

func NewBOMScanner(debug bool) (bs *BOMScanner, err error) {
	bomScannerOnce.Do(func() {
		bomScanner = &BOMScanner{
			scheduler:   gocron.NewScheduler(time.Local),
			debugOutput: debug,
		}

		if bomScanner.bomUpdateJob, err = bomScanner.scheduler.Every("5m").SingletonMode().Tag("BOMUpdate").Do(bomUpdate); err != nil {
			return
		}

		bomScanner.scheduler.StartAsync()
	})

	bomUpdate()
	bs = bomScanner
	return
}

func convertDir(dir string) float64 {
	switch strings.ToUpper(dir) {
	case "CALM":
		return 0.0
	case "N":
		return 0.0
	case "NNE":
		return 22.5
	case "NE":
		return 45
	case "ENE":
		return 67.5
	case "E":
		return 90
	case "ESE":
		return 112.5
	case "SE":
		return 135
	case "SSE":
		return 157.5
	case "S":
		return 180
	case "SSW":
		return 202.5
	case "SW":
		return 225
	case "WSW":
		return 247.5
	case "W":
		return 270
	case "WNW":
		return 292.5
	case "NW":
		return 315
	case "NNW":
		return 337.5
	default:
		fmt.Println("Unknown direction:", dir)
		return 0.0
	}
}

func bomUpdate() {
	var err error
	httpClient := http.Client{
		Timeout: time.Second * 30,
	}

	var req *http.Request
	if req, err = http.NewRequest("GET", BOMURL, nil); err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("User-Agent", "Chrome/103.0.0.0 Safari/537.36")

	var resp *http.Response
	if resp, err = httpClient.Do(req); err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		fmt.Println(err)
		return
	}

	qr := new(BOMQueryResult)
	if err = json.Unmarshal(body, qr); err != nil {
		fmt.Println(err)
		return
	}

	if qr.Observations.Data != nil && len(qr.Observations.Data) > 0 {
		bomScanner.AirTemp = qr.Observations.Data[0].AirTemp
		bomScanner.WindSpeed = float64(qr.Observations.Data[0].WindSpdKmh)
		bomScanner.WindGust = float64(qr.Observations.Data[0].GustKmh)
		bomScanner.WindDir = convertDir(qr.Observations.Data[0].WindDir)
		_, t := bomScanner.scheduler.NextRun()
		if bomScanner.debugOutput {
			fmt.Printf("BOM: air temp = %f, wind = %f,%f, next: %v\n", bomScanner.AirTemp, bomScanner.WindSpeed, bomScanner.WindDir, t)
		}
	}
}
