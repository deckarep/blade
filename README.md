Blade
=====

NOTE: Blade is an unstable alpha API -- constructive feedback welcome.

Blade is an SSH based remote command runner tool that attempts to capture best-practices when
managing remote infrastructure inside .yaml files. These .yaml files are meant to be under source control and shared with team-mates. A Blade `.blade.yaml` file holds the declarative instructions for running one or more remote commands on one or more servers.

This gives you the power of running well-defined commands on fleets of servers in a simple, expressive and easy to use CLI interface. Add some `.blade.yaml` recipe commands to your `recipe/` folder in a hierarchy that makes sense to you and Blade will build a slick CLI dynamically for you based on this folder structure. Then you're ready to start using the tool.

### Demo
TODO

### Tutorial

In this tutorial, we're going to simulate creating a very basic command that we want to run on a infrastructure named: `tutorial`. Blade doesn't care how you organize your folder hierarchy but you should model your folder hierachy based on the command hierarchy that you makes sense to you and your organization.

In the `recipes/tutorial/` folder create this file and name it: `hostname.blade.yaml`. This file has a single command that will be run on a single host.

```yaml
hosts: ["blade-dev"]
exec:
  - hostname
```

Place the file above in the following folder hierarchy.

```
recipes
└── tutorial
    └── hostname.blade.yaml
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
  tutorial

Flags:
  -h, --help              help for run
  ...
```

Notice that Blade's `run` command is aware of the `tutorial` folder. The `run` command is the primary entry point into executing .yaml files.

Now run the `tutorial` command and notice that you now have a `hostname` command available.

```sh
./blade run tutorial
```

```
Usage:
  blade run tutorial [command]

Available Commands:
  hostname

Flags:
  -h, --help   help for tutorial
```

The idea here is that a folder represents a command of which sub-commands can be executed. This implies that "folder" based command does nothing more than group sub-commands of which *can* be executed like so:

```sh
./blade run tutorial hostname
```

```
# Output below:
2018/02/15 22:15:44 Starting recipe: tutorial.hostname.blade.yaml
blade-dev: blade-dev
2018/02/15 22:15:46 Completed recipe: tutorial.hostname.blade.yaml - 1 success | 0 failed | 1 total
```

At this point, we've executed a single remote command called `hostname` on a single remote host called `blade-dev`. 
Let's modify our single `hostname.blade.yaml` file to run on more hosts.

```yaml
hosts: ["blade-dev", "blade-prod"]
exec:
  - hostname
```

Here we've defined another host we have access to and we can now rerun our command:

```sh
./blade run tutorial hostname
```

```
# Output below:
2018/02/15 22:19:14 Starting recipe: tutorial.hostname.blade.yaml
blade-dev: blade-dev
blade-prod: blade-prod
2018/02/15 22:19:17 Completed recipe: tutorial.hostname.blade.yaml - 2 success | 0 failed | 2 total
```

As you can see, Blade has now executed a single command on each remote host. This execution happened in a serial fashion where only a single host was executed at a time.

Let's modify the `hostname.blade.yaml` to execute an additional command per host and save that change.

```yaml
hosts: ["blade-dev", "blade-prod"]
exec:
  - hostname
  - sleep 5
```

Rerun the command: `./blade run tutorial hostname` and observe that for each host running there is a 5 second delay due to the sleep command. This means, that because we execute these commands in serial on one host first, then the other Blade will take a total of 10 seconds to complete for both hosts.

But, with the power of concurrency, we can update our `hostname.blade.yaml` file to have our commands executed at a concurrency level of 2. Let's also add a third `echo` command so we can observe how this changes the behavior of our run.

```yaml
hosts: ["blade-dev", "blade-prod"]
exec:
  - echo "starting `hostname`"
  - sleep 5
overrides:
  concurrency: 2
```

Rerun the command: `/.blade run tutorial hostname` and now observe that because we added a concurrency override of 2 and even though we have a sleep delay of 5 seconds, both servers start and execute these remote commands and the entire Blade session finishes in about 5 seconds.

Instead of updating our `hostname.blade.yaml` file we additionally could have used Blade's command-line flags to override the concurrency behavior like so:

```sh
./blade run tutorial hostname -c2 # or --concurrency 2
```

This effectively acheives the same thing but instead controls the concurrency amount via the usage of an ad-hoc command line flag.

### Features
* Blade is incredibly light-weight: 1 goroutine per ssh connection vs 1 os thread per ssh connection.
* Recipes are composed commands to enforce better and consistent administration across an organization.
* Enforces proper concurrency restrictions when running remote commands.
* Colorized output for easier groking.
* Automatically ensures all commands run successfully with optional retry.
* TODO: Recipes of Recipes, recipes are composable.
* TODO: Summaries for when you don't want to see a bunch git-hashes streaming by, just tell me if everything matches please.
* TODO: Allows user-specific recipe overrides.
* TODO: Caches host lookup queries for faster execution (configurable).
* TODO: Built-in safety for destructive commands.

### Possible Future Features
* Command locks, is someone already running this remote command?  Let's not step on each other...yours will have to wait.
