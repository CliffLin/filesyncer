package syncer

import (
	"filesyncer/internal/filemanager"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type RealTimeSyncer struct {
	RemotePath  string
	LocalPath   string
	Config      *RealTimeSyncerConfig
	watcher     *fsnotify.Watcher
	mutex       sync.Mutex
	renameQueue []string
}

type RealTimeSyncerConfig struct{}

func (r *RealTimeSyncer) Run() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		return
	}
	defer watcher.Close()

	r.watcher = watcher

	r.watchAllDir()
	r.fullSync()

	r.listening()
}

func (r *RealTimeSyncer) getFileManager(event fsnotify.Event) (filemanager.FileManagerInterface, error) {
	switch event.Op {
	case fsnotify.Create, fsnotify.Write, fsnotify.Chmod:
		return filemanager.GetManager(event.Name)
	case fsnotify.Remove, fsnotify.Rename:
		return filemanager.GetManager(r.RemotePath)
	}
	return nil, fmt.Errorf("cannot get file manager")
}

func (r *RealTimeSyncer) ProcessEvent(event fsnotify.Event) error {
	relative := filemanager.RelativelyPath(r.LocalPath, event.Name)
	remotePath := fmt.Sprintf("%s%s", r.RemotePath, relative)

	manager, err := r.getFileManager(event)
	if err != nil {
		return err
	}

	log.Printf("%s|%s", manager.Type(), relative)

	switch event.Op {
	case fsnotify.Create:
		defer func() {
			if manager.Type() == "DIR" {
				r.watcher.Add(event.Name)
			}
		}()
		log.Println("Create")
		if len(r.renameQueue) == 0 {
			err := manager.OnCreate(remotePath, event.Name)
			return err
		}
		r.mutex.Lock()
		defer r.mutex.Unlock()

		err := manager.OnRename(r.renameQueue[len(r.renameQueue)-1], remotePath)
		r.renameQueue = r.renameQueue[:len(r.renameQueue)-1]

		return err

	case fsnotify.Write:
		log.Println("write")
		return manager.OnWrite(remotePath, event.Name)

	case fsnotify.Remove:
		log.Println("remove")

		return manager.OnRemove(remotePath, event.Name)

	case fsnotify.Rename:
		log.Println("rename")
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.renameQueue = append(r.renameQueue, remotePath)
		return nil

	case fsnotify.Chmod:
		log.Println("chmod")
		return manager.OnChmod(remotePath, event.Name)
	}
	return nil

}

func (r *RealTimeSyncer) listening() {
	for {
		select {
		case event, ok := <-r.watcher.Events:
			log.Println("EVENT:", event)
			if !ok {
				log.Println("cannot get event")
				continue
			}
			err := filemanager.EventProcess(event, r.LocalPath, r.RemotePath, r.watcher)
			if err != nil {
				log.Println("Cannot Process, err", err)
			}
		case err, ok := <-r.watcher.Errors:
			log.Println("ERR:", err, ok)
		}
	}

}

func (r *RealTimeSyncer) watch(path string) {
	if err := r.watcher.Add(path); err != nil {
		log.Println(err)
	}

	log.Printf("Start to watch: %s", path)
}

func (r *RealTimeSyncer) watchAllDir() {
	filepath.Walk(
		r.LocalPath,
		func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				r.watch(path)
			}
			return nil
		},
	)

}

func (r *RealTimeSyncer) fullSync() error {
	return FullSync(r.LocalPath, r.RemotePath)
}
