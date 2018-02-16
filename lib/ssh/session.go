/*
Open Source Initiative OSI - The MIT License (MIT):Licensing
The MIT License (MIT)
Copyright (c) 2017 Ralph Caraveo (deckarep@gmail.com)
Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package ssh

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/deckarep/blade/lib/recipe"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

var (
	hostQueue             = make(chan string)
	hostWg                sync.WaitGroup
	successfullyCompleted int32
	failedCompleted       int32
	sessionLogger         = log.New(os.Stdout, "", 0)
)

// StartSession kicks off a Blade Recipe as a session of work to be completed.
func StartSession(recipe *recipe.BladeRecipeYaml, modifier *SessionModifier) {
	// Assumme root.
	if recipe.Overrides.User == "" {
		recipe.Overrides.User = "root"
	}

	sshConfig := &ssh.ClientConfig{
		User: recipe.Overrides.User,
		Auth: []ssh.AuthMethod{
			SSHAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Apply recipe will apply the recipe arguments to the commands
	// assumming they're defined.
	// TODO: don't loop twice here, then later for each cmd processing.
	sshCmds, err := applyRecipeArgs(recipe.Args, recipe.Exec)
	if err != nil {
		log.Fatal("Failed to apply recipe arguments to commands with err:", err.Error())
	}

	actualConcurrency := modifier.FlagOverrides.Concurrency
	if actualConcurrency == 0 {
		actualConcurrency = recipe.Overrides.Concurrency
	}

	// Concurrency must be at least 1 to make progress.
	if actualConcurrency == 0 {
		actualConcurrency = 1
	}

	go consumeAndLimitConcurrency(sshConfig, sshCmds, actualConcurrency)

	// Flags take precedence.
	allHosts := modifier.FlagOverrides.Hosts
	if len(allHosts) == 0 {
		// Then Hosts declared on the recipe.
		allHosts = recipe.Hosts
	}

	// Finally HostLookup takes last precedence.
	if len(allHosts) == 0 && recipe.HostLookup != "" {
		// Otherwise do dynamic host lookup here.
		commandSlice := strings.Split(recipe.HostLookup, " ")
		out, err := exec.Command(commandSlice[0], commandSlice[1:]...).Output()
		if err != nil {
			fmt.Println("Couldn't execute command:", err.Error())
			return
		}

		allHosts = strings.Split(string(out), ",")
	}

	if len(allHosts) == 0 {
		log.Fatalf("No host or hostlookup defined for this recipe, alternatively use the --servers flag.")
	}

	log.Print(color.GreenString(fmt.Sprintf("Starting recipe: %s", recipe.Name)))

	actualPort := modifier.FlagOverrides.Port
	if actualPort == 0 {
		actualPort = recipe.Overrides.Port
	}
	if actualPort == 0 {
		actualPort = 22
	}

	totalHosts := len(allHosts)
	for _, h := range allHosts {
		enqueueHost(h, actualPort)
	}

	hostWg.Wait()
	log.Print(color.GreenString(fmt.Sprintf("Completed recipe: %s - %d success | %d failed | %d total",
		recipe.Name,
		atomic.LoadInt32(&successfullyCompleted),
		atomic.LoadInt32(&failedCompleted),
		totalHosts)))
}

func executeSession(sshConfig *ssh.ClientConfig, hostname string, commands []string) {

	// inline user overrides sshConfig.User
	userHost := strings.Split(hostname, "@")
	if len(userHost) == 2 {
		newSSHConfig := *sshConfig
		sshConfig = &newSSHConfig
		sshConfig.User = userHost[0]
		hostname = userHost[1]
	}

	backoff.RetryNotify(func() error {
		return startSSHSession(sshConfig, hostname, commands)
	}, backoff.WithMaxTries(backoff.NewExponentialBackOff(), 3),
		func(err error, dur time.Duration) {
			// TODO: handle this better.
			log.Println("Retry notify callback: ", err.Error())
		},
	)
}

func startSSHSession(sshConfig *ssh.ClientConfig, hostname string, commands []string) error {
	var finalError error
	defer func() {
		if finalError != nil {
			log.Println(color.YellowString(hostname) + fmt.Sprintf(" error %s", finalError.Error()))
		}
	}()

	client, err := ssh.Dial("tcp", hostname, sshConfig)
	if err != nil {
		finalError = fmt.Errorf("Failed to dial remote host: %s", err.Error())
		return finalError
	}

	// Since we can run multiple commands, we need to keep track of intermediate failures
	// and log accordingly or do some type of aggregate report.
	// Commands within a single session are executed in serial by design.
	for i, cmd := range commands {
		se := newSingleExecution(client, hostname, cmd, i+1)
		se.execute()
	}

	// Technically this is only successful when errors didn't occur above.
	atomic.AddInt32(&successfullyCompleted, 1)
	return nil
}

func applyRecipeArgs(args recipe.BladeRecipeArguments, commands []string) ([]string, error) {
	if len(args) == 0 {
		return commands, nil
	}

	// TODO: ensure at least all args are used at least once to minimize end-user errors.
	// TODO: also allow the args to become flags so you can override at the command line.

	var appliedResults []string
	for _, cmd := range commands {
		replacedCmd := cmd
		for _, arg := range args {
			argToken := fmt.Sprintf("{{%s}}", arg.Name())
			appliedFlagValue := arg.FlagValue()
			if arg.FlagValue() != "" {
				replacedCmd = strings.Replace(replacedCmd, argToken, appliedFlagValue, -1)
			} else {
				replacedCmd = strings.Replace(replacedCmd, argToken, arg.Value, -1)
			}
		}
		appliedResults = append(appliedResults, replacedCmd)
	}

	return appliedResults, nil
}
