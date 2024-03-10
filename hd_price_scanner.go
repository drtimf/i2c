package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"golang.org/x/net/html"
)

type hdTypeCapacity int

const (
	external14TB hdTypeCapacity = iota
	external16TB
	external18TB
	external20TB
	external22TB
	red14TB
	red16TB
	red18TB
	red20TB
	red22TB
	hdTypeCapacitySize
)

type node html.Node

func (n *node) getAttribute(key string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}

	return "", false
}

func (n *node) checkId(id string) bool {
	if n.Type == html.ElementNode {
		s, ok := n.getAttribute("id")

		if ok && s == id {
			return true
		}
	}

	return false
}

func (n *node) getElementById(id string) *node {
	var traverse func(n *node, id string) *node
	traverse = func(n *node, id string) *node {
		if n.checkId(id) {
			return n
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			res := traverse((*node)(c), id)
			if res != nil {
				return res
			}
		}

		return nil
	}

	return traverse(n, id)
}

func (n *node) getElementsByName(name string) (nodes []*node) {
	var traverse func(n *node, name string, nodes *[]*node)
	traverse = func(n *node, name string, nodes *[]*node) {
		if n.Data == name {
			*nodes = append(*nodes, n)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse((*node)(c), name, nodes)
		}
	}

	nodes = make([]*node, 0)
	traverse(n, name, &nodes)
	return
}

func (n *node) Render(w io.Writer) error {
	return html.Render(w, (*html.Node)(n))
}

func ParseHTML(r io.Reader) (n *node, err error) {
	var doc *html.Node
	if doc, err = html.Parse(r); err != nil {
		return
	}

	return (*node)(doc), nil
}

type DiskInfo struct {
	PricePerGB float64
	PricePerTB float64
	Price      float64
	Capacity   string
	FormFactor string
	Name       string
}

