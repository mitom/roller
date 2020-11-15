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
	"github.com/mitom/roller/internal"

	"github.com/spf13/cobra"

	"time"
)

var includeNamed bool

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up the aws configuration files.",
	Long: `Removes all credentials which were created 
    by Roller and expired for over an hour. By default,
    it will leave named profiles.`,
	Run: func(cmd *cobra.Command, args []string) {
		profiles = internal.ReadProfiles()
		credentials = internal.ReadCredentials()
		limit := time.Now()
		limit.Add(-60 * 60 * 1000 * 1000) // 1 hour in nanoseconds

		dirty := false
		for name, v := range credentials.Credentials {
			profile, ok := profiles.Profiles[name]

			// if it doesn't have a profile or not roller managed
			// or not expired for over an hour yet, ignore it.
			if !ok ||
				!profile.Roller ||
				!v.Expiration.Before(limit) ||
				(!includeNamed && name != profile.GenerateName()) {
				continue
			}

			credentials.Delete(name)
			profiles.Delete(name)

			dirty = true
		}

		if dirty {
			profiles.Save()
			credentials.Save()
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().BoolVar(&includeNamed, "include-named", false, "Include named profiles.")
}
