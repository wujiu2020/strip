package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/teapots/config"
	"github.com/teapots/teapot"
)

type SliceValue []string

func (s *SliceValue) Set(v string) error {
	for _, p := range strings.Split(v, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			*s = append(*s, v)
		}
	}
	return nil
}

func (s *SliceValue) String() string {
	return strings.Join(*s, ",")
}

func (s *SliceValue) Unique() {
	v := make(SliceValue, 0, len(*s))
	m := make(map[string]bool, len(*s))
	for _, p := range *s {
		if m[p] {
			continue
		}
		m[p] = true
		v = append(v, p)
	}
	*s = v
}

func LoadClassicEnv(tea *teapot.Teapot, app string, env interface{}, dir string, addFiles ...string) {
	LoadConfigFiles(tea, app, env, dir, addFiles...)
}

func LoadConfigFiles(tea *teapot.Teapot, app string, env interface{}, dir string, addFiles ...string) (conf config.Configer) {
	directory, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	wd, _ := os.Getwd()
	wd, _ = filepath.Abs(wd)

	directories := []string{directory, wd}
	if dir != "" {
		dir, _ = filepath.Abs(dir)
		directories = append(directories, dir)
	}

	files := []string{}
	for _, d := range directories {
		files = append(files, []string{
			filepath.Join(d, "default.ini"),
			filepath.Join(d, fmt.Sprintf("default.%s.ini", tea.Config.RunMode)),
		}...)
	}
	for _, d := range directories {
		files = append(files, []string{
			filepath.Join(d, app+".ini"),
			filepath.Join(d, app+fmt.Sprintf(".%s.ini", tea.Config.RunMode)),
		}...)
	}
	for _, d := range directories {
		files = append(files, []string{
			filepath.Join(d, "override.ini"),
			filepath.Join(d, fmt.Sprintf("override.%s.ini", tea.Config.RunMode)),
		}...)
	}
	skipNum := len(files)
	files = append(files, addFiles...)

	maps := make(map[string]bool, len(files))
	var last config.Configer
	for n, path := range files {
		if n < skipNum && maps[path] {
			continue
		}
		maps[path] = true
		conf, err := config.LoadIniFile(path)
		if err != nil {
			if !os.IsNotExist(err) || n >= skipNum {
				tea.Logger().Errorf("%s load err: %v", path, err)
			}
		} else {
			if last != nil {
				conf.SetParent(last)
			}
			last = conf
			tea.Logger().Infof("%s load success", path)
		}
	}

	if last == nil {
		last = config.Global
	}

	conf = last
	tea.ImportConfig(last)
	config.Decode(tea.Config, env)
	return
}

func UseGlobalLogger(tea *teapot.Teapot) {
	X.SetFlatLine(tea.Config.RunMode.IsProd())
	X.SetColorMode(tea.Config.RunMode.IsDev())
	X.SetShortLine(tea.Config.RunMode.IsProd())
	tea.SetLogger(X)
	tea.ProvideAs(X, (*teapot.ReqLogger)(nil))
}
