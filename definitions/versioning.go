package definitions

import (
	"os"
	"time"

	"github.com/gopher-fleece/gleece/infrastructure/utils"
)

type FileVersion struct {
	Path    string    // The file's full path
	ModTime time.Time // The file's last modification time
	Hash    string    // Content hash, used when ModTime differs
}

func NewFileVersion(fullPath string) (FileVersion, error) {

	stats, err := os.Stat(fullPath)
	if err != nil {
		return FileVersion{}, err
	}

	hash, err := utils.Sha256File(fullPath)
	if err != nil {
		return FileVersion{}, err
	}

	return FileVersion{
		Path:    fullPath,
		ModTime: stats.ModTime(),
		Hash:    hash,
	}, nil
}

func (v *FileVersion) HasChanged(selfUpdate bool) (bool, error) {
	stats, err := os.Stat(v.Path)
	if err != nil {
		return false, err
	}

	if v.ModTime.Equal(stats.ModTime()) {
		return false, nil
	}

	hash, err := utils.Sha256File(v.Path)
	if err != nil {
		return true, err
	}

	hasChanged := v.Hash != hash

	if selfUpdate {
		v.Hash = hash
		v.ModTime = stats.ModTime()
	}

	return hasChanged, nil
}
