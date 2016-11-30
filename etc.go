package etc

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"runtime"

	"github.com/midbel/ini"
)

type Config struct {
	Name      string
	Files     []string
	Locations []string
}

var DefaultConfig *Config

func init() {
	switch runtime.GOOS {
	case "linux":
		DefaultConfig = &Config{
			Name:      os.Args[0],
			Files:     []string{os.Args[0]},
			Locations: []string{"/etc", "/usr/local/etc"},
		}
	}
}

func Configure(v interface{}, others ...string) error {
	paths := configPaths(DefaultConfig.Dirs(), DefaultConfig.Files, others...)
	return configure(v, paths)
}

func (c Config) Dirs() []string {
	if c.Name == "" {
		return c.Locations
	}
	paths := make([]string, 0, len(c.Locations))
	for _, l := range c.Locations {
		if filepath.Base(l) != c.Name {
			l = filepath.Join(l, c.Name)
		}
		paths = append(paths, l)
	}
	return paths
}

func (c Config) Configure(v interface{}) error {
	return configure(v, configPaths(c.Dirs(), c.Files))
}

func configure(v interface{}, paths []string) error {
	var err error
	for _, p := range paths {
		r, err := os.Open(p)
		if err != nil {
			continue
		}
		switch filepath.Ext(p) {
		case ".json":
			err = json.NewDecoder(r).Decode(v)
		case ".xml":
			err = xml.NewDecoder(r).Decode(v)
		case ".ini", "":
			err = ini.Read(r, v, "")
		}
		r.Close()
	}
	return err
}

func configPaths(dirs []string, files []string, others ...string) []string {
	for _, d := range others {
		if d == "" {
			continue
		}
		dir, base := filepath.Split(d)
		dirs = append(dirs, dir)
		if base == "" {
			continue
		}
		files = append(files, base)
	}
	paths := make([]string, 0, len(dirs)*len(files))
	seens := make(map[string]bool)

	for _, d := range dirs {
		for _, f := range files {
			p := filepath.Join(d, f)
			if _, ok := seens[p]; ok {
				continue
			}
			seens[p] = true
			paths = append(paths, p)
		}
	}
	return paths
}