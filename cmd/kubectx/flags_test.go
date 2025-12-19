// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/ahmetb/kubectx/internal/cmdutil"
	"github.com/ahmetb/kubectx/internal/env"
	"github.com/ahmetb/kubectx/internal/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type parseArgsTest struct {
	name string
	args []string
	want Op
}

// Test cases which are the same for all modes
func parseArgCommonTests() []parseArgsTest {
	return []parseArgsTest{
		{name: "help shorthand",
			args: []string{"-h"},
			want: HelpOp{}},
		{name: "help long form",
			args: []string{"--help"},
			want: HelpOp{}},
		{name: "current shorthand",
			args: []string{"-c"},
			want: CurrentOp{}},
		{name: "current long form",
			args: []string{"--current"},
			want: CurrentOp{}},
		{name: "unset shorthand",
			args: []string{"-u"},
			want: UnsetOp{}},
		{name: "unset long form",
			args: []string{"--unset"},
			want: UnsetOp{}},
		{name: "delete - current context",
			args: []string{"-d", "."},
			want: DeleteOp{[]string{"."}}},
		{name: "delete - multiple contexts",
			args: []string{"-d", ".", "a", "b"},
			want: DeleteOp{[]string{".", "a", "b"}}},
		{name: "rename context",
			args: []string{"a=b"},
			want: RenameOp{"a", "b"}},
		{name: "rename context with old=current",
			args: []string{"a=."},
			want: RenameOp{"a", "."}},
		{name: "unrecognized flag",
			args: []string{"-x"},
			want: UnsupportedOp{Err: fmt.Errorf("unsupported option '-x'")}},
	}
}

func Test_parseArgs_nonInteractive(t *testing.T) {
	tests := []parseArgsTest{
		{name: "nil Args",
			args: nil,
			want: ListOp{}},
		{name: "empty Args",
			args: []string{},
			want: ListOp{}},
		{name: "switch by name",
			args: []string{"foo"},
			want: SwitchOp{Target: "foo"}},
		{name: "switch by swap",
			args: []string{"-"},
			want: SwitchOp{Target: "-"}},
		{name: "delete - without contexts",
			args: []string{"-d"},
			want: UnsupportedOp{fmt.Errorf("'-d' needs arguments")}},
		{name: "too many args",
			args: []string{"a", "b", "c"},
			want: UnsupportedOp{Err: fmt.Errorf("too many arguments")}},
	}
	tests = append(tests, parseArgCommonTests()...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &argParser{isInteractiveMode: func(*os.File) bool { return false }, isFZFFallbackEnabled: cmdutil.IsFZFFallbackEnabled}
			got := parser.ParseArgs(tt.args)

			if diff := cmp.Diff(got, tt.want, cmpOpts()...); diff != "" {
				t.Errorf("parseArgs(%#v) diff: %s", tt.args, diff)
			}
		})
	}
}

func Test_parseArgs_interactive_fzfFallbackDisabled(t *testing.T) {
	t.Cleanup(testutil.WithEnvVar(env.EnvFZFFallback, ""))
	tests := []parseArgsTest{
		{name: "nil Args",
			args: nil,
			want: InteractiveSwitchOp{}},
		{name: "empty Args",
			args: []string{},
			want: InteractiveSwitchOp{}},
		{name: "switch by name",
			args: []string{"foo"},
			want: SwitchOp{Target: "foo"}},
		{name: "switch by swap",
			args: []string{"-"},
			want: SwitchOp{Target: "-"}},
		{name: "delete - without contexts",
			args: []string{"-d"},
			want: InteractiveDeleteOp{}},
		{name: "too many args",
			args: []string{"a", "b", "c"},
			want: UnsupportedOp{Err: fmt.Errorf("too many arguments")}},
	}
	tests = append(tests, parseArgCommonTests()...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &argParser{isInteractiveMode: func(*os.File) bool { return true }, isFZFFallbackEnabled: cmdutil.IsFZFFallbackEnabled}
			got := parser.ParseArgs(tt.args)

			if diff := cmp.Diff(got, tt.want, cmpOpts()...); diff != "" {
				t.Errorf("parseArgs(%#v) diff: %s", tt.args, diff)
			}
		})
	}
}

func Test_parseArgs_interactive_fzfFallbackEnabled(t *testing.T) {
	t.Cleanup(testutil.WithEnvVar(env.EnvFZFFallback, "1"))
	tests := []parseArgsTest{
		{name: "nil Args",
			args: nil,
			want: InteractiveSwitchOp{}},
		{name: "empty Args",
			args: []string{},
			want: InteractiveSwitchOp{}},
		{name: "switch by name",
			args: []string{"foo"},
			want: InteractiveSwitchOp{Queries: []string{"foo"}}},
		{name: "switch by swap",
			args: []string{"-"},
			want: SwitchOp{Target: "-"}},
		{name: "delete - without contexts",
			args: []string{"-d"},
			want: InteractiveDeleteOp{}},
		{name: "multiple args",
			args: []string{"a", "b", "c"},
			want: InteractiveSwitchOp{Queries: []string{"a", "b", "c"}}},
	}
	tests = append(tests, parseArgCommonTests()...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &argParser{isInteractiveMode: func(*os.File) bool { return true }, isFZFFallbackEnabled: cmdutil.IsFZFFallbackEnabled}
			got := parser.ParseArgs(tt.args)

			if diff := cmp.Diff(got, tt.want, cmpOpts()...); diff != "" {
				t.Errorf("parseArgs(%#v) diff: %s", tt.args, diff)
			}
		})
	}
}

func cmpOpts() cmp.Options {
	return cmp.Options{
		cmp.Comparer(func(x, y UnsupportedOp) bool {
			return (x.Err == nil && y.Err == nil) || (x.Err.Error() == y.Err.Error())
		}),
		cmpopts.IgnoreFields(InteractiveSwitchOp{}, "SelfCmd"),
		cmpopts.IgnoreFields(InteractiveDeleteOp{}, "SelfCmd"),
	}
}
