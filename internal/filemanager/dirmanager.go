package filemanager

import "os"

type DirManager struct{}

func (f DirManager) OnCreate(remote, local string) error {
	localInfo, err := os.Stat(local)
	if err != nil {
		return err
	}

	return os.Mkdir(remote, localInfo.Mode().Perm())
}

func (f DirManager) OnWrite(remote, local string) error {
	return Copy(local, remote)
}

func (f DirManager) OnRemove(remote, local string) error {
	return os.RemoveAll(remote)
}

func (f DirManager) OnRename(old_path, new_path string) error {
	return os.Rename(old_path, new_path)
}

func (f DirManager) OnChmod(remote, local string) error {
	return CopyPerm(local, remote)
}

func (f DirManager) Type() string {
	return "DIR"
}
