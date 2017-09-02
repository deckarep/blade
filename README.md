Bladerunner
===========

NOTE: Bladerunner is an unstable alpha API -- constructive feedback welcome.

Bladerunner is an SSH based remote command runner tool that attempts to capture best-practices when
managing remote infrastructure inside TOML files. These TOML files are meant to be under source control and shared with team-mates. A Bladerunner .toml file holds the declarative instructions for running one or more remote commands on one or more servers. 

This gives you the power of running well-defined commands on fleets of servers in a simple, expressive and easy to use CLI interface.

### Demo
[![asciicast](https://asciinema.org/a/LrTB1qOLYyRCVw7S4yEHnl950.png)](https://asciinema.org/a/LrTB1qOLYyRCVw7S4yEHnl950)

### Tutorial

In the `recipes/infra-a/` folder create this file and name it: `hostname.blade.toml`. This file has a single command that will be run on a single host.

```toml
[Required]
  Commands = [
    "hostname"
  ]
  Hosts = ["blade-prod"]
```

Place the file above in the following folder hierarchy.

```
recipes
└── infra-a
    └── hostname.blade.toml
```

Run the following command:

```sh
./blade run
```

```
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

At this point you've observed that a series of subcommands were dynamically added to Bladerunner based on your folder hierarchy and your defined TOML commands.  The folders allow you to organize commands into a hierarchy that reflects your ideal infrastructure. Folders although subcommands, are not executable themselves but simply a means of giving you the ability to build a smart command hierarchy that is intuitive and easy to remember.

At this point, we've executed a single remote command called `hostname` on a single remote host called `blade-prod`. `blade-prod` is a remote server that I've set up on Vultr for the purpose of building this tool but it ultimately could be any server that you have access to where your SSH public-key is configured.

Let's modify our single `hostname.blade.toml` file to run on more hosts.

```toml
[Required]
  Commands = [
    "hostname"
  ]
  Hosts = ["blade-prod", "blade-prod-a"]
```

Here we've defined another host we have access to and we can now rerun our command:

```sh
./blade run infra-a hostname

# Output below:
2017/09/02 13:47:50 Starting recipe: infra-a.hostname
blade-prod: blade2
blade-prod-a: blade2
2017/09/02 13:47:52 Completed recipe: infra-a.hostname - 2 sucess | 0 failed | 2 total
```

As you can see, Bladerunner now executed a single command on each remote host defined. This execution happened in a serial fashion where only a single host was executed at a time. Note: The reason you see the same output is because I currently have my `/etc/hosts` file modified to have multiple aliases pointing to the same server instance.

Let's modify the `hostname.blade.toml` to execute an additional command per host and save that change.

```toml
[Required]
  Commands = [
    "sleep 5",
    "hostname"
  ]
  Hosts = ["blade-prod", "blade-prod-a"]
```

Rerun the command: `./blade run infra-a hostname` and observe that for each host running there is a 5 second delay due to the first sleep command. This means, that because we execute these commands in serial on one host first, then the other Bladerunner will take a total of 10 seconds to complete.

But, with the power of concurrency, we can update our `hostname.blade.toml` file to have our commands executed at a concurrency level of 2. Let's also add a third `echo` command so we can observe how this changes the behavior of our run.

```toml
[Required]
  Commands = [
    "echo 'before sleep'",
    "sleep 5",
    "hostname"
  ]
  Hosts = ["blade-prod", "blade-prod-a"]

[Overrides] 
  Concurrency = 2
```

Rerun the command: `/.blade run infra-a hostname` and now observe that because we added a concurrency override of 2 that although we have a sleep delay of 5 seconds, both servers start and execute these remote commands and the entire Bladerunner session finishes in about 5 seconds.

Instead of updating our `hostname.blade.toml` file we additionally could have used Bladerunner's command-line flags to override the concurrency behavior like so:

```sh
./blade run infra-a hostname -c1 # or --concurrency 1
```

This effectively acheives the same thing but instead controls the concurrency amount via the usage of an ad-hoc command line.

What if we wanted to introduce some additional command line flags to our commands to dynamically change their behavior in an ad-hoc fashion before execution? We can do this by introducing `Argument Sets` as in the following:

```toml
[Argument]
  [[Argument.Set]]
    Arg = "name"
    Value = "Bob"
    Help = "is the name to echo"
    
[Required]
  Commands = [
    "echo '{{name}}'",
    "hostname"
  ]
  Hosts = ["blade-prod", "blade-prod-a"]

[Overrides] 
  Concurrency = 2
```

Now at the command prompt if we type the following: `./blade run infra-a hostname --help`. We can see that we have introduced a new recipe flag called: `--name`:

```sh
Usage:
  blade run infra-a hostname [flags]

Flags:
  -h, --help          help for hostname
      --name string   is the name to echo (recipe flag)
...
```

Therefore executing the `hostname.blade.toml` file without a flag will simply perform the `echo` command with the name `Bob`.

And overriding this command is simply a matter of: `./blade run infra-a hostname --name Deckarep`.

And the output now looks like the following:

```sh
2017/09/02 14:07:27 Starting recipe: infra-a.hostname
blade-prod-a: Jerry
blade-prod: Jerry
blade-prod-a: blade2
blade-prod: blade2
2017/09/02 14:07:28 Completed recipe: infra-a.hostname - 2 sucess | 0 failed | 2 total
```


### Features
* Bladeruuner is incredibly light-weight: 1 goroutine per ssh connection vs 1 os thread per ssh connection.
* And Recipes which are composed commands to enforce better and consistent administration across the organization
* Enforces proper concurrency restrictions when running remote commands.
* Colorized output for easier groking.
* Automatically ensures all commands run properly and possibly retried.
* TODO: Recipes of Recipes, recipes are composable.
* TODO: Summaries for when you don't want to see a bunch git-hashes streaming by, just tell me if everything matches please.
* TODO: Allows user-specific recipe overrides.
* TODO: Caches host lookup queries for faster execution (configurable).
* TODO: Built-in safety for destructive commands.

### Possible Future Features
* Command locks, is someone already running this remote command?  Let's not step on each other...yours will have to wait.
