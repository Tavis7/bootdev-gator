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

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("No arguments expected")
	}

	feedURL := "https://www.wagslane.dev/index.xml"

	feed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", feed)

	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Exactly two arguments expected")
	}

	feedName := cmd.args[0]
	feedURL := cmd.args[1]

	username := s.config.Current_user_name

	user, err := s.database.GetUser(context.Background(), username)
	if err != nil {
		return err
	}

	fmt.Printf("Adding feed %v @ %v for %v\n", feedName, feedURL, username)

	now := time.Now().UTC()
	feed, err := s.database.CreateFeed(context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			Name:      feedName,
			Url:       feedURL,
			UserID:    user.ID,
		})
	if err != nil {
		return err
	}

	fmt.Printf("Added feed %v\n", feed)

	return helperFollow(s, feedURL)
}

func handlerListFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("No arguments expected")
	}

	feeds, err := s.database.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf(`"%v": %v (created by %v)`+"\n", feed.Name, feed.Url, feed.Username.String)
	}

	return nil
}

func helperFollow(s *state, feedURL string) error {
	username := s.config.Current_user_name

	user, err := s.database.GetUser(context.Background(), username)
	if err != nil {
		return err
	}

	feed, err := s.database.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return err
	}

	fmt.Printf("Following feed %v @ %v for %v\n", feed.Name, feedURL, username)

	now := time.Now().UTC()
	followed, err := s.database.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			UserID:    user.ID,
			FeedID:    feed.ID,
		})
	if err != nil {
		return err
	}

	fmt.Printf("%#v\n", followed)

	return nil
}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Exactly one argument expected")
	}

	feedURL := cmd.args[0]

	return helperFollow(s, feedURL)
}

func handlerFollowing(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("No arguments expected")
	}

	username := s.config.Current_user_name
	feeds, err := s.database.GetFeedFollowsForUser(context.Background(), username)
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("%v] %v @ %v\n", feed.Username, feed.Feedname, feed.FeedUrl)
	}

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
