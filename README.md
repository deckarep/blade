Bladerunner
===========

Bladerunner is an SSH based remote command runner tool that attempts to capture best-practices when
managing remote infrastructure inside TOML files. 

The design goal of Bladerunner is that recipes can be created which describe one or more commands to be executed on one more remote hosts.

Commands are defined in a recipe folder and the intent is that recipes can be shared amongst your
team. A recipe ideally captures the best-practice around running a remote command on one or more
servers. Recipes are placed in a folder hiearchy that you define which best reflects your command hiearchy when executing commands.

### Why a new tool?
* I've been wanting to build a tool like this in Go for awhile thanks to Go's awesome concurrency
primitives and great SSH tooling.
* I want to capture and document the best practice around running remote commands on my infrastructure. In other words, I want to make sure that when I run commands, they're documented well, they are safe, they restrict concurrency to the right amount, they prompt when necessary and they can be shared so the next team-mate can do it the same.
* I want to leverage something that is high-performance and very light-weight when it comes to threading and handling of many TCP connections.

### Tutorial

Here is the most basic recipe which consists two *required* fields: *Commands* and *Hosts*. Both of which are defined as Toml lists within a toml file which means you can execute multiple commands on multiple hosts if you so desire. As it stands, the commands will be executed in serial on each host specified. And with no concurrency limit defined; this recipe will be applied to one host at a time until all commands have completed on all hosts.

```toml
[Required]
  Commands = [
    "hostname"
  ]
  Hosts = ["blade-prod"]
```

Bladerunner reads recipes defined in a recipe folder somewhere on your file-system. The intention being that the recipes are just data defined in Toml based files. The folder structure you use has implications on how Bladerunner interpresets your command hiearchy. Let's place the file above in the following folder hiearchy: `recipes/infra-a/hostname.blade.toml`.

```
.
└── infra-a
    └── hostname.blade.toml
```

In the directory structure defined above Bladerunner will now recognize that you have a command located within the `recipes` folder. Bladerunner does not care about your folder structure but you should care about it. Because not only does it give you the opportunity to organize your Bladerunner Toml files. But Bladerunner will also create a command hiearchy based on this folder structure like so:

```sh
./blade run

# Output below:
run [command]

Usage:
  blade run [command]

Available Commands:
  infra-a

Flags:
  -c, --concurrency int   Max concurrency when running ssh commands
  -h, --help              help for run
  -p, --port int          The ssh port to use (default 22)
  -q, --quiet             quiet mode will keep Blade as silent as possible.
  -r, --retries int       Number of times to retry until a successful command returns (default 3)
  -s, --servers string    servers flag is one or more comma-delimited servers.
  -u, --user string       user for ssh host login. (default "root")
  -v, --verbose           verbose mode will keep Blade as verbose as possible.
```

Notice that Bladerunner has a `run` command. This is the primary entry point into executing commands. But additionally, Bladerunner has a subcommand named `infra-a`.  Let's run that subcommand now.

```sh
./blade run infra-a

# Output below:
./blade run infra-a
Usage:
  blade run infra-a [command]

Available Commands:
  hostname

Flags:
  ...
```

You'll notice that Bladerunner dumps another help synopsis showing that a single available command exists named `hostname`. 

At which now you can run like so:

```sh
./blade run infra-a hostname

# Output below:
2017/09/02 13:35:34 Starting recipe: infra-a.hostname
blade-prod: blade2
2017/09/02 13:35:35 Completed recipe: infra-a.hostname - 1 sucess | 0 failed | 1 total
```

At this point you've observed that a series of subcommands were dynamically added to Bladerunner based on your folder hiarchy and your defined TOML commands.  The folders allow you to organize commands into a hiearchy that reflects your ideal infrastructure. Folders although subcommands, are not executable themselves but simply a means of giving you the ability to build a smart command hiearchy that is intuitive and easy to remember.



### Features
* Bladeruuner is incredibly light-weight: 1 goroutine per ssh connection vs 1 os thread per ssh connection.
* Enforces a consistent style for common tasks ie: we all deploy the same.
* Consists of components which are single one-off commands
* And Recipes which are composed commands to enforce better and consistent administration across the organization
* Enforces proper concurrency restrictions when running remote commands.
* Colorized output for easier groking.
* Automatically ensures all commands run properly and possibly retried.
* TODO: Recipes of Recipes, recipes are composeable.
* TODO: Summaries for when you don't want to see a bunch git-hashes streaming by, just tell me if everything matches please.
* TODO: Allows user-specific recipe overrides.
* TODO: Caches host lookup queries for faster execution (configurable).
* TODO: Built-in safety for destructive commands.

### Possible Future Features
* Command locks, is someone already running this remote command?  Let's not step on each other...yours will have to wait.