func queryDisks(diskTypes string) (doc *node, err error) {
	httpClient := http.Client{
		Timeout: time.Second * 30,
	}

	query := "https://diskprices.com/?locale=au&condition=new&disk_types="+diskTypes
	var req *http.Request
	if req, err = http.NewRequest("GET", query, nil); err != nil {
		err = fmt.Errorf("failed to create new request for \"%s\": %v", query, err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")

	var resp *http.Response
	if resp, err = httpClient.Do(req); err != nil {
		err = fmt.Errorf("failed to query \"%s\": %v", query, err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("failed to query \"%s\", the GET request failed with response %d", query, resp.StatusCode)
		return
	}

	if doc, err = ParseHTML(resp.Body); err != nil {
		return
	}

	return
}

func GetDistInfo(diskTypes string) (disks []DiskInfo, err error) {
	var doc *node
	if doc, err = queryDisks(diskTypes); err != nil {
		return
	}

	var tbl *node
	if tbl = doc.getElementById("diskprices"); tbl == nil {
		err = fmt.Errorf("table with Id = diskprices was not found")
		return
	}

	var tbody []*node
	if tbody = tbl.getElementsByName("tbody"); len(tbody) < 1 {
		err = fmt.Errorf("table with Id = diskprices does not have a tbody")
	}

	disks = make([]DiskInfo, 0)

	for _, row := range tbody[0].getElementsByName("tr") {
		var cols []*node
		if cols = row.getElementsByName("td"); len(cols) > 8 {
			pricePerGB, _ := strconv.ParseFloat(strings.ReplaceAll(cols[0].FirstChild.Data[2:], ",", ""), 64)
			pricePerTB, _ := strconv.ParseFloat(strings.ReplaceAll(cols[1].FirstChild.Data[2:], ",", ""), 64)
			price, _ := strconv.ParseFloat(strings.ReplaceAll(cols[2].FirstChild.Data[2:], ",", ""), 64)

			disks = append(disks, DiskInfo{
				PricePerGB: pricePerGB,
				PricePerTB: pricePerTB,
				Price:      price,
				Capacity:   cols[3].FirstChild.Data,
				FormFactor: cols[5].FirstChild.Data,
				Name:       cols[8].FirstChild.FirstChild.Data,
			})
		}
	}

	return
}

type WesternDigitalDiskPrices struct {
	pricesPerTB [10]float64
}

func GetWesternDigitalDiskPrices() (wd *WesternDigitalDiskPrices, err error) {
	var disks []DiskInfo
	if disks, err = GetDistInfo("external_hdd"); err != nil {
		return
	}

	wd = &WesternDigitalDiskPrices{}

	for _, disk := range disks {
		name := strings.ToLower(disk.Name)

		if strings.Contains(disk.FormFactor, "External") &&
			(strings.Contains(name, "wd") || strings.Contains(name, "western digital")) {
			// fmt.Printf("%s - %s\n", disk.Capacity, disk.Name)
			var capacityIndex hdTypeCapacity = -1
			switch disk.Capacity {
			case "14 TB":
				capacityIndex = external14TB
			case "16 TB":
				capacityIndex = external16TB
			case "18 TB":
				capacityIndex = external18TB
			case "20 TB":
				capacityIndex = external20TB
			case "22 TB":
				capacityIndex = external22TB
			}

			if capacityIndex >= 0 {
				if wd.pricesPerTB[capacityIndex] == 0 || disk.PricePerTB < wd.pricesPerTB[capacityIndex] {
					wd.pricesPerTB[capacityIndex] = disk.PricePerTB
				}
			}
		}
	}

	if disks, err = GetDistInfo("internal_hdd"); err != nil {
		return
	}

	for _, disk := range disks {
		name := strings.ToLower(disk.Name)

		if strings.Contains(disk.FormFactor, "Internal") &&
			(strings.Contains(name, "wd") || strings.Contains(name, "western digital")) &&
			strings.Contains(name, "red") {
			// fmt.Printf("%s - %s\n", disk.Capacity, disk.Name)
			var capacityIndex hdTypeCapacity = -1
			switch disk.Capacity {
			case "14 TB":
				capacityIndex = red14TB
			case "16 TB":
				capacityIndex = red16TB
			case "18 TB":
				capacityIndex = red18TB
			case "20 TB":
				capacityIndex = red20TB
			case "22 TB":
				capacityIndex = red22TB
			}

			if capacityIndex >= 0 {
				if wd.pricesPerTB[capacityIndex] == 0 || disk.PricePerTB < wd.pricesPerTB[capacityIndex] {
					wd.pricesPerTB[capacityIndex] = disk.PricePerTB
				}
			}
		}
	}

	return
}

func (wd *WesternDigitalDiskPrices) pricePerTB(capacityIndex hdTypeCapacity) float64 {
	return wd.pricesPerTB[capacityIndex]
}

type HDPriceScanner struct {
	scheduler    *gocron.Scheduler
	bomUpdateJob *gocron.Job
	wd           *WesternDigitalDiskPrices
}

var hdPriceScannerOnce sync.Once
var hdPriceScanner *HDPriceScanner

func NewHDPriceScanner() (hdp *HDPriceScanner, err error) {
	hdPriceScannerOnce.Do(func() {
		hdPriceScanner = new(HDPriceScanner)
		hdPriceScanner.scheduler = gocron.NewScheduler(time.Local)

		if hdPriceScanner.bomUpdateJob, err = hdPriceScanner.scheduler.Every("2h").SingletonMode().Tag("HDPriceUpdate").Do(cronHDPriceUpdate); err != nil {
			return
		}

		hdPriceScanner.scheduler.StartAsync()
	})

	cronHDPriceUpdate()
	hdp = hdPriceScanner
	return
}

func cronHDPriceUpdate() {
	var err error
	var wd *WesternDigitalDiskPrices
	if wd, err = GetWesternDigitalDiskPrices(); err != nil {
		fmt.Printf("Failed to get hard disk price data: %v\n", err)
		return
	}

	hdPriceScanner.wd = wd
}
