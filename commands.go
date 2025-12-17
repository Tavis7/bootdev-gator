package main

import (
	"context"
	"fmt"
	"github.com/Tavis7/bootdev-gator/internal/database"
	"github.com/google/uuid"
	"time"
)

type command struct {
	name string
	args []string
}

type commands struct {
	commandList map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("username expected")
	}

	username := cmd.args[0]

	user, err := s.database.GetUser(context.Background(), username)
	if err != nil {
		return err
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User has been set to '%v'\n", user.Name)
	fmt.Printf("'%#v'\n", user)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Username expected")
	}
	username := cmd.args[0]
	fmt.Printf("Registering %v\n", username)

	now := time.Now().UTC()
	user, err := s.database.CreateUser(context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Name:      username,
		})
	if err != nil {
		return err
	}

	s.config.SetUser(username)

	fmt.Printf("User %v was created: %v\n", username, user)

	return nil
}

func handlerUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("No arguments expected")
	}

	users, err := s.database.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		current := ""
		if user == s.config.Current_user_name {
			current = " (current)"
		}
		fmt.Printf("%v%v\n", user, current)
	}

	return nil
}

func handlerResetUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("No arguments expected")
	}

	err := s.database.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}

	fmt.Printf("Deleted all users\n")

	return nil
}

func handlerHelp(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("No arguments expected")
	}

	fmt.Printf("Available commands:\n")
	for k := range s.commands.commandList {
		fmt.Printf("    %v\n", k)
	}

	return nil
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.commandList[cmd.name]
	if !ok {
		return fmt.Errorf("Command '%v' does not exist", cmd.name)
	}

	return f(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandList[name] = f
}
