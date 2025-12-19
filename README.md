# Requirements

You will need the following to run bootdev-gator:

- [go](https://go.dev/doc/install)
- [postgres](https://www.postgresql.org/download/)

And to make setup more convenient:

- [goose](https://github.com/pressly/goose?tab=readme-ov-file#install)

Once you have them installed, clone this repository, `cd` into it, and run `go
install`

Get postgres set up (good luck), create a database and add its URL to
`~/.gatorconfig.json`, making sure to include `?sslmode=disable` at the end.

These instructions will assume your postgres url is
"postgres://postgres:postgres@localhost:5432/gator".

```json
{
    "db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
}
```

Then up-migrate the database using goose:

```sh
# from the repository's root directory
cd sql/migrations
goose postgres "postgres://postgres:postgres@localhost:5432/gator" up
cd ../..
```

# Run bootdev-gator

To list the available commands, run:

```sh
bootdev-gator help
```

Create a user using `bootdev-gator register <username>`. You can then add RSS
feeds using `bootdev-gator addfeed <name> <url>`.

If a feed is already added, you will need to follow it instead with
`bootdev-gator follow <url>`.

Run `bootdev-gator agg 1m` in the background to refresh one feed per minute.

List the latest items from the feeds you follow with `bootdev-gator browse
<limit>`.

Viewing the content is not supported. You will need to open the listed URL in a
browser for that.

# TODO (realistically maybe never)

- Command to set up config file automatically
- Command to test database connection
- Command to run database migrations
- Default agg delay to something nice, like 10m
- Refresh the stalest n feeds every \<delay\> instead of only the single stalest
