Blade
======

Administer servers the Deckarep's  way: Blade is bigger, better, faster than Knife.

### Features
* Blade eats ssh connections for breakfast: Translation 1 goroutine per ssh connection vs 1 os thread per ssh connection.
* Caches knife queries for faster execution (configurable)
* Enforces a consistent style for common tasks
* Built-in safety for destructive commands
* Consists of components which are single one-off commands
* And Recipes which are composed commands to enforce better and consistent administration across the organization
* Summaries: for when you don't want to see a bunch git-hashes streaming by, just tell me if everything matches please.
* Enforces proper rolling for deploying, in other words: don't deploy all proxies in one shot please...
* Colorized output for easier groking
* Automatically ensures all commands run properly: asserts exit(0)
* Allows for custom user-specific components and recipes
* Deploy locks, is someone else already deploying?  Let's not step on each other...yours will have to wait.


### Why a new tool?
Knife is cool for what it does but it's written in Ruby and actually spins up a full OS dedicated thread per connection.
Additionally everyone deploys/manages a little differently when doing common tasks like deploying, etc yet we should
ALL be doing it the same to ensure proper state and consistency.
