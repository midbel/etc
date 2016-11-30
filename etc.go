package etc

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"runtime"

	"github.com/midbel/ini"
)

var (
	XDGDataDir = os.Getenv("XDG_DATA_DIRS")
	XDGDataHome = os.Getenv("XDG_DATA_HOME")
	XDGConfigHome = os.Getenv("XDG_CONFIG_HOME")
	XDGConfigDirs = os.Getenv("XDG_CONFIG_DIRS")
)

type Config struct {
	Name      string
	Files     []string
	Locations []string
}

var DefaultConfig *Config

func init() {
	prgname := os.Args[0]
	switch runtime.GOOS {
	case "linux":
		p := fmt.Sprintf("%s_DIRNAME", strings.ToUpper(prgname))
		DefaultConfig = &Config{
			Name:      prgname,
			Files:     []string{prgname},
			Locations: []string{"/etc", "/usr/local/etc", XDGConfigHome, os.Getenv(p)},
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
			err = nil
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
