package pkg

import (
	"net"
	"open-website-defender/internal/infrastructure/logging"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

var (
	geoReader *geoip2.Reader
	geoOnce   sync.Once
	geoErr    error
)

// InitGeoIP initializes the GeoIP reader from the given MMDB file path.
func InitGeoIP(dbPath string) error {
	geoOnce.Do(func() {
		geoReader, geoErr = geoip2.Open(dbPath)
		if geoErr != nil {
			logging.Sugar.Warnf("Failed to open GeoIP database: %v", geoErr)
		} else {
			logging.Sugar.Infof("GeoIP database loaded: %s", dbPath)
		}
	})
	return geoErr
}

// LookupCountry returns the ISO country code for an IP address.
// Returns empty string if GeoIP is not initialized or lookup fails.
func LookupCountry(ipStr string) string {
	if geoReader == nil {
		return ""
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}
	record, err := geoReader.Country(ip)
	if err != nil {
		return ""
	}
	return record.Country.IsoCode
}

// CloseGeoIP closes the GeoIP reader.
func CloseGeoIP() {
	if geoReader != nil {
		geoReader.Close()
	}
}
