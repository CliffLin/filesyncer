package syncer

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

func GetSyncer(syncer_type string, local, remote string, config interface{}) (SyncerInterface, error) {
	switch syncer_type {
	case "RealTime":
		var realTimeSyncerConfig RealTimeSyncerConfig
		mapstructure.Decode(config, &realTimeSyncerConfig)
		return &RealTimeSyncer{
			RemotePath: remote,
			LocalPath:  local,
			Config:     &realTimeSyncerConfig,
		}, nil
	case "Periodic":
		var periodicSyncerConfig PeriodicSyncerConfig
		mapstructure.Decode(config, &periodicSyncerConfig)
		return &PeriodicSyncer{remote, local, periodicSyncerConfig}, nil
	}
	return nil, fmt.Errorf("Cannot get syncer, type: %s", syncer_type)
}
