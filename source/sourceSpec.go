// Copyright 2017-present The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package source

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/gohugoio/hugo/config"
	"github.com/gohugoio/hugo/helpers"
	"github.com/gohugoio/hugo/hugofs"
	"github.com/spf13/cast"
)

// SourceSpec abstracts language-specific file creation.
// TODO(bep) rename to Spec
type SourceSpec struct {
	Cfg config.Provider
	Fs  *hugofs.Fs

	// This is set if the ignoreFiles config is set.
	ignoreFilesRe []*regexp.Regexp

	Languages              map[string]interface{}
	DefaultContentLanguage string
}

// NewSourceSpec initializes SourceSpec using languages from a given configuration.
func NewSourceSpec(cfg config.Provider, fs *hugofs.Fs) *SourceSpec {
	defaultLang := cfg.GetString("defaultContentLanguage")
	languages := cfg.GetStringMap("languages")

	if len(languages) == 0 {
		l := helpers.NewDefaultLanguage(cfg)
		languages[l.Lang] = l
		defaultLang = l.Lang
	}

	ignoreFiles := cast.ToStringSlice(cfg.Get("ignoreFiles"))
	var regexps []*regexp.Regexp
	if len(ignoreFiles) > 0 {
		for _, ignorePattern := range ignoreFiles {
			re, err := regexp.Compile(ignorePattern)
			if err != nil {
				helpers.DistinctErrorLog.Printf("Invalid regexp %q in ignoreFiles: %s", ignorePattern, err)
			} else {
				regexps = append(regexps, re)
			}

		}
	}

	return &SourceSpec{ignoreFilesRe: regexps, Cfg: cfg, Fs: fs, Languages: languages, DefaultContentLanguage: defaultLang}
}

func (s *SourceSpec) IgnoreFile(filename string) bool {
	base := filepath.Base(filename)

	if len(base) > 0 {
		first := base[0]
		last := base[len(base)-1]
		if first == '.' ||
			first == '#' ||
			last == '~' {
			return true
		}
	}

	if len(s.ignoreFilesRe) == 0 {
		return false
	}

	for _, re := range s.ignoreFilesRe {
		if re.MatchString(filename) {
			return true
		}
	}

	return false
}

func (s *SourceSpec) IsRegularSourceFile(filename string) (bool, error) {
	fi, err := helpers.LstatIfOs(s.Fs.Source, filename)
	if err != nil {
		return false, err
	}

	if fi.IsDir() {
		return false, nil
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		link, err := filepath.EvalSymlinks(filename)
		fi, err = helpers.LstatIfOs(s.Fs.Source, link)
		if err != nil {
			return false, err
		}

		if fi.IsDir() {
			return false, nil
		}
	}

	return true, nil
}
