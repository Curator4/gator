package main

import (
	"os"
	"fmt"
	"log"
	"errors"
	"time"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/curator4/gator/internal/config"
	"github.com/curator4/gator/internal/database"
	_ "github.com/lib/pq"
)

// constants
const dbURL = "postgres://curator:solitude4@localhost:5432/gator?sslmode=disable"

// structs
type state struct {
	cfg *config.Config
	db *database.Queries
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
		db: database.New(db),
	}

	c := &commands{
		handlers: make(map[string]func(*state, command) error),
	}

	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	
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
		ID: uuid.New(),
		CreatedAt: current_time,
		UpdatedAt: current_time,
		Name: username,
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
