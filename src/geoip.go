package main

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gobuffalo/packr"
	"github.com/oschwald/geoip2-golang"
)

var (
	geoipBytes []byte
	geoip      *geoip2.Reader
)

func initGeoIP() {
	var err error
	box := packr.NewBox("../ui")
	geoipBytes, err = box.Find("site/GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatalf("Failed to load geoip database file")
		return
	}
	geoip, err = geoip2.FromBytes(geoipBytes)
	if err != nil {
		log.Fatalf("Failed to load geoip database")
		return
	}
	//captchaKey = h
}

func getCountry(r *http.Request) string {
	userIP := ""
	//log.Printf("Headers %v", r.Header)
	if len(r.Header.Get("CF-Connecting-IP")) > 1 {
		userIP = strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))
	}
	if len(r.Header.Get("X-Forwarded-For")) > 1 && len(userIP) == 0 {
		userIP = strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	}
	if len(r.Header.Get("X-Real-IP")) > 1 && len(userIP) == 0 {
		userIP = strings.TrimSpace(r.Header.Get("X-Real-IP"))
	}
	if len(userIP) == 0 {
		userIP = r.RemoteAddr
	}
	if strings.Contains(userIP, ",") {
		userIP = strings.Split(userIP, ",")[0]
	}
	if strings.Contains(userIP, " ") {
		userIP = strings.Split(userIP, " ")[0]
	}
	if len(userIP) == 0 {
		return ""
	}
	ip := net.ParseIP(userIP)
	if ip == nil {
		if strings.Count(userIP, ":") == 1 {
			userIP = strings.Split(userIP, ":")[0]
			ip = net.ParseIP(userIP)
		}
		if ip == nil {
			log.Printf("Failed to parse userIP: %s", userIP)
			return ""
		}
	}
	record, err := geoip.Country(ip)
	if err != nil {
		log.Printf("Failed to detect country using IP address: %s", userIP)
		return ""
	}
	return record.Country.IsoCode
}
