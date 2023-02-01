package filemanager

import (
	"fmt"
	"os"
)

type FileManager struct{}

func (f FileManager) OnCreate(remote, local string) error {
	return Copy(local, remote)
}

func (f FileManager) OnWrite(remote, local string) error {
	fmt.Println("gg", local, remote)
	return Copy(local, remote)
}

func (f FileManager) OnRemove(remote, local string) error {
	return os.Remove(remote)
}

func (f FileManager) OnRename(old_path, new_path string) error {
	return os.Rename(old_path, new_path)
}

func (f FileManager) OnChmod(remote, local string) error {
	return CopyPerm(local, remote)
}

func (f FileManager) Type() string {
	return "FILE"
}
