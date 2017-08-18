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

package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// Flag variables
var (
	retries     int
	concurrency int
	hosts       string
	port        int
	user        string
)

// App variables
var (
	concurrencySem chan int
	hostQueue      = make(chan string)
	hostWg         sync.WaitGroup
)

// blade ssh deploy-cloud-server-a // matches a recipe and therefore will follow the recipe guidelines against servers defined in recipe
// blade ssh deploy 127.0.0.1,128.0.0.1,129.0.0.1 //matches a recipe but uses the servers noted instead
// A recipe can be defined as the following
// 1. a single and simple command, with no servers specified, it's on you to provide a list.
// 2. a complex command, your recipe defines the server list.
// 3. a timed, tuned sequence of commands...perhaps it denotes a deploy order of multiple clusters, some sleep time, and perhaps even prompting to continue.
// 4. also, blade can cache server fetches...and has a default timeout
// 5. recipes are meant to be shared in some repository, this means that you can share with the team
// 6. recipes can be overriden per user, usually not a good idea...but possible still.
// 7. a recipe can be either:
//    a. a single command no servers specified/servers specified
//    b. a series of steps no servers specified/servers specified
//    c. a series of steps with sleep intervals baked in/continue prompt baked in
//    d. an aggregate, intended to execute on all servers but returning aggregates
//    e. banner prompts baked in, when dealing with production services (think: are you sure you want to continue?)
//    f. validation commands, intended on revealing some truthyness about infastructure
//    g. all recipes should emit proper log reports, and can tie into checklist as a service
//    h. integration into hipchat/slack/etc. when key events occur.
//    i. ability to view a simplified output of progress, notify end-user with gosx-notifier
//    j. ability to tie into ANY query infrastructure by allowing you to either pass in comma delimited servers
//          - or add something like this to your recipe: `ips prod redis-hosts -c`
// 8. blade generate where generate a database of all your recipes as Cobra Commands.
// 9. A seperate bolt database stores hosts cache for speed.

func init() {
	sshCmd.Flags().StringVarP(&hosts, "hosts", "x", "", "--hosts flag is one or more comma delimited hosts.")
	sshCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "Max concurrency when running ssh commands")
	sshCmd.Flags().IntVarP(&retries, "retries", "r", 3, "Number of times to retry until a successful command returns")
	sshCmd.Flags().IntVarP(&port, "port", "p", 22, "The ssh port to use")
	sshCmd.Flags().StringVarP(&user, "user", "u", "root", "--user for ssh")
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "ssh [command]",
	Long:  "ssh [command] will execute the [command] on all servers. If the command matches a recipe name first, the recipe will be used in place of the command.",
	Run: func(cmd *cobra.Command, args []string) {
		sshConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				sshAgent(),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		sshCmd := "hostname"
		if len(args) > 0 {
			sshCmd = args[0]
		}

		concurrencySem = make(chan int, concurrency)
		go consumeAndLimitConcurrency(sshConfig, sshCmd)

		allHosts := strings.Split(hosts, ",")
		totalHosts := len(allHosts)
		for _, h := range allHosts {
			startHost(h)
		}

		hostWg.Wait()
		log.Print(color.GreenString(fmt.Sprintf("Finished: x out of %d.", totalHosts)))
	},
}

// TODO: most of this code was copied from the ssh command code above
// TODO: REFACTOR BRO
func startSSHSession(recipe *Recipe) {
	sshConfig := &ssh.ClientConfig{
		User: recipe.User,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshCmd := recipe.Command

	concurrencySem = make(chan int, recipe.Concurrency)
	go consumeAndLimitConcurrency(sshConfig, sshCmd)

	allHosts := recipe.Hosts
	totalHosts := len(allHosts)
	for _, h := range allHosts {
		startHost(h)
	}

	hostWg.Wait()
	log.Print(color.GreenString(fmt.Sprintf("Finished: x out of %d.", totalHosts)))
}

func startHost(host string) {
	trimmedHost := strings.TrimSpace(host)

	// If it doesn't contain port :22 add it
	if !strings.Contains(trimmedHost, ":") {
		trimmedHost = fmt.Sprintf("%s:%d", trimmedHost, port)
	}

	// Ignore what you can't parse as host:port
	_, _, err := net.SplitHostPort(trimmedHost)
	if err != nil {
		log.Printf("Couldn't parse: %s", trimmedHost)
		return
	}

	//hostSet.Add(trimmedHost)

	// Finally queue it up for processing
	hostQueue <- trimmedHost
	hostWg.Add(1)
}

func consumeAndLimitConcurrency(sshConfig *ssh.ClientConfig, command string) {
	for host := range hostQueue {
		concurrencySem <- 1
		go func(h string) {
			defer func() {
				<-concurrencySem
				hostWg.Done()
			}()
			doSSH(sshConfig, h, command)
		}(host)
	}
}

func doSSH(sshConfig *ssh.ClientConfig, hostname string, command string) {
	var finalError error
	defer func() {
		if finalError != nil {
			log.Println(color.YellowString(hostname) + fmt.Sprintf(" error %s", finalError.Error()))
			//hostErrorSet.Add(hostname)
		}
	}()

	client, err := ssh.Dial("tcp", hostname, sshConfig)
	if err != nil {
		finalError = fmt.Errorf("Failed to dial remote host: %s", err.Error())
		return
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		finalError = fmt.Errorf("Failed to create session: %s", err.Error())
		return
	}
	defer session.Close()

	out, err := session.StdoutPipe()
	if err != nil {
		log.Fatal("Couldn't create pipe to session stdout... :(")
	}
	scanner := bufio.NewScanner(out)

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	if err := session.Run(command); err != nil {
		log.Printf("Failed to run command: %s", err.Error())
	}

	currentHost := strings.Split(hostname, ":")[0]
	logHost := color.GreenString(currentHost)

	for scanner.Scan() {
		fmt.Println(logHost + " " + scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Print(color.RedString(currentHost) + ": Error reading output from this host.")
	}
}
