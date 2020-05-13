package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wujiu2020/strip"
	"github.com/wujiu2020/strip/config"
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

func LoadClassicEnv(sp *strip.Strip, app string, env interface{}, dir string, addFiles ...string) {
	LoadConfigFiles(sp, app, env, dir, addFiles...)
}

func LoadConfigFiles(sp *strip.Strip, app string, env interface{}, dir string, addFiles ...string) (conf config.Configer) {
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
			filepath.Join(d, fmt.Sprintf("default.%s.ini", sp.Config.RunMode)),
		}...)
	}
	for _, d := range directories {
		files = append(files, []string{
			filepath.Join(d, app+".ini"),
			filepath.Join(d, app+fmt.Sprintf(".%s.ini", sp.Config.RunMode)),
		}...)
	}
	for _, d := range directories {
		files = append(files, []string{
			filepath.Join(d, "override.ini"),
			filepath.Join(d, fmt.Sprintf("override.%s.ini", sp.Config.RunMode)),
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
				sp.Logger().Errorf("%s load err: %v", path, err)
			}
		} else {
			if last != nil {
				conf.SetParent(last)
			}
			last = conf
			sp.Logger().Infof("%s load success", path)
		}
	}

	if last == nil {
		last = config.Global
	}

	conf = last
	sp.ImportConfig(last)
	config.Decode(sp.Config, env)
	return
}

func UseGlobalLogger(sp *strip.Strip) {
	X.SetFlatLine(sp.Config.RunMode.IsProd())
	X.SetColorMode(sp.Config.RunMode.IsDev())
	X.SetShortLine(sp.Config.RunMode.IsProd())
	sp.SetLogger(X)
	sp.ProvideAs(X, (*strip.ReqLogger)(nil))
}
