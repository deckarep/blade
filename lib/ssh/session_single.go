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
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

func newSingleExecution(client *ssh.Client, hostname, command string, index int) *singleExecution {
	return &singleExecution{
		client:       client,
		command:      command,
		commandIndex: index,
		hostname:     hostname,
	}
}

type singleExecution struct {
	attempt int
	success int

	client *ssh.Client

	command      string
	commandIndex int

	hostname string
}

func (se *singleExecution) execute() {
	backoff.RetryNotify(func() error {
		return se.do()
	}, backoff.WithMaxTries(backoff.NewExponentialBackOff(), 2),
		func(err error, dur time.Duration) {
			// TODO: handle this better.
			//log.Println("Retry notify single command callback: ", err.Error())
			// I can do this...we're in the context of a single goroutine.
			se.attempt++
		},
	)
}

// do - the rule is one command can only ever occur per session.
func (se *singleExecution) do() error {
	var finalError error
	defer func() {
		if finalError != nil {
			sessionLogger.Println(color.YellowString(se.hostname) + fmt.Sprintf(" error %s", finalError.Error()))
		}
	}()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := se.client.NewSession()
	if err != nil {
		finalError = fmt.Errorf("Failed to create session: %s", err.Error())
		return finalError
	}
	defer session.Close()

	out, err := session.StdoutPipe()
	if err != nil {
		finalError = fmt.Errorf("Couldn't create pipe to Stdout for session: %s", err.Error())
		return finalError
	}

	errOut, err := session.StderrPipe()
	if err != nil {
		finalError = fmt.Errorf("Couldn't create pipe to Stderr for session: %s", err.Error())
		return finalError
	}

	currentHost := strings.Split(se.hostname, ":")[0]

	// Consume session Stdout, Stderr pipe async.
	go consumeReaderPipes(currentHost, out, false, 0)
	go consumeReaderPipes(currentHost, errOut, true, se.attempt)

	// Once a Session is created, you can only ever execute a single command.
	if err := session.Run(se.command); err != nil {
		// TODO: use this line for more verbose error logging since Stderr is also displayed.
		//sessionLogger.Print(color.RedString(currentHost+":") + fmt.Sprintf(" Failed to run the %s command: `%s` - %s", humanize.Ordinal(index), command, err.Error()))
		return err
	}
	se.success++
	return nil
}
