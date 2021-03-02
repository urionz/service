package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urionz/goutil/fsutil"
	"github.com/urionz/goutil/strutil"
)

type Factory interface {
	Disk(name ...string) IFilesystem
}

type IFilesystem interface {
	Exists(path string) bool
	Get(path string) (string, error)
}

type Filesystem struct {
}

func (f *Filesystem) Exists(path string) bool {
	return fsutil.FileExists(path)
}

func (f *Filesystem) Missing(path string) bool {
	return !f.Exists(path)
}

func (f *Filesystem) Get(path string) ([]byte, error) {
	if fsutil.IsFile(path) {
		return ioutil.ReadFile(path)
	}
	return []byte{}, fmt.Errorf("file does not exist at path %s", path)
}

func (f *Filesystem) Hash(path string) string {
	return strutil.Md5File(path)
}

func (f *Filesystem) Put(p string, contents []byte) error {
	handle := fsutil.MustCreateFile(p, os.ModePerm, os.ModePerm)
	defer handle.Close()
	if _, err := handle.Write(contents); err != nil {
		return err
	}
	return nil
}
