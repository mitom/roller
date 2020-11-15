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

package pkg

type CacheLoader interface {
	Load(config *CacheConfig) []CachedProfile
}

type SwitchRoleParameters struct {
	FromProfile string
	AccountId   string
	Role        string
}

func (p SwitchRoleParameters) Valid() bool {
	return true
}

type CachedProfile struct {
	Name       string
	Parameters SwitchRoleParameters
}

type CacheConfig struct {
	name    string
	options map[string]interface{}
	ttl     int
	plugin  string
}

func (c CacheConfig) GetOptions() map[string]interface{} {
	return c.options
}

func (c CacheConfig) GetName() string {
	return c.name
}

func (c CacheConfig) GetTtl() int {
	return c.ttl
}

func (c CacheConfig) GetPlugin() string {
	return c.plugin
}

func NewCacheConfig(name string, pluginPath string, options map[string]interface{}, ttl int) *CacheConfig {

	c := CacheConfig{
		name:    name,
		plugin:  pluginPath,
		ttl:     ttl,
		options: options,
	}

	return &c
}
