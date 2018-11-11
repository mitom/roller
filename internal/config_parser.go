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
    "gopkg.in/ini.v1"
    "strings"
    "time"
    "fmt"
    "path"
)


type Profile struct {
    Profile string `ini:"roller_profile,omitempty"`
    Account string `ini:"roller_account"`
    Role string `ini:"roller_role"`
    Region string `ini:"region,omitempty"`
    Roller bool `ini:"roller"`
}

func (p Profile) GenerateName() string {
    return fmt.Sprintf("%s/%s", p.Account, p.Role)
}

type Profiles struct {
    data *ini.File
    Profiles map[string]*Profile
}

func NewProfiles(data *ini.File) Profiles {
    profiles := Profiles{data, make(map[string]*Profile)}

    return profiles
}

func (p Profiles) Add(name string, profile *Profile) {
    section, err := p.data.NewSection("profile " + name)
    PanicOnError(err)
    profile.Roller = true
    section.ReflectFrom(profile)
    p.Profiles[name] = profile
}

func (p Profiles) Update(name string, profile *Profile) {
    section := p.data.Section("profile " + name)
    section.ReflectFrom(profile)
}

func (p Profiles) Delete(name string) {
    delete(p.Profiles, name)
    p.data.DeleteSection("profile " + name)
}

func (p Profiles) Load(section *ini.Section) {
    profile := new(Profile)
    section.MapTo(profile)
    if section.Name() == "DEFAULT" {
        return
    }

    p.Profiles[strings.TrimPrefix(section.Name(), "profile ")] = profile
}

func (p Profiles) Save() {
    p.data.SaveTo(HomePath()+"/.aws/config")
}

type Credential struct {
    Expiration time.Time `ini:"expiration"`
    AccessKey string `ini:"aws_access_key_id"`
    SecretKey string `ini:"aws_secret_access_key"`
    Token string `ini:"aws_session_token"`
}

type Credentials struct {
    data *ini.File
    Credentials map[string]*Credential
}

func NewCredentials(data *ini.File) Credentials {
    credentials := Credentials{data, make(map[string]*Credential)}

    return credentials
}

func (c Credentials) Add(name string, credential *Credential) {
    section, err := c.data.NewSection(name)
    PanicOnError(err)
    section.ReflectFrom(credential)
    c.Credentials[name] = credential
}

func (c Credentials) Delete(name string) {
    delete(c.Credentials, name)
    c.data.DeleteSection(name)
}

func (c Credentials) Load(section *ini.Section) {
    credential := new(Credential)
    section.MapTo(credential)
    if section.Name() == "DEFAULT" {
        return
    }

    c.Credentials[section.Name()] = credential
}

func (c Credentials) Save() {
    c.data.SaveTo(path.Join(HomePath(), ".aws", "credentials"))
}


func ReadProfiles() *Profiles{
    cfg, err := ini.Load(path.Join(HomePath(), ".aws", "config"))
    ExitOnError(err)

    profiles := NewProfiles(cfg)

    for _, section := range cfg.Sections() {
        profiles.Load(section)
    }

    return &profiles
}


func ReadCredentials() *Credentials{
    cfg, err := ini.Load(path.Join(HomePath(), ".aws", "credentials"))
    ExitOnError(err)

    credentials := NewCredentials(cfg)

    for _, section := range cfg.Sections() {
        credentials.Load(section)
    }

    return &credentials
}

