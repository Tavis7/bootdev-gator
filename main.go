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
		make(map[string]func(*state, command) error),
	}
	commandList := s.commands
	commandList.register("login", handlerLogin)
	commandList.register("register", handlerRegister)
	commandList.register("reset", handlerResetUsers)
	commandList.register("users", handlerUsers)
	commandList.register("agg", handlerAgg)
	commandList.register("addfeed", handlerAddFeed)
	commandList.register("listfeeds", handlerListFeeds)
	commandList.register("help", handlerHelp)

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

	conf, err = config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(conf)
}
