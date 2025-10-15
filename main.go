package main

import (
	"os"
	"fmt"
	"log"
	"errors"
	"github.com/curator4/gator/internal/config"
)

// structs
type state struct {
	cfg *config.Config
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

	s := &state{
		cfg: &cfg,
	}
	c := &commands{
		handlers: make(map[string]func(*state, command) error),
	}

	c.register("login", handlerLogin)

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
		fmt.Printf("could not run command?\n")
		os.Exit(1)
	}
}


// functions
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("login expects a single argument")
	}
	username := cmd.args[0]
	if err := (*s.cfg).SetUser(username); err != nil {
		return err
	}
	fmt.Printf("user: %s has been logged in\n", username)
	return nil
}
