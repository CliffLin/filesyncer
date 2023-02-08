package syncer

import (
	"log"
	"time"
)

type PeriodicSyncer struct {
	RemotePath string
	LocalPath  string
	Config     PeriodicSyncerConfig
}

type PeriodicSyncerConfig struct {
	Interval int64 `mapstructure:"interval"`
}

func (p *PeriodicSyncer) Run() {
	err := p.fullSync()
	if err != nil {
		log.Println(err)
	}

	ticker := time.NewTicker(time.Duration(p.Config.Interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			err := p.fullSync()
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (p *PeriodicSyncer) fullSync() error {
	return FullSync(p.LocalPath, p.RemotePath)
}
