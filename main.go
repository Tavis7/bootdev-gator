package main

import (
	"fmt"
	"os"
	"database/sql"

	"github.com/Tavis7/bootdev-gator/internal/config"
	"github.com/Tavis7/bootdev-gator/internal/database"
)

import _ "github.com/lib/pq"

type state struct {
	config *config.Config
	database *database.Queries
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

	commandList := commands{
		make(map[string]func(*state, command) error),
	}
	commandList.register("login", handlerLogin)
	commandList.register("register", handlerRegister)
	commandList.register("reset", handlerResetUsers)
	commandList.register("users", handlerUsers)

	args := os.Args
	if len(args) < 2 {
		fmt.Printf("Error: expected at least one command line argument\n")
		os.Exit(1)
	}

	cmd := command {
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
