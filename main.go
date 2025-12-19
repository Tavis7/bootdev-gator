package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Tavis7/bootdev-gator/internal/config"
	"github.com/Tavis7/bootdev-gator/internal/database"
)

import _ "github.com/lib/pq"

type state struct {
	config   *config.Config
	database *database.Queries
	commands *commands
}

func main() {
	conf, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s := state{
		config: &conf,
	}

	db, err := sql.Open("postgres", s.config.Db_url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dbQueries := database.New(db)
	s.database = dbQueries

	s.commands = &commands{
		commandList: make(map[string]func(*state, command) error),
		commandDocs: []commandDoc{},
	}
	commandList := s.commands
	commandList.register("login", "<username>",
		"Log in as <username>",
		handlerLogin)
	commandList.register("register", "<username>",
		"Register and log in as <username>",
		handlerRegister)
	commandList.register("reset", "",
		"Delete everything and reset the database",
		handlerResetUsers)
	commandList.register("users", "",
		"List users",
		handlerUsers)
	commandList.register("agg", "<delay>",
		"Refresh one old feed every <delay> seconds, minutes, hours, etc.",
		handlerAgg)
	commandList.register("addfeed", "<name> <url>",
		"Add and follow feed",
		middlewareLoggedIn(handlerAddFeed))
	commandList.register("feeds", "",
		"List feeds",
		handlerListFeeds)
	commandList.register("follow", "<url>",
		"Follow a feed",
		middlewareLoggedIn(handlerFollow))
	commandList.register("following", "",
		"List feeds you are following",
		middlewareLoggedIn(handlerFollowing))
	commandList.register("unfollow", "<url>",
		"Unfollow a feed you are following",
		middlewareLoggedIn(handlerUnfollow))
	commandList.register("browse", "[<limit>]",
		"List the last <limit> posts from feeds you are following, default 2",
		middlewareLoggedIn(handlerBrowse))
	commandList.register("help", "",
		"Print this help and exit",
		handlerHelp)

	args := os.Args
	if len(args) < 2 {
		fmt.Printf("Error: expected at least one command line argument\n")
		os.Exit(1)
	}

	cmd := command{
		name: args[1],
		args: args[2:],
	}

	err = commandList.run(&s, cmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
