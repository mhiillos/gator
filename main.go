package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
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

// For reading RSS data
type RSSFeed struct {
	Channel struct {
		Title string       `xml:"title"`
		Link string        `xml:"link"`
		Description string `xml:"description"`
		Item []RSSItem     `xml:"item"`
	}	`xml:"channel"`
}

type RSSItem struct {
		Title string       `xml:"title"`
		Link string        `xml:"link"`
		Description string `xml:"description"`
		PubDate string     `xml:"pubDate"`
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

// Resets user table
func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error resetting table: %s", err)
	}
	fmt.Println("Users table reset")
	return nil
}

// Prints all the users
func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	currentUser := s.cfg.CurrentUsername
	for _, user := range users {
		userStr := fmt.Sprintf("* %s", user.Name)
		if user.Name == currentUser {
			userStr += " (current)"
		}
		fmt.Printf("%s\n", userStr)
	}
	return nil
}

func fetchFeed (ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent", "gator")
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error fetching RSS Feed with status %d: %s", res.StatusCode, res.Body)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("Error reading bytes from response")
	}
	rss := &RSSFeed{}
	err = xml.Unmarshal(raw, rss)
	if err != nil {
		return nil, errors.New("Error unmarshaling XML")
	}

	// Decode escaped HTML entities
	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for i := range rss.Channel.Item {
		rss.Channel.Item[i].Description = html.UnescapeString(rss.Channel.Item[i].Description)
		rss.Channel.Item[i].Title = html.UnescapeString(rss.Channel.Item[i].Title)
	}
	return rss, nil
}

// Adds a feed
func handlerAddfeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return errors.New("Please pass the Name and URL of the feed as arguments")
	}
	name := cmd.args[0]
	url := cmd.args[1]
	currentUser := s.cfg.CurrentUsername
	usr, err := s.db.GetUser(context.Background(), currentUser)
	if err != nil {
		return fmt.Errorf("User %q not found", currentUser)
	}
	resFeed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    usr.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Feed created (%q)", resFeed.ID)
	// Add FeedFollow for the user
	resFeedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: usr.ID,
		FeedID: resFeed.ID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Feed follow created (%q)", resFeedFollow.ID)
	return nil
}

// 
func handlerAgg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), url)
	if err != nil {
		return err
	}
	if feed == nil {
		return errors.New("feed is nil")
	}
	jsonData, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return fmt.Errorf("error mashaling feed to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// Prints all the feeds
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedsWithCreators(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting feeds: %q", err)
	}
	fmt.Printf("List of feeds:\n")
	for _, feed := range feeds {
		fmt.Printf("Name: %s, URL: %s, CreatedBy: %s\n", feed.Name, feed.Url, feed.UserName)
	}
	return nil
}

// Adds the current user to follow an RSS feed
func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return errors.New("Please pass URL as an argument")
	}
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("URL %q does not exist in db", cmd.args[0])
	}
	user, _ := s.db.GetUser(context.Background(), s.cfg.CurrentUsername)
	res, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed follow: %w", err)
	}
	fmt.Printf("%s started following feed %s\n", res.UserName, res.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command) error {
	user, _ := s.db.GetUser(context.Background(), s.cfg.CurrentUsername)
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("User %q does not follow any feeds", s.cfg.CurrentUsername)
	}
	fmt.Printf("User %q is following:\n", s.cfg.CurrentUsername)
	for _, feedFollow := range(feedFollows) {
		fmt.Printf("  - %s\n", feedFollow.FeedName)
	}
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
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddfeed)
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", handlerFollow)
	cmds.register("following", handlerFollowing)

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
