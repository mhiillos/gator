package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/mhiillos/go-blog-aggregator/internal/config"
	"github.com/mhiillos/go-blog-aggregator/internal/database"
)

type state struct {
	db *database.Queries
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
	_, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		return fmt.Errorf("User %q does not exist in database", userName)
	}
	err = s.cfg.SetUser(userName)
	if err != nil {
		return err
	}
	fmt.Printf("User %q set.\n", userName)
	return nil
}

// Adds a new user to db
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("Please state the username to register")
	}
	userName := cmd.args[0]
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: userName,
	})
	if err != nil {
		return fmt.Errorf("User with name %q already exists", userName)
	}
	err = s.cfg.SetUser(userName)
	if err != nil {
		return err
	}
	fmt.Printf("User %s created\n", userName)
	fmt.Println(user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error resetting table: %s", err)
	}
	fmt.Println("Users table reset")
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	s := &state{cfg: &cfg}
	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	s.db = dbQueries

	// Initalize commands struct
	cmds := commands{
		commands: make(map[string]func(*state, command) error),
	}

	// Register commands
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)

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
		os.Exit(1)
	}
}
