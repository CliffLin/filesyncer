package syncer

import (
	"filesyncer/internal/filemanager"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func createRootDir(local, remote string) {
	// Create root dir if there is no in remote.
	if _, err := os.Stat(remote); err != nil {
		if os.IsNotExist(err) {
			log.Print("Create remote dir")
			filemanager.DirManager{}.OnCreate(remote, local)
		}
	}

}

func FullSync(local, remote string) error {
	createRootDir(local, remote)

	// Remove remote file/dir if these are deleted from local.
	filepath.Walk(remote, func(path string, info fs.FileInfo, err error) error {
		relative := filemanager.RelativelyPath(remote, path)
		if _, err := os.Stat(fmt.Sprintf("%s%s", local, relative)); err != nil {
			if os.IsNotExist(err) {
				err := filemanager.DirManager{}.OnRemove(remote+relative, local+relative)
				if err != nil {
					log.Println(err)
					return err
				}
			}
		}
		return nil
	})

	filepath.Walk(local, func(path string, info fs.FileInfo, err error) error {
		manager, err := filemanager.GetManager(path)
		if err != nil {
			log.Println("cannot get manager, err", err)
			return err
		}

		relative := filemanager.RelativelyPath(local, path)
		remoteStat, err := os.Stat(fmt.Sprintf("%s%s", remote, relative))
		if err != nil && os.IsNotExist(err) {
			err := manager.OnCreate(remote+relative, path)
			if err != nil {
				fmt.Println(err)
			}
			return err
		} else {
			return err
		}

		if info.ModTime().After(remoteStat.ModTime()) {
			return manager.OnWrite(remote+relative, path)
		}
		return nil
	})
	return nil
}
