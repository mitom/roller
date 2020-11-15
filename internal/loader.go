// Copyright © 2018 Tamas Millian <tamas.millian@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"plugin"
	"regexp"
	"time"

	"github.com/mitom/roller/internal/csv_loader"
	"github.com/mitom/roller/pkg"

	"github.com/spf13/viper"
)

var AccountCache map[string]*pkg.LoadedProfile
var nameReplaceRe = regexp.MustCompile(`[^\d\w\-]`)

type SerialisedCache struct {
	ValidUntil time.Time
	Data       *[]pkg.LoadedProfile
}

func createLoaderConfig(name string, givenConfig interface{}) *pkg.LoaderConfig {
	cf := givenConfig.(map[string]interface{})

	loader, ok := cf["loader"].(string)
	if !ok {
		ExitWithError(fmt.Sprintf("Invalid value for loader: %s", cf["loader"]), 1)
	}
	options, ok := cf["options"].(map[string]interface{})
	if !ok {
		ExitWithError(fmt.Sprintf("Invalid value for options: %s", cf["options"]), 1)
	}
	ttl, ok := cf["ttl"].(int)
	if !ok {
		ExitWithError(fmt.Sprintf("Invalid value for ttl: %s", cf["ttl"]), 1)
	}

	return pkg.NewLoaderConfig(name, loader, options, ttl)
}

func ClearCache() {
	os.RemoveAll(path.Join(viper.GetString("cache_dir")))
}

func LoadCache() {
	results := make(map[string]*pkg.LoadedProfile)
	givenConfig := viper.Get("loader")
	configs := givenConfig.(map[string]interface{})
	now := time.Now()
	os.Mkdir(viper.GetString("cache_dir"), 0700)
	for cacheName, c := range configs {
		cfg := createLoaderConfig(cacheName, c)
		var loaded *[]pkg.LoadedProfile
		if cfg.GetTtl() > 0 {
			read, _ := ioutil.ReadFile(path.Join(viper.GetString("cache_dir"), cfg.GetName()+".json"))
			if read != nil {
				var serialised SerialisedCache
				if err := json.Unmarshal(read, &serialised); err != nil {
					fmt.Printf("Warning: can not read the cache for %s, ignoring it.\n", cfg.GetName())
				} else if serialised.ValidUntil.After(now) {
					loaded = serialised.Data
				}
			}
		}
		if loaded == nil {
			var loader pkg.Loader
			if cfg.GetLoader() == "csv" {
				loader = csv_loader.Loader
			} else {
				pluginPath := viper.GetString("plugin_dir")

				modulePath := path.Join(pluginPath, cfg.GetLoader())
				plug, err := plugin.Open(modulePath)
				ExitOnError(err)

				// LoadCache the module
				symLoader, err := plug.Lookup("Loader")
				ExitOnError(err)

				// Assert that the interface is implemented
				var ok bool
				loader, ok = symLoader.(pkg.Loader)
				if !ok {
					ExitWithError(fmt.Sprintf("%s is either outdated or invalid! Make sure it implements the proper interface.\n", pluginPath), 1)
				}
			}

			l := loader.Load(cfg)
			loaded = &l

			if cfg.GetTtl() > 0 {
				expiration := now.Add(time.Duration(cfg.GetTtl()) * time.Second)
				toSerialise := SerialisedCache{
					expiration,
					loaded,
				}
				serialised, err := json.Marshal(toSerialise)
				if err != nil {
					fmt.Printf("Warning: could not serialise %s, can not cache it.", cfg.GetName())
					fmt.Println(err)
				}

				err = ioutil.WriteFile(path.Join(viper.GetString("cache_dir"), cfg.GetName()+".json"), serialised, 0600)
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		for _, r := range *loaded {
			if !r.Parameters.Valid() {
				continue
			}
			name := nameReplaceRe.ReplaceAllString(r.Name, "-")
			name += "/" + r.Parameters.Role
			current, exists := results[name]

			if exists && !viper.GetBool("shell") && current.Parameters.AccountID != r.Parameters.AccountID {
				fmt.Fprintf(os.Stderr,
					"There is an account ID mismatch between 2 loaders for %s. Already had %s, ignoring %s.\n",
					name,
					current.Parameters.AccountID,
					r.Parameters.AccountID)
			} else {
				//make a copy of the account as range reuses the memory for r
				m := r
				results[name] = &m
			}
		}
	}

	AccountCache = results
}
