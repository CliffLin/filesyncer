package filemanager

type FileManagerInterface interface {
	OnCreate(string, string) error
	OnWrite(string, string) error
	OnRemove(string, string) error
	OnRename(string, string) error
	OnChmod(string, string) error
	Type() string
}
