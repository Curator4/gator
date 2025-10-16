package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/curator4/gator/internal/config"
	"github.com/curator4/gator/internal/database"
	"github.com/curator4/gator/internal/rss"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"os"
	"time"
)

// constants
const dbURL = "postgres://curator:solitude4@localhost:5432/gator?sslmode=disable"

// structs
type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return errors.New("command does not exist")
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

// main
func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	s := &state{
		cfg: &cfg,
		db:  database.New(db),
	}

	c := &commands{
		handlers: make(map[string]func(*state, command) error),
	}

	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", handlerAddFeed)
	c.register("feeds", handlerFeeds)
	c.register("follow", handlerFollow)
	c.register("following", handlerFollowing)

	if len(os.Args) < 2 {
		fmt.Printf("needs at least 2 arguments\n")
		os.Exit(1)
	}

	commandName := os.Args[1]
	commandArgs := os.Args[2:]
	cmd := command{
		name: commandName,
		args: commandArgs,
	}

	if err := c.run(s, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// functions
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("login expects a single argument")
	}
	username := cmd.args[0]

	if _, err := s.db.GetUser(context.Background(), username); err != nil {
		return fmt.Errorf("user does not exist: %w", err)
	}

	if err := (*s.cfg).SetUser(username); err != nil {
		return err
	}
	fmt.Printf("user: %s has been logged in\n", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("register expects a single argument")
	}
	username := cmd.args[0]
	if _, err := s.db.GetUser(context.Background(), username); err == nil {
		return errors.New("user already exists")
	}

	current_time := time.Now()
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: current_time,
		UpdatedAt: current_time,
		Name:      username,
	}

	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("user %s was created\n", user.Name)
	fmt.Printf("user data: %+v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("succesfully reset users table\n")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	feed, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return errors.New("addfeed expects exactly 2 arguments, name and url")
	}
	name := cmd.args[0]
	url := cmd.args[1]

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	current_time := time.Now()
	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: current_time,
		UpdatedAt: current_time,
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed_follow_params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: current_time,
		UpdatedAt: current_time,
		UserID:    user.ID,
		FeedID:    params.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}

	_, err = s.db.CreateFeedFollow(context.Background(), feed_follow_params)
	if err != nil {
		return err
	}

	fmt.Printf("new feed:\n %+v \n", feed)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return errors.New("expects no arguments")
	}

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("%+v\n", feed)
	}

	return nil
}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("expects 1 argument")
	}

	url := cmd.args[0]
	current_time := time.Now()

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: current_time,
		UpdatedAt: current_time,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	feed_follow, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("user: %s now following feed: %s\n", feed_follow.UserName, feed_follow.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return errors.New("expects no argument")
	}
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	feed_follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("user: %s following feeds:\n", s.cfg.CurrentUserName)
	for _, feed_follow := range feed_follows {
		fmt.Printf("%s\n", feed_follow.FeedName)
	}

	return nil
}
