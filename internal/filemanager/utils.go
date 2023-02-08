package filemanager

import (
	"fmt"
	"io"
	"log"
	"os"
)

func Copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		log.Println("cannot get src stat")
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file.", src)
	}

	source, err := os.Open(src)
	if err != nil {
		log.Println("cannot get src")
		return err
	}
	defer source.Close()
	tmpDstPath := fmt.Sprintf("%s.tmp", dst)
	destination, err := os.Create(tmpDstPath)
	if err != nil {
		log.Println("cannot create dst")
		return err
	}

	defer func() error {
		destination.Close()
		return os.Rename(tmpDstPath, dst)
	}()

	bufferSize := 512
	buf := make([]byte, bufferSize)

	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("cannot read file")
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			log.Println("cannot write file")

			return err
		}
	}
	return nil
}

func CopyPerm(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	dstFileStat, err := os.Stat(dst)
	if err != nil {
		return err
	}

	if sourceFileStat.Mode().Perm() != dstFileStat.Mode().Perm() {
		return os.Chmod(dst, sourceFileStat.Mode())
	}
	return nil
}

func RelativelyPath(root, path string) string {
	return path[len(root):]
}

func GetManager(path string) (FileManagerInterface, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return DirManager{}, nil
	}
	return FileManager{}, nil
}
