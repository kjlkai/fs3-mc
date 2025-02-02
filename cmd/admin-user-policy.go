// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/fatih/color"
	jsoniter "github.com/json-iterator/go"
	"github.com/minio/cli"
	"github.com/filswan/fs3-mc/pkg/probe"
	"github.com/minio/minio/pkg/console"
	iampolicy "github.com/minio/minio/pkg/iam/policy"
)

var adminUserPolicyCmd = cli.Command{
	Name:         "policy",
	Usage:        "export user policies in JSON format",
	Action:       mainAdminUserPolicy,
	OnUsageError: onUsageError,
	Before:       setGlobalsFromContext,
	Flags:        globalFlags,
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} TARGET USERNAME

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
  1. Display the policy document of a user "foobar" in JSON format.
     {{.Prompt}} {{.HelpName}} myminio foobar

`,
}

// checkAdminUserPolicySyntax - validate all the passed arguments
func checkAdminUserPolicySyntax(ctx *cli.Context) {
	if len(ctx.Args()) != 2 {
		cli.ShowCommandHelpAndExit(ctx, "policy", 1) // last argument is exit code
	}
}

// mainAdminUserPolicy is the handler for "mc admin user policy" command.
func mainAdminUserPolicy(ctx *cli.Context) error {
	checkAdminUserPolicySyntax(ctx)

	console.SetColor("UserMessage", color.New(color.FgGreen))

	// Get the alias parameter from cli
	args := ctx.Args()
	aliasedURL := args.Get(0)

	// Create a new MinIO Admin Client
	client, err := newAdminClient(aliasedURL)
	fatalIf(err, "Unable to initialize admin connection.")

	user, e := client.GetUserInfo(globalContext, args.Get(1))
	fatalIf(probe.NewError(e).Trace(args...), "Unable to get user info")

	var combinedPolicy iampolicy.Policy

	policies := strings.Split(user.PolicyName, ",")

	for _, p := range policies {
		buf, e := client.InfoCannedPolicy(globalContext, p)
		fatalIf(probe.NewError(e).Trace(args...), "Unable to fetch user policy document")
		policy, e := iampolicy.ParseConfig(bytes.NewReader(buf))
		fatalIf(probe.NewError(e).Trace(args...), "Unable to parse user policy document")
		combinedPolicy = combinedPolicy.Merge(*policy)
	}

	var jsoniter = jsoniter.ConfigCompatibleWithStandardLibrary
	policyJSON, e := jsoniter.MarshalIndent(combinedPolicy, "", "   ")
	fatalIf(probe.NewError(e).Trace(args...), "Unable to parse user policy document")

	fmt.Println(string(policyJSON))

	return nil
}
