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
	"html"
	_ "github.com/lib/pq"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))

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

// handlers
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
	if len(cmd.args) != 1 {
		return errors.New("expects single argument, time between reqs, format 1m ex")
	}

	time_between_reqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Printf("collecting feeds every %v\n", time_between_reqs)

	ticker := time.NewTicker(time_between_reqs)
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Println("Error scraping:", err)
		}
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return errors.New("addfeed expects exactly 2 arguments, name and url")
	}
	name := cmd.args[0]
	url := cmd.args[1]

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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("expects 1 argument")
	}

	url := cmd.args[0]
	current_time := time.Now()

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

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		return errors.New("expects no argument")
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

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("expects url as argument (only)")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}

	params := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	if err := s.db.DeleteFeedFollow(context.Background(), params); err != nil {
		return err
	}

	fmt.Printf("unfollowed feed: %s\n", feed.Name)

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 1 {
		return errors.New("too many args")
	}
	limit := 8
	if len(cmd.args) == 1 {
		var err error
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}
	}

	params := database.GetUserPostsParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	posts, err := s.db.GetUserPosts(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d posts:\n", len(posts))
	for _, post := range posts {
		fmt.Println("=====================================")
		fmt.Printf("Title: \033]8;;%s\033\\%s\033]8;;\033\\\n", post.Url, post.Title)
		fmt.Printf("Feed: %s\n", post.FeedName)
		if post.Description.Valid {
			// Strip HTML tags and unescape HTML entities
			desc := post.Description.String
			re := regexp.MustCompile(`<[^>]*>`)
			desc = re.ReplaceAllString(desc, "")
			desc = html.UnescapeString(desc)
			fmt.Printf("Description: %s\n", desc)
		}
		fmt.Printf("Published: %s\n", post.PublishedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("=====================================")
		fmt.Println()
	}

	return nil
}

// helper
func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	current_time := time.Now()

	params := database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: current_time, Valid: true},
		UpdatedAt:     current_time,
		ID:            feed.ID,
	}

	if err = s.db.MarkFeedFetched(context.Background(), params); err != nil {
		return err
	}

	rss_feed, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, item := range rss_feed.Channel.Item {
		// Parse published date
		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			// Try alternate format
			publishedAt, err = time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				// Fallback to current time if parsing fails
				publishedAt = current_time
			}
		}

		params := database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: current_time,
			UpdatedAt: current_time,
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  item.Description != "",
			},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		}

		if err := s.db.CreatePost(context.Background(), params); err != nil {
			// Ignore duplicate URL errors
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
				continue
			}
			// Log other errors
			fmt.Println("Error creating post:", err)
		}
	}

	return nil
}

// middleware
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}
