package server

import (
	"errors"
	"file-transfer/internal/log"
	"github.com/shirou/gopsutil/v3/disk"
)

type mount struct {
	path      string
	totalSize int64
	freeSize  int64
	allocSize int64
	nonce     int
}

type storage struct {
	mounts []*mount
	index  int
}

func newStorage() *storage {
	return &storage{
		mounts: make([]*mount, 0),
		index:  -1,
	}
}

func newMount(path string) (*mount, error) {
	stat, err := disk.Usage(path)
	if err != nil {
		return nil, err
	}

	return &mount{
		path:      path,
		totalSize: int64(stat.Total),
		freeSize:  int64(stat.Free),
		nonce:     0,
	}, nil
}

func (s *storage) resize() {
	for _, m := range s.mounts {
		stat, err := disk.Usage(m.path)
		if err != nil {
			log.Log.Errorf("Disk status error: %v", stat)
			continue
		}
		m.freeSize = int64(stat.Free)
	}
}

func (s *storage) addPath(path string) error {
	m, err := newMount(path)
	if err != nil {
		return err
	}
	s.mounts = append(s.mounts, m)
	return nil
}

func (s *storage) getPath(reqSize int64) (string, error) {
	ms := len(s.mounts)
	if ms == 0 {
		return "", errors.New("no available storage space")
	}

	var path string

	for i := 0; i < ms; i++ {
		if s.index >= ms-1 {
			s.index = -1
		}
		s.index = s.index + 1

		m := s.mounts[s.index]

		if m.freeSize > reqSize {
			m.freeSize = m.freeSize - reqSize
			path = m.path
			log.Log.Infof("%s free space remaining %d", path, m.freeSize)
			break
		}
	}

	if path == "" {
		return "", errors.New("unable to find available storage")
	}

	return path, nil
}

func (s *storage) release(path string, requestSize int64) {
	for _, m := range s.mounts {
		if m.path == path {
			m.freeSize = m.freeSize + requestSize
		}
	}
}
