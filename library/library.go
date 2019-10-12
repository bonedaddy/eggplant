package library

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/boreq/eggplant/logging"
	"github.com/pkg/errors"
)

type Id string

func (id Id) String() string {
	return string(id)
}

type track struct {
	name     string
	fileHash string
}

func newTrack(path string) (*track, error) {
	h, err := getFileHash(path)
	if err != nil {
		return nil, err
	}
	t := &track{
		fileHash: h,
	}
	return t, nil
}

type directory struct {
	name        string
	directories map[Id]*directory
	tracks      map[Id]*track
}

func newDirectory(name string) *directory {
	return &directory{
		name:        name,
		directories: make(map[Id]*directory),
		tracks:      make(map[Id]*track),
	}
}

type Library struct {
	directory string
	root      *directory
	log       logging.Logger
}

func Open(directory string) (*Library, error) {
	l := &Library{
		log:       logging.New("library"),
		root:      newDirectory("Music Library"),
		directory: directory,
	}

	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			l.log.Debug("file", "name", info.Name(), "path", path)
			if err := l.addFile(path); err != nil {
				return errors.Wrap(err, "could not add a file")
			}
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "walk failed")
	}
	return l, nil

}

func (l *Library) addFile(path string) error {
	relativePath, err := filepath.Rel(l.directory, path)
	if err != nil {
		return errors.Wrap(err, "could not get relative filepath")
	}

	dir, file := filepath.Split(relativePath)
	dirs := strings.Split(strings.Trim(dir, string(os.PathSeparator)), string(os.PathSeparator))

	directory, err := l.getOrCreateDirectory(dirs)
	if err != nil {
		return errors.Wrap(err, "could not get directory")
	}
	track, err := newTrack(path)
	if err != nil {
		return err
	}
	trackId, err := getHash(file)
	if err != nil {
		return err
	}
	directory.tracks[trackId] = track
	return nil
}

func (l *Library) getOrCreateDirectory(names []string) (*directory, error) {
	var dir *directory = l.root
	for _, name := range names {
		id, err := getHash(name)
		if err != nil {
			return nil, err
		}

		subdir, ok := dir.directories[id]
		if !ok {
			subdir = newDirectory(name)
			dir.directories[id] = subdir
		}
		dir = subdir
	}
	return dir, nil
}

func (l *Library) getDirectory(ids []Id) (*directory, error) {
	var dir *directory = l.root
	for _, id := range ids {
		subdir, ok := dir.directories[id]
		if !ok {
			return nil, errors.Errorf("subdirectory '%s' not found", id)
		}
		dir = subdir
	}
	return dir, nil
}

type Track struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Directory struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	Directories []Directory `json:"directories,omitempty"`
	Tracks      []Track     `json:"tracks,omitempty"`
}

func (l *Library) Browse(ids []Id) (Directory, error) {
	listed := Directory{}

	//if len(parts) > 0 {
	//	listed.Name = parts[len(parts)-1]
	//}

	dir, err := l.getDirectory(ids)
	if err != nil {
		return Directory{}, errors.Wrap(err, "failed to get directory")
	}

	for id, directory := range dir.directories {
		d := Directory{
			Id:   id.String(),
			Name: directory.name,
		}
		listed.Directories = append(listed.Directories, d)
	}

	for id, track := range dir.tracks {
		t := Track{
			Id:   id.String(),
			Name: track.name,
		}
		listed.Tracks = append(listed.Tracks, t)
	}

	return listed, nil
}

func getHash(s string) (Id, error) {
	buf := bytes.NewBuffer([]byte(s))
	hasher := sha256.New()
	if _, err := io.Copy(hasher, buf); err != nil {
		return "", err
	}
	var sum []byte
	sum = hasher.Sum(sum)
	return Id(hex.EncodeToString(sum)), nil
}

func getFileHash(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	var sum []byte
	sum = hasher.Sum(sum)
	return hex.EncodeToString(sum), nil
}