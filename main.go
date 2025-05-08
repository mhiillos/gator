package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/mhiillos/go-blog-aggregator/internal/config"
)

type state struct {
	cfg *config.Config
}

// Stores command names and their arguments
type command struct {
	name string
	args []string
}

// Stores all available commands
type commands struct {
	commands map[string]func(*state, command) error
}

// This function calls the handlers for different commands
func (c *commands) run(s *state, cmd command) error {
	fun, ok := c.commands[cmd.name]
	if !ok {
		return errors.New("Command does not exist")
	}
	err := fun(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

// This function registers a new a new handler function for a command name
func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

// Sets the user ("login")
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("Please state the username as the argument")
	}
	userName := cmd.args[0]
	err := s.cfg.SetUser(userName)
	if err != nil {
		return err
	}
	fmt.Printf("User %s set.\n", userName)
	return nil
}

func main() {
	cfg, err := config.Read()
	s := &state{cfg: &cfg}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initalize commands struct
	cmds := commands{
		commands: make(map[string]func(*state, command) error),
	}

	// Register commands
	cmds.register("login", handlerLogin)

	// Read user input
	args := os.Args
	if len(args) < 2 {
		fmt.Print("No arguments provided.\nUsage:\n  go run . <command> [args]\n")
		os.Exit(1)
	}
	cmdName := args[1]
	cmdArgs := args[2:]
	cmd := command{name: cmdName, args: cmdArgs}
	err = cmds.run(s, cmd)
	if err != nil {
		fmt.Println(err)
	}
}
