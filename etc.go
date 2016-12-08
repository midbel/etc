package etc

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/midbel/ini"
)

type Config struct {
	Default   string
	Name      string
	Files     []string
	Locations []string
}

var DefaultConfig *Config

func init() {
	prgname := strings.ToUpper(filepath.Base(os.Args[0]))
	dirname := os.Getenv(fmt.Sprintf("%s_DIRNAME", prgname))
	filename := os.Getenv(fmt.Sprintf("%s_FILENAME", prgname))
	config := os.Getenv(fmt.Sprintf("%s_CONFIG", prgname)) 
	
	var locations []string
	switch runtime.GOOS {
	case "linux":
		locations = append(locations, "/etc", "/usr/local/etc")
		if dirname != "" {
			locations = append(locations, dirname)
		}
		if filename == "" {
			filename = os.Args[0]
		}
		DefaultConfig = &Config{
			Default:   config,
			Name:      os.Args[0],
			Files:     []string{filename},
			Locations: locations,
		}
	}
}

func Configure(v interface{}) error {
	return DefaultConfig.Configure(v)
}

func (c Config) Configure(v interface{}) error {
	paths := make([]string, 0, 1 + len(c.Files) * len(c.Locations))
	if c.Default != "" {
		paths = append(paths, c.Default)
	}
	for _,l := range c.Locations {
		l = filepath.Join(l, c.Name)
		for _, f := range c.Files {
			paths = append(paths, filepath.Join(l, f))
		}
	}
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
			err = ini.NewReader(r).Read(v)
		}
		r.Close()
		if p == c.Default && err == nil {
			break
		}
	}
	return err
}


