package cache

import (
	"fmt"
	"open-website-defender/internal/infrastructure/event"
	"open-website-defender/internal/infrastructure/logging"
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
	Store().Del(KeyWhiteListRules)
}

func clearBlackList(e event.Event, data any) {
	logging.Sugar.Debug("Cache invalidation: blacklist rules")
	Store().Del(KeyBlackListRules)
}

func clearGeoBlock(e event.Event, data any) {
	logging.Sugar.Debug("Cache invalidation: geoblock codes")
	Store().Del(KeyGeoBlockCodes)
}

func clearSystemSettings(e event.Event, data any) {
	logging.Sugar.Debug("Cache invalidation: system settings")
	Store().Del(KeySystemSettings)
}

func clearLicenseToken(e event.Event, data any) {
	if tokenHash, ok := data.(string); ok && tokenHash != "" {
		logging.Sugar.Debugf("Cache invalidation: license token %s...", tokenHash[:8])
		Store().Del(KeyLicenseToken + tokenHash)
	}
}

func clearUserInfo(e event.Event, data any) {
	if userID, ok := data.(uint); ok {
		logging.Sugar.Debugf("Cache invalidation: user %d", userID)
		Store().Del(fmt.Sprintf("%s%d", KeyUserInfo, userID))
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
