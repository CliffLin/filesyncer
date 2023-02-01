package filemanager

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

var (
	rename_queue = []string{}
)

func EventProcess(event fsnotify.Event, localRoot, remoteRoot string, watcher *fsnotify.Watcher) error {
	relative := RelativelyPath(localRoot, event.Name)
	remotePath := fmt.Sprintf("%s%s", remoteRoot, relative)

	var manager filemanagerInterface
	switch event.Op {
	case fsnotify.Create, fsnotify.Write, fsnotify.Chmod:
		var err error
		manager, err = GetManager(event.Name)
		if err != nil {
			return err
		}
	case fsnotify.Remove, fsnotify.Rename:
		var err error
		manager, err = GetManager(remotePath)
		if err != nil {
			return err
		}
	}
	log.Printf("%s|%s", manager.Type(), relative)

	switch event.Op {
	case fsnotify.Create:
		defer func() {
			if manager.Type() == "DIR" {
				watcher.Add(event.Name)
			}
		}()
		log.Println("Create")
		if len(rename_queue) == 0 {
			err := manager.OnCreate(remotePath, event.Name)
			return err
		}

		err := manager.OnRename(rename_queue[len(rename_queue)-1], remotePath)
		rename_queue = rename_queue[:len(rename_queue)-1]
		return err

	case fsnotify.Write:
		log.Println("write")
		return manager.OnWrite(remotePath, event.Name)

	case fsnotify.Remove:
		log.Println("remove")

		return manager.OnRemove(remotePath, event.Name)

	case fsnotify.Rename:
		log.Println("rename")
		rename_queue = append(rename_queue, remotePath)
		return nil

	case fsnotify.Chmod:
		log.Println("chmod")
		return manager.OnChmod(remotePath, event.Name)
	}
	return nil
}
