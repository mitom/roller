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
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/mitom/roller/internal"

	"github.com/spf13/cobra"
)

const ZSH = `export ROLLER_PATH='%s';

function __roller_args {
    case $line[1] in
        switch|sw)
            _arguments \
                '*: :->roles' \
                '-w: :->roles' \
                '--web: :->roles' \
                '-n: :->profiles' \
                '--name: :->profiles' \
                '-p: :->profiles' \
                '--profile: :->profiles'

            case $state in
                roles)
                    local -a roles
                    roles=($(ROLLER_SHELL=true ${ROLLER_PATH} cache))
                    _multi_parts -M 'm:{a-z}={A-Z}' / roles
                    ;;
                profiles)
                    compadd $(grep "\[profile .*\]" ~/.aws/config |sed -e 's/\[profile \([a-zA-Z0-9_\.\/\-]*\)\]/\1/')
                    ;;
        esac
        ;;
    esac
};

function __roller_autocomplete {
    _arguments \
        '1: :__roller_cmds' \
        '*: :__roller_args'
};

function roller {
    ROLLER_SHELL=true ${ROLLER_PATH} "$@" | {
      while IFS= read -r line
      do
        if [[ "$line" == export\ * ]]; then
            eval "$line"
        else
            echo "$line"
        fi
      done
    }
};

function __roller_cmds {
    local commands
    commands=(%s)
    _describe 'roller' commands
};

compdef __roller_autocomplete roller;
`

const COMMANDS_TEMPLATE = `{{ range . }}
'{{ .Name }}:{{ .Short }}'
{{- end }}`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up the current shell.",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := os.Executable()
		tmpl, err := template.New("cmds").Parse(COMMANDS_TEMPLATE)
		internal.ExitOnError(err)

		var cmds bytes.Buffer
		tmpl.Execute(&cmds, RootCmd.Commands())

		fmt.Printf(ZSH, path, cmds.String())
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
