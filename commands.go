package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/Tavis7/bootdev-gator/internal/database"
)

type command struct {
	name string
	args []string
}

type commandDoc struct {
	name string
	args string
	doc  string
}

type commands struct {
	commandList         map[string]func(*state, command) error
	commandDocs         []commandDoc
	maxCommandArgLength int
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
	if len(cmd.args) != 1 {
		return fmt.Errorf("Exactly one argument expected")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		fmt.Printf("Scraping feed\n")
		err = scrapeFeeds(s)
		if err != nil {
			return err
		}
	}

	feedURL := "https://www.wagslane.dev/index.xml"

	feed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", feed)

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Exactly two arguments expected")
	}

	feedName := cmd.args[0]
	feedURL := cmd.args[1]

	fmt.Printf("Adding feed %v @ %v for %v\n", feedName, feedURL, user.Name)

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

	return helperFollow(s, feed.ID, user.ID)
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

func helperFollow(s *state, feed uuid.UUID, user uuid.UUID) error {
	now := time.Now().UTC()
	followed, err := s.database.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			UserID:    user,
			FeedID:    feed,
		})
	if err != nil {
		return err
	}

	fmt.Printf(`%v followed "%v"`+"\n", followed.Username.String, followed.Feedname.String)

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Exactly one argument expected")
	}

	feedURL := cmd.args[0]

	feed, err := s.database.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return err
	}

	return helperFollow(s, feed.ID, user.ID)
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("No arguments expected")
	}

	feeds, err := s.database.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf(`%v is following "%v" @ %v`+"\n", feed.Username, feed.Feedname, feed.FeedUrl)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Exactly one argument expected")
	}

	feedURL := cmd.args[0]

	feed, err := s.database.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return err
	}

	_, err = s.database.DeleteFeedFollow(context.Background(),
		database.DeleteFeedFollowParams{
			UserID: user.ID,
			FeedID: feed.ID,
		})
	if err != nil {
		return err
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 1 {
		return fmt.Errorf("Only one argument expected")
	}

	limit := int32(2)
	if len(cmd.args) == 1 {
		l, err := strconv.ParseInt(cmd.args[0], 10, 32)
		if err != nil {
			return err
		}
		limit = int32(l)
	}

	feed, err := s.database.GetPostsForUser(context.Background(),
		database.GetPostsForUserParams{
			ID:    user.ID,
			Limit: limit,
		})
	if err != nil {
		return err
	}

	for _, item := range feed {
		fmt.Printf(`[%v] "%v": "%v"`+"\n    %v\n",
			item.PublishedAt, item.FeedName, item.Title.String, item.Url)
	}

	return nil
}

func handlerHelp(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("No arguments expected")
	}

	fmt.Printf("Available commands:\n")
	for _, doc := range s.commands.commandDocs {
		docstring := doc.name
		if len(doc.args) > 0 {
			docstring += " " + doc.args
		}
		length := len(docstring)
		fmt.Printf("    %v: %v%v\n", docstring,
			strings.Repeat(" ", max(s.commands.maxCommandArgLength-length+1, 0)), doc.doc)
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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	return func(s *state, cmd command) error {
		user, err := s.database.GetUser(context.Background(), s.config.Current_user_name)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func parseDate(dateString string) (time.Time, error) {
	day := ""
	year := "06"

	dateString = strings.Trim(dateString, " ")
	parts := strings.Fields(dateString)
	offset := 0
	if len(parts[0]) > 2 {
		day = "Mon, "
		offset = 1
	}
	if len(parts[2+offset]) > 2 {
		year = "2006"
	}

	layout := day + "02 Jan " + year + " 15:04:05 -0700"
	date, err := time.Parse(layout, dateString)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

func createPost(s *state, item RSSItem, feedID uuid.UUID) error {
	// fmt.Printf("      Creating post...\n")
	now := time.Now().UTC()
	publishedAt, err := parseDate(item.PubDate)
	if err != nil {
		return err
	}

	_, err = s.database.CreatePost(context.Background(),
		database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   now,
			UpdatedAt:   now,
			Title:       sql.NullString{String: item.Title, Valid: true},
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: publishedAt,
			FeedID:      feedID,
		})
	if err != nil {
		perr, ok := err.(*pq.Error)
		if !ok ||
			(perr.Code.Name() != "unique_violation") ||
			(perr.Constraint != "posts_url_key") {
			return err
		}
	}

	return nil
}

func scrapeFeeds(s *state) error {
	dbFeed, err := s.database.GetStalestFeed(context.Background())
	if err != nil {
		return fmt.Errorf("Getting stale feed: %w", err)
	}

	now := time.Now().UTC()
	_, err = s.database.MarkFeedFetched(context.Background(),
		database.MarkFeedFetchedParams{
			LastFetchedAt: sql.NullTime{Time: now, Valid: true},
			ID:            dbFeed.ID,
		})
	if err != nil {
		return fmt.Errorf("Marking feed fetched: %w", err)
	}

	feedURL := dbFeed.Url

	feed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("Fetching feed: %w", err)
	}

	fmt.Printf(`Articles from "%v"`+"\n", feed.Channel.Title)
	for _, item := range feed.Channel.Item {
		/*
			if len(item.Title) > 0 {
				fmt.Printf(` - "%v"`+"\n", item.Title)
			} else {
				fmt.Printf(" - %#v\n", item)
			}
		*/
		err := createPost(s, item, dbFeed.ID)
		if err != nil {
			return err
		}
	}
	fmt.Printf("---\n")

	return nil
}

func (c *commands) register(name, args, doc string, f func(*state, command) error) {
	c.commandList[name] = f
	c.commandDocs = append(c.commandDocs, commandDoc{name: name, args: args, doc: doc})
	c.maxCommandArgLength = max(c.maxCommandArgLength, len(name)+len(args))
}
