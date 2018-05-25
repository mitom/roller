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
    "fmt"
    "os"
    "os/user"
    "path"
)

func ExitOnError(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func ExitWithError(err string, code int) {
    fmt.Println(err)
    os.Exit(code)
}

func PanicOnError(err error) {
    if err != nil {
        panic(err)
    }
}

func HomePath() (string) {
    usr, err := user.Current()
    ExitOnError(err)
    return usr.HomeDir
}

// Return the path to the base dir of the application config/cache
func AppHomePath() (string) {
    return path.Join(HomePath(), ".roller")
}
