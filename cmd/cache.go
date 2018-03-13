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

package cmd

import (
	"fmt"
	"path"

	"roller/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Dump the account cache.",
	Run: func(cmd *cobra.Command, args []string) {
		for k := range internal.AccountCache {
			fmt.Println(k)
		}
	},
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the account cache.",
	Run: func(cmd *cobra.Command, args []string) {
		internal.ClearCache()
	},
}

func init() {
	RootCmd.AddCommand(cacheCmd)
	cacheCmd.Flags().Bool("shell", false, "Avoid printing warnings.")
	cacheCmd.AddCommand(cacheClearCmd)
	viper.SetDefault("plugin_dir", path.Join(internal.AppHomePath(), "plugins"))
	viper.SetDefault("cache_dir", path.Join(internal.AppHomePath(), "cache"))
}
