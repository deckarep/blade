package ssh

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/deckarep/blade/lib/recipe"
	humanize "github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

var (
	concurrencySem        chan int
	hostQueue             = make(chan string)
	hostWg                sync.WaitGroup
	successfullyCompleted int32
	failedCompleted       int32
	sessionLogger         = log.New(os.Stdout, "", 0)
)

// StartSSHSession kicks off a session of work.
func StartSSHSession(recipe *recipe.BladeRecipe) {
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

	sshCmds := recipe.Required.Commands

	// Concurrency must be at least 1 to make progress.
	if recipe.Overrides.Concurrency == 0 {
		recipe.Overrides.Concurrency = 1
	}

	concurrencySem = make(chan int, recipe.Overrides.Concurrency)
	go consumeAndLimitConcurrency(sshConfig, sshCmds)

	// If Hosts is defined, just use that as a discrete list.
	allHosts := recipe.Required.Hosts

	if len(allHosts) == 0 {
		// Otherwise do dynamic host lookup here.
		commandSlice := strings.Split(recipe.Required.HostLookupCommand, " ")
		out, err := exec.Command(commandSlice[0], commandSlice[1:]...).Output()
		if err != nil {
			fmt.Println("Couldn't execute command:", err.Error())
			return
		}

		allHosts = strings.Split(string(out), ",")
	}

	log.Print(color.GreenString(fmt.Sprintf("Starting recipe: %s", recipe.Meta.Name)))

	totalHosts := len(allHosts)
	for _, h := range allHosts {
		startHost(h, recipe.Overrides.Port)
	}

	hostWg.Wait()
	log.Print(color.GreenString(fmt.Sprintf("Completed recipe: %s - %d sucess | %d failed | %d total",
		recipe.Meta.Name,
		atomic.LoadInt32(&successfullyCompleted),
		atomic.LoadInt32(&failedCompleted),
		totalHosts)))
}

func startHost(host string, port int) {
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

	// Finally queue it up for processing
	hostQueue <- trimmedHost
	hostWg.Add(1)
}

func consumeAndLimitConcurrency(sshConfig *ssh.ClientConfig, command []string) {
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

func doSSH(sshConfig *ssh.ClientConfig, hostname string, commands []string) {
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

	// Since we can run multiple commands, we need to keep track of intermediate failures
	// and log accordingly or do some type of aggregate report.
	for i, cmd := range commands {
		doSingleSSHCommand(i+1, client, hostname, cmd)
	}

	// Technically this is only successful when errors didn't occur above.
	atomic.AddInt32(&successfullyCompleted, 1)
}

// doSingleSSHCommand - the rule is one command can only ever occur per session.
func doSingleSSHCommand(index int, client *ssh.Client, hostname, command string) {
	var finalError error
	defer func() {
		if finalError != nil {
			sessionLogger.Println(color.YellowString(hostname) + fmt.Sprintf(" error %s", finalError.Error()))
		}
	}()

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
		finalError = fmt.Errorf("Couldn't create pipe to Stdout for session: %s", err.Error())
		return
	}

	errOut, err := session.StderrPipe()
	if err != nil {
		finalError = fmt.Errorf("Couldn't create pipe to Stderr for session: %s", err.Error())
		return
	}

	currentHost := strings.Split(hostname, ":")[0]

	// Consume session Stdout, Stderr pipe async.
	go consumeReaderPipes(currentHost, out)
	go consumeReaderPipes(currentHost, errOut)

	// Once a Session is created, you can only ever execute a single command.
	if err := session.Run(command); err != nil {
		sessionLogger.Print(color.RedString(currentHost+":") + fmt.Sprintf(" Failed to run the %s command: `%s` - %s", humanize.Ordinal(index), command, err.Error()))
	}
}

func consumeReaderPipes(host string, rdr io.Reader) {
	logHost := color.CyanString(host + ":")

	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		sessionLogger.Println(logHost + " " + scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		sessionLogger.Print(color.RedString(host) + ": Error reading output from this host.")
	}
}
