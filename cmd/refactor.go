package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/deckarep/golang-set"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Usage: go run main.go --help (for help)
// Usage: go run main.go -k "chef_environment:production* AND role:filter"  -m 'sudo which jq || sudo yum -y install jq' -C 5
const (
	queryBucket = "queries"
)

// TODO: move ssh crap into it's own package
var sshConfig *ssh.ClientConfig
var hostSet mapset.Set
var hostErrorSet mapset.Set
var queue chan string
var wg sync.WaitGroup
var db *bolt.DB
var userKnifeQuery string
var userKnifeCommand string
var sshUserName string
var sem chan int

// Flags
var knifeExpression string
var commandExpression string
var concurrencyLevel int

// Create the config once from the SSH-AGENT
func init() {
	parseFlags()

	// Where to place items
	queue = make(chan string)

	// Takes care of storing currently in progress host names
	hostSet = mapset.NewSet()

	// Deals with all the hosts that are either timing error or not dialing
	hostErrorSet = mapset.NewSet()

	// User's ssh config
	sshUserName, err := loadSSHUserName()
	if err != nil {
		log.Fatal(err.Error())
	}
	sshConfig = &ssh.ClientConfig{
		User: sshUserName,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
	}

	userKnifeQuery = knifeExpression
	userKnifeCommand = commandExpression

	sem = make(chan int, concurrencyLevel)
}

var RootCmd = &cobra.Command{
	Use:   "blade",
	Short: "do blade stuff",
	Run: func(cmd *cobra.Command, args []string) {
		startRootCmd()
	},
}

func startRootCmd() {
	defer un(trace("main"))

	openDB()
	defer db.Close()

	// Kick off enqueue in a goroutine
	go consume()

	// Start getting data
	output := knifeSearch(userKnifeQuery)
	scanHosts(strings.NewReader(output))

	// Wait for all hosts
	wg.Wait()

	// Remove bad hosts from cash? (TODO: prompt user)
	removeBadHosts()

	log.Println("Finished.")
}

func parseFlags() {
	flag.IntVar(&concurrencyLevel, "C", 5, "Concurrency level for remote ssh commands")
	flag.StringVar(&knifeExpression, "k", "", "Knife expression like: 'chef_environment:staging_* AND role:apid'")
	flag.StringVar(&commandExpression, "m", "hostname", "Command like: 'hostname'")

	// Parse 'em
	flag.Parse()
}

func loadSSHUserName() (string, error) {
	sshFilePath := path.Join(os.Getenv("HOME"), ".ssh", "config")
	sshConfig, err := ioutil.ReadFile(sshFilePath)
	if err != nil {
		return "", errors.New("could not load your SSH config.")
	}

	re, _ := regexp.Compile(`User (\w+)`)
	matches := re.FindStringSubmatch(string(sshConfig))
	if len(matches) > 1 {
		return matches[1], nil
	} else {
		return "", errors.New("could not find your SSH username. is it defined in ~/.ssh/config?")
	}
}

func trace(s string) (string, time.Time) {
	log.Println("START:", s)
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("  END:", s, "ElapsedTime in seconds:", endTime.Sub(startTime))
}

// Idea make this a running agent?
// Agent is always refreshing in the background your most common queries?
// It can update every 5 minutes and on failures will just log the error.

func storeBolt(key, value string) {
	defer un(trace("storeBolt"))
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(queryBucket))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(key), []byte(value))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func getBolt(key string) string {
	defer un(trace("getBolt"))
	result := ""
	// retrieve the data
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(queryBucket))
		if bucket == nil {
			// Bucket not created yet
			return nil
		}

		val := bucket.Get([]byte(key))
		result = string(val)

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return result
}

func openDB() {
	defer un(trace("openDB"))

	var err error
	db, err = bolt.Open("blade-boltdb.db", 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func removeBadHosts() {
	if hostErrorSet.Cardinality() > 0 {
		successHosts := hostSet.Difference(hostErrorSet)
		l := make([]string, 0)
		s := successHosts.ToSlice()
		for _, h := range s {
			l = append(l, h.(string))
		}

		storeBolt(userKnifeQuery, strings.Join(l, "\n"))
	}
}

func scanHosts(r io.Reader) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		enqueueHost(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func knifeSearch(query string) string {
	// Use the cached query if possible
	if cachedQuery := getBolt(query); cachedQuery != "" {
		log.Println(color.MagentaString("Using cached query..."))
		return cachedQuery
	}

	log.Println(color.MagentaString("Performing knife lookup..."))
	out, err := exec.Command("knife", "search", "node", query, "-i").Output()
	if err != nil {
		log.Fatal(err)
	}
	results := string(out)
	storeBolt(query, results)
	return results
}

func endOnSignal() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	<-done
}

// Queries the actual ssh-agent
func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

func enqueueHost(host string) {
	trimmedHost := strings.TrimSpace(host)

	// Ignore commented out lines or empty lines
	if strings.HasPrefix(trimmedHost, "#") || trimmedHost == "" {
		return
	}

	// If it doesn't contain port :22 add it
	if !strings.Contains(trimmedHost, ":22") {
		trimmedHost = trimmedHost + ":22"
	}

	// Ignore what you can't parse as host:port
	_, _, err := net.SplitHostPort(trimmedHost)
	if err != nil {
		log.Printf("Couldn't parse: %s", trimmedHost)
		return
	}

	hostSet.Add(trimmedHost)

	// Finally queue it up for processing
	queue <- trimmedHost
	wg.Add(1)
}

func consume() {
	for host := range queue {
		sem <- 1
		//log.Println(color.CyanString(host) + " is processing command...")

		// Launch a goroutine to fetch the URL.
		go func(h string) {

			// When complete release semaphore and mark as done
			defer func() {
				<-sem
				wg.Done()
			}()

			// Process command
			hitServer(h)
		}(host)
	}
}

// TODO: Don't Fatal in this block OR, catch the panic
func hitServer(hostname string) {
	var finalError error
	defer func() {
		if finalError != nil {
			log.Println(color.YellowString(hostname) + fmt.Sprintf(" error %s", finalError.Error()))
			hostErrorSet.Add(hostname)
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

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var outputBuffer bytes.Buffer
	session.Stdout = &outputBuffer

	// if err := session.Run("cd /opt/company/serverd/current && git rev-parse HEAD"); err != nil {
	// 	log.Printf("Failed to run command: %s", err.Error())
	// }

	if err := session.Run(userKnifeCommand); err != nil {
		log.Printf("Failed to run command: %s", err.Error())
	}

	// Hostname
	log.Print(color.GreenString(hostname) + " " + outputBuffer.String())
}
