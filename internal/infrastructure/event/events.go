package event

const (
	WhiteListChanged      Event = "whitelist:changed"
	BlackListChanged      Event = "blacklist:changed"
	GeoBlockChanged       Event = "geoblock:changed"
	SystemSettingsChanged Event = "system:changed"
	LicenseChanged        Event = "license:changed"
	UserChanged           Event = "user:changed"
)
