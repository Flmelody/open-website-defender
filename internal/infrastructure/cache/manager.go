package cache

import (
	"fmt"
	"open-website-defender/internal/infrastructure/event"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/pkg"
)

// cache key constants — single source of truth for all cache keys
const (
	KeyWhiteListRules = "whitelist:rules"
	KeyBlackListRules = "blacklist:rules"
	KeyGeoBlockCodes  = "geoblock:codes"
	KeySystemSettings = "system:settings"
	KeyLicenseToken   = "license:token:"
	KeyUserInfo       = "user:info:"
)

func clearWhiteList(e event.Event, data any) {
	logging.Sugar.Debug("Cache invalidation: whitelist rules")
	pkg.Cacher().Del([]byte(KeyWhiteListRules))
}

func clearBlackList(e event.Event, data any) {
	logging.Sugar.Debug("Cache invalidation: blacklist rules")
	pkg.Cacher().Del([]byte(KeyBlackListRules))
}

func clearGeoBlock(e event.Event, data any) {
	logging.Sugar.Debug("Cache invalidation: geoblock codes")
	pkg.Cacher().Del([]byte(KeyGeoBlockCodes))
}

func clearSystemSettings(e event.Event, data any) {
	logging.Sugar.Debug("Cache invalidation: system settings")
	pkg.Cacher().Del([]byte(KeySystemSettings))
}

func clearLicenseToken(e event.Event, data any) {
	if tokenHash, ok := data.(string); ok && tokenHash != "" {
		logging.Sugar.Debugf("Cache invalidation: license token %s...", tokenHash[:8])
		pkg.Cacher().Del([]byte(KeyLicenseToken + tokenHash))
	}
}

func clearUserInfo(e event.Event, data any) {
	if userID, ok := data.(uint); ok {
		logging.Sugar.Debugf("Cache invalidation: user %d", userID)
		pkg.Cacher().Del([]byte(fmt.Sprintf("%s%d", KeyUserInfo, userID)))
	}
}

// Init registers all cache invalidation handlers on the event bus.
// Call this once during application startup.
func Init() {
	b := event.Bus()
	b.Subscribe(event.WhiteListChanged, clearWhiteList)
	b.Subscribe(event.BlackListChanged, clearBlackList)
	b.Subscribe(event.GeoBlockChanged, clearGeoBlock)
	b.Subscribe(event.SystemSettingsChanged, clearSystemSettings)
	b.Subscribe(event.LicenseChanged, clearLicenseToken)
	b.Subscribe(event.UserChanged, clearUserInfo)
}
