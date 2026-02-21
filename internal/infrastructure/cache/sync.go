package cache

import (
	"open-website-defender/internal/domain/entity"
	"open-website-defender/internal/infrastructure/event"
	"open-website-defender/internal/infrastructure/logging"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// versionKeyToEvent maps database version keys to event bus events.
var versionKeyToEvent = map[string]event.Event{
	"blacklist": event.BlackListChanged,
	"whitelist": event.WhiteListChanged,
	"geoblock":  event.GeoBlockChanged,
	"system":    event.SystemSettingsChanged,
	"waf_rules": event.WafRulesChanged,
}

// eventToVersionKey maps event bus events to database version keys.
var eventToVersionKey = map[event.Event]string{
	event.BlackListChanged:      "blacklist",
	event.WhiteListChanged:      "whitelist",
	event.GeoBlockChanged:       "geoblock",
	event.SystemSettingsChanged: "system",
	event.WafRulesChanged:       "waf_rules",
}

// CacheSyncService polls the database for version changes and publishes
// events to invalidate local caches across multiple instances.
type CacheSyncService struct {
	db       *gorm.DB
	versions map[string]int64
	mu       sync.Mutex
	stopCh   chan struct{}
	interval time.Duration
	polling  atomic.Bool // true while poll() is publishing events, prevents re-entrant BumpVersion
}

var (
	syncService *CacheSyncService
	syncMu      sync.Mutex
	syncDB      *gorm.DB // stored for RestartSync
)

// InitSync creates and starts the cache sync service.
// intervalSeconds controls the polling frequency; 0 disables syncing.
func InitSync(db *gorm.DB, intervalSeconds int) {
	syncMu.Lock()
	defer syncMu.Unlock()

	syncDB = db // store for RestartSync

	if intervalSeconds <= 0 {
		logging.Sugar.Info("Cache sync disabled (sync-interval=0)")
		return
	}

	svc := &CacheSyncService{
		db:       db,
		versions: make(map[string]int64),
		stopCh:   make(chan struct{}),
		interval: time.Duration(intervalSeconds) * time.Second,
	}

	// Load initial versions so we don't trigger spurious invalidations on startup.
	svc.loadVersions()

	// Subscribe to local events so writes on this instance bump the DB version.
	b := event.Bus()
	for evt, key := range eventToVersionKey {
		evt, key := evt, key // capture
		b.Subscribe(evt, func(_ event.Event, _ any) {
			// Skip if this event was fired by the poller (remote change detection).
			// Only local service writes should bump the DB version.
			if svc.polling.Load() {
				return
			}
			svc.BumpVersion(key)
		})
	}

	svc.start()
	syncService = svc
	logging.Sugar.Infof("Cache sync started (interval=%ds)", intervalSeconds)
}

// StopSync gracefully stops the cache sync service.
func StopSync() {
	syncMu.Lock()
	defer syncMu.Unlock()

	if syncService != nil {
		syncService.stop()
		syncService = nil
	}
}

// RestartSync restarts the cache sync service with a new interval.
// If intervalSeconds is 0, the sync service is stopped.
func RestartSync(intervalSeconds int) {
	syncMu.Lock()
	defer syncMu.Unlock()

	// Stop existing service if running
	if syncService != nil {
		// If same interval, do nothing
		if syncService.interval == time.Duration(intervalSeconds)*time.Second {
			return
		}
		syncService.stop()
		syncService = nil
	}

	if intervalSeconds <= 0 {
		logging.Sugar.Info("Cache sync disabled via settings update")
		return
	}

	svc := &CacheSyncService{
		db:       syncDB,
		versions: make(map[string]int64),
		stopCh:   make(chan struct{}),
		interval: time.Duration(intervalSeconds) * time.Second,
	}

	svc.loadVersions()

	b := event.Bus()
	for evt, key := range eventToVersionKey {
		evt, key := evt, key
		b.Subscribe(evt, func(_ event.Event, _ any) {
			if svc.polling.Load() {
				return
			}
			svc.BumpVersion(key)
		})
	}

	svc.start()
	syncService = svc
	logging.Sugar.Infof("Cache sync restarted (interval=%ds)", intervalSeconds)
}

func (s *CacheSyncService) start() {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.poll()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *CacheSyncService) stop() {
	close(s.stopCh)
}

// BumpVersion increments the version for the given key in the database
// and updates the local snapshot to avoid self-triggering on the next poll.
func (s *CacheSyncService) BumpVersion(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{"version": gorm.Expr("version + 1"), "updated_at": time.Now().UTC()}),
	}).Create(&entity.CacheVersion{
		Key:       key,
		Version:   1,
		UpdatedAt: time.Now().UTC(),
	})

	if result.Error != nil {
		logging.Sugar.Warnf("Cache sync: failed to bump version for %s: %v", key, result.Error)
		return
	}

	// Read back the current version so our local snapshot stays in sync.
	var cv entity.CacheVersion
	if err := s.db.Where("`key` = ?", key).First(&cv).Error; err == nil {
		s.versions[key] = cv.Version
	}
}

// loadVersions reads all current versions from the database into the local snapshot.
func (s *CacheSyncService) loadVersions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var rows []entity.CacheVersion
	if err := s.db.Find(&rows).Error; err != nil {
		logging.Sugar.Warnf("Cache sync: failed to load initial versions: %v", err)
		return
	}

	for _, row := range rows {
		s.versions[row.Key] = row.Version
	}
}

// poll checks the database for version changes and publishes events for any
// keys whose remote version is newer than our local snapshot.
func (s *CacheSyncService) poll() {
	var rows []entity.CacheVersion
	if err := s.db.Find(&rows).Error; err != nil {
		logging.Sugar.Warnf("Cache sync: poll failed: %v", err)
		return
	}

	// Determine which keys changed under the lock, then publish events outside
	// to avoid deadlock (Publish calls handlers synchronously, which may call BumpVersion).
	type change struct {
		evt    event.Event
		key    string
		oldVer int64
		newVer int64
	}
	var changes []change

	s.mu.Lock()
	for _, row := range rows {
		localVer, exists := s.versions[row.Key]
		if !exists || row.Version > localVer {
			s.versions[row.Key] = row.Version
			if !exists {
				// First time seeing this key — no need to invalidate.
				continue
			}
			if evt, ok := versionKeyToEvent[row.Key]; ok {
				changes = append(changes, change{evt: evt, key: row.Key, oldVer: localVer, newVer: row.Version})
			}
		}
	}
	s.mu.Unlock()

	if len(changes) == 0 {
		return
	}

	// Mark as polling so event handlers skip BumpVersion for these events.
	s.polling.Store(true)
	defer s.polling.Store(false)

	b := event.Bus()
	for _, c := range changes {
		logging.Sugar.Infof("Cache sync: %s version changed %d → %d, invalidating", c.key, c.oldVer, c.newVer)
		b.Publish(c.evt)
	}
}
