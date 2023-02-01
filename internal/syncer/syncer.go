package syncer

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"

	"filesyncer/internal/filemanager"
)

type Syncer struct {
	RemotePath string
	LocalPath  string
}

func (s *Syncer) Run() {
	log.Println("Start ", s.LocalPath, s.RemotePath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
	}
	defer watcher.Close()

	s.watchAllDir(watcher, s.LocalPath)
	s.FullSync()
	s.onChange(watcher)
}

func (s *Syncer) watch(watcher *fsnotify.Watcher, path string) {
	if err := watcher.Add(path); err != nil {
		log.Println(err)
	}

	log.Printf("Start to watch: %s", path)
}

func (s *Syncer) watchAllDir(watcher *fsnotify.Watcher, root string) {
	filepath.Walk(
		s.LocalPath,
		func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				s.watch(watcher, path)
			}
			return nil
		},
	)
}

func (s *Syncer) onChange(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			log.Println("EVENT:", event)
			if !ok {
				log.Println("cannot get event")
				continue
			}
			err := filemanager.EventProcess(event, s.LocalPath, s.RemotePath, watcher)
			if err != nil {
				log.Println("Cannot Process, err", err)
			}
		case err, ok := <-watcher.Errors:
			log.Println("ERR:", err, ok)
		}
	}
}

func (s *Syncer) FullSync() error {
	// Create root dir if there is no in remote.
	if _, err := os.Stat(s.RemotePath); err != nil {
		if os.IsNotExist(err) {
			log.Print("Create remote dir")
			filemanager.DirManager{}.OnCreate(s.RemotePath, s.LocalPath)
		}
	}

	// Remove remote file/dir if these are deleted from local.
	filepath.Walk(s.RemotePath, func(path string, info fs.FileInfo, err error) error {
		relative := filemanager.RelativelyPath(s.RemotePath, path)
		if _, err := os.Stat(fmt.Sprintf("%s%s", s.LocalPath, relative)); err != nil {
			if os.IsNotExist(err) {
				err := filemanager.DirManager{}.OnRemove(s.RemotePath+relative, s.LocalPath+relative)
				if err != nil {
					log.Println(err)
					return err
				}
			}
		}
		return nil
	})

	filepath.Walk(s.LocalPath, func(path string, info fs.FileInfo, err error) error {
		manager, err := filemanager.GetManager(path)
		if err != nil {
			log.Println("cannot get manager, err", err)
			return err
		}

		relative := filemanager.RelativelyPath(s.LocalPath, path)
		remoteStat, err := os.Stat(fmt.Sprintf("%s%s", s.RemotePath, relative))
		if err != nil && os.IsNotExist(err) {
			err := manager.OnCreate(s.RemotePath+relative, path)
			if err != nil {
				fmt.Println(err)
			}
			return err
		} else {
			return err
		}

		if info.ModTime().After(remoteStat.ModTime()) {
			return manager.OnWrite(s.RemotePath+relative, path)
		}
		return nil
	})
	return nil
}
