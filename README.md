# gator

## Prerequisites and Installation

This project requires Postgres and Go installed to run this program.

You can install gator CLI using `go install`:

    go install github.com/mhiillos/gator@latest

You also need to run a postgres database locally, you can easily create that with:

```bash
# Install PostgreSQL if you haven't already
# For Ubuntu/Debian:
sudo apt update
sudo apt install postgresql postgresql-contrib

# For macOS (using Homebrew):
brew install postgresql
brew services start postgresql

# For Windows:
# Download and install from https://www.postgresql.org/download/windows/

# Create a database for gator
psql -U postgres -c "CREATE DATABASE gator;"

# Install Goose migration tool
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations from the project directory
goose -dir sql/schema postgres "postgres://postgres:password@localhost:5432/gator?sslmode=disable" up
```


## Usage

gator reads from a config file `gatorconfig.json`, which should be in your home directory.
It should be set up as follows:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
}
```
where `db_url` is the address of your database.

gator can then be used by supplying a command and optional arguments:

    gator <command> [arguments]

Available commands:

```
register <username>:   Registers the specified username to the database
login <username>:      Logs the specified user in
reset:                 Removes all the users from the database
users:                 Prints all the users in the database
addfeed <name> <URL>:  Adds an RSS feed to the database
feeds:                 Prints all feeds in the database
follow <URL>:          Follow an RSS feed
unfollow <URL>:        Unfollow an RSS feed
following:             Lists the RSS feeds you are following
agg <duration_string>: Collects the RSS feeds at the specified interval from followed feeds
browse <limit>:        Outputs information of the latest feeds specified by the limit, defaults to two recent feeds
```

The intended use for this CLI tool is to run the `agg` command at given intervals (E.g. `gator agg 1m`), while using another terminal window to see the results.
