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
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"roller/internal"
	"roller/pkg"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var profileName string
var fromProfile string
var accountID string
var region string
var role string
var browser bool

var awsSession *session.Session
var profiles *internal.Profiles
var credentials *internal.Credentials
var switchRoleParameters *pkg.SwitchRoleParameters

const tokenLifetime = 3600


// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:     "switch",
	Aliases: []string{"sw"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		_,exists := internal.AccountCache[args[0]]

		if !exists {
			return fmt.Errorf("The given role can not be loaded from the cache: %s", args[0])
		}

		return nil
	},
	Short:   "Switch to an AWS role.",
	Long: `Create a set of temporary credentials using STS and store them amongst the default AWS configurations.
After the credentials were created successfully, they can be used in the same way as any other AWS profile by the
name (-n, --name) specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			switchRoleParameters = &internal.AccountCache[args[0]].Parameters
			profileName = args[0]
		} else {
			switchRoleParameters = &pkg.SwitchRoleParameters{
				fromProfile,
				accountID,
				role,
			}
		}
		if accountID == "" && role == "" && profileName == "" && len(args) == 0 && os.Getenv("ROLLER_ACTIVE_PROFILE") != "" {
			profileName = os.Getenv("ROLLER_ACTIVE_PROFILE")
		}

		profiles = internal.ReadProfiles()
		credentials = internal.ReadCredentials()

		var activeCredentials *sts.Credentials

		if profileName == "" {
			syncAnonRoleParameters()
		} else {
			syncNamedRoleParameters(profileName)
		}

		if browser {
			openBrowser()
		} else {
			needsRefresh := true
			creds, ok := credentials.Credentials[profileName]

			if ok {
				limit := time.Now()
				limit.Add(5 * 60 * 1000 * 1000) // 5 minutes in nanoseconds

				if creds.Expiration.After(limit) {
					needsRefresh = false
				}
			}

			if needsRefresh {
				activeCredentials = switchTo(*switchRoleParameters)
				credentials.Add(profileName, &internal.Credential{
					Expiration: *activeCredentials.Expiration,
					AccessKey:  *activeCredentials.AccessKeyId,
					SecretKey:  *activeCredentials.SecretAccessKey,
					Token:      *activeCredentials.SessionToken,
				})

				profiles.Save()
				credentials.Save()
			}

			if viper.GetBool("shell") {
				fmt.Printf("export ROLLER_ACTIVE_PROFILE=%s " +
					"&& export AWS_PROFILE=%s " +
					"&& export RPROMPT='<aws:%s>'\n", profileName, profileName, profileName)
			}
		}
	},
}

func createSession() *session.Session {
	if awsSession != nil {
		return awsSession
	}
	awsSession = session.Must(session.NewSessionWithOptions(session.Options{
		Profile: switchRoleParameters.FromProfile,
	}))

	return awsSession
}

func currentAccount() (string, string) {
	svc := iam.New(createSession())
	regex, _ := regexp.Compile(`User: (arn:aws:iam::\d+:user/.+) is`)
	result, err := svc.GetUser(&iam.GetUserInput{})
	var arn string
	var match []string
	if err != nil {
		match = regex.FindStringSubmatch(err.Error())

		if match != nil {
			arn = match[1]
		} else {
			internal.ExitOnError(err)
		}
	} else {
		arn = *result.User.Arn
	}

	regex, _ = regexp.Compile(`arn:aws:iam::(\d+):user/(.+)`)
	match = regex.FindStringSubmatch(arn)

	return match[1], match[2]
}

func switchTo(role pkg.SwitchRoleParameters) *sts.Credentials {
	current_account_id, username := currentAccount()
	svc := sts.New(createSession())

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter your MFA for %s for your %s profile: \n", username, role.FromProfile)
	mfa, _ := reader.ReadString('\n')
	result, err := svc.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", role.AccountId, role.Role)),
		RoleSessionName: aws.String("ROLLER"),
		SerialNumber:    aws.String(fmt.Sprintf("arn:aws:iam::%s:mfa/%s", current_account_id, username)),
		TokenCode:       aws.String(strings.Trim(mfa, "\n")),
		DurationSeconds: aws.Int64(tokenLifetime),
	})
	internal.ExitOnError(err)

	return result.Credentials
}

func syncNamedRoleParameters(name string) {
	profile, ok := profiles.Profiles[name]
	if !ok {
		p := internal.Profile{}
		syncParameters(&p)
		profiles.Add(name, &p)
	} else {
		syncParameters(profile)
	}
}

func syncAnonRoleParameters() {
	profile := internal.Profile{}
	syncParameters(&profile)
	profileName = profile.GenerateName()
	profiles.Add(profileName, &profile)
}

func syncParameters(profile *internal.Profile) {
	if switchRoleParameters.FromProfile != "" {
		profile.Profile = switchRoleParameters.FromProfile
	} else if profile.Profile != "" {
		switchRoleParameters.FromProfile = profile.Profile
	} else {
		switchRoleParameters.FromProfile = viper.GetString("profile")
	}

	if switchRoleParameters.AccountId != "" {
		profile.Account = switchRoleParameters.AccountId
	} else if profile.Account != "" {
		switchRoleParameters.AccountId = profile.Account
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("AWS account ID you want to switch to: ")
		input, _ := reader.ReadString('\n')
		input = strings.Trim(input, "\n")
		switchRoleParameters.AccountId = input
		profile.Account = input
	}

	if switchRoleParameters.Role != "" {
		profile.Role = switchRoleParameters.Role
	} else if profile.Role != "" {
		switchRoleParameters.Role = profile.Role
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Name of the role you want to switch to: ")
		input, _ := reader.ReadString('\n')
		input = strings.Trim(input, "\n")
		switchRoleParameters.Role = input
		profile.Role = input
	}

	if region != "" {
		profile.Region = region
	}
}

func openBrowser() {
	url := fmt.Sprintf("https://signin.aws.amazon.com/switchrole?account=%s&roleName=%s&displayName=%s",
		switchRoleParameters.AccountId,
		switchRoleParameters.Role,
		profileName,
	)

	open.Run(url)
}

func init() {
	switchCmd.Flags().Bool("shell", false, "Write an eval'able statement on successful switching to wrap in a shell function.")
	switchCmd.Flags().StringVarP(&profileName, "name", "n", "", "The name of the role to switch to. If the role is not set up, it will be saved if successful.")
	switchCmd.Flags().StringVarP(&fromProfile, "profile", "p", "", "The name of an existing profile to use to switch from. Defaults to 'default'.")
	switchCmd.Flags().StringVar(&region, "region", "", "The default region to set for the role.")
	switchCmd.Flags().StringVar(&accountID, "account", "", "The account id to switch to.")
	switchCmd.Flags().StringVar(&role, "role", "", "The AWS role name to switch to.")
	switchCmd.Flags().BoolVarP(&browser, "web", "w", false, "Open a browser tab to switch to the role.")

	RootCmd.AddCommand(switchCmd)
	viper.SetDefault("profile", "default")
}
