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
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/user"
	"path"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/mikkeloscar/sshconfig"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var (
	ConfigMapping []*sshconfig.SSHHost
)

func init() {
	mapping, err := ParseSSHConfig()
	if err != nil {
		log.Println("Failed to parse SSHConfig mapping")
	}
	ConfigMapping = mapping
}

// SSHAgent queries the host operating systems SSH agent.
func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

// // LoadSSHUserName - thanks O.G. Sam.
// func LoadSSHUserName() (string, error) {
// 	sshFilePath := path.Join(os.Getenv("HOME"), ".ssh", "config")
// 	sshConfig, err := ioutil.ReadFile(sshFilePath)
// 	if err != nil {
// 		return "", errors.New("could not load your SSH config")
// 	}

// 	re, _ := regexp.Compile(`User (\w+)`)
// 	matches := re.FindStringSubmatch(string(sshConfig))
// 	if len(matches) > 1 {
// 		return matches[1], nil
// 	}
// 	return "", errors.New("could not find your SSH username. is it defined in ~/.ssh/config?")
// }

func ParseSSHConfig() ([]*sshconfig.SSHHost, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	hosts, err := sshconfig.ParseSSHConfig(path.Join(usr.HomeDir, ".ssh", "config"))
	if err != nil {
		return nil, err
	}

	return hosts, nil
}

func consumeAndLimitConcurrency(sshConfig *ssh.ClientConfig, commands []string) {
	for host := range hostQueue {
		concurrencySem <- 1
		go func(h string) {
			defer func() {
				<-concurrencySem
				hostWg.Done()
			}()
			executeSession(sshConfig, h, commands)
		}(host)
	}
}

func enqueueHost(host string, port int) {
	host = strings.TrimSpace(host)

	// If it doesn't contain the port; add it.
	if !strings.Contains(host, ":") {
		host = fmt.Sprintf("%s:%d", host, port)
	}

	// Ignore what can't be parsed as host:port.
	_, _, err := net.SplitHostPort(host)
	if err != nil {
		log.Printf("Couldn't parse: %s", host)
		return
	}

	// Finally, enqueue it up for processing.
	hostWg.Add(1)
	hostQueue <- host
}

func consumeReaderPipes(host string, rdr io.Reader, isStdErr bool, attempt int) {
	logHost := color.CyanString(host + ":")

	if isStdErr {
		attemptString := ""
		if attempt > 1 {
			attemptString = fmt.Sprintf(" (%s attempt)", humanize.Ordinal(attempt))
		}
		logHost = color.RedString(host + attemptString + ":")
	}

	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		sessionLogger.Println(logHost + " " + scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		sessionLogger.Print(color.RedString(host) + ": Error reading output from this host.")
	}
}
