# Setup
my completion of the [bootdev](https://www.boot.dev/dashboard) gator (blog aggregator) project. in short.. a project to integrate http calls (rss feeds) with a local postgres database. I didn't do much outside the scope of the assignment.


frankly, I had been wanting to make some sort of rudimentary plugin to like show important news or cool quotes etc. So i might come back to this at some point, or well, not this. but use it as reference to a more involved tui version. but am kind of in rush lately.

Thx to bootdev team, good project.

## Software needed:
- postgres
- Go (compilation)

## Config

To run you need a `~/.gatorconfig.json` file with your database URL and username:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": "your_username"
}
```

## Install
- create postgres gator db
- run migrations in sql/schema with goose
- `go run . <command>` to run for testing
- **go install** to compile and install the binary to the go bin directory, then runnable with just gator

# Commands

Run `gator help` to see all available commands.

Intended usage: Keep `agg` running in background, then use `browse` to read posts.

### User Management
```bash
gator register <username>  # Create a new user
gator login <username>     # Login as a user
gator users                # List all users
```

### Feed Management
```bash
gator addfeed <name> <url> # Add an RSS feed
gator feeds                # List all feeds
gator follow <url>         # Follow a feed
gator following            # Show feeds you're following
gator unfollow <url>       # Unfollow a feed
```

### Reading
```bash
gator browse [limit]       # Browse recent posts (default 8)
gator agg <duration>       # Run feed aggregator (e.g., 1m, 30s)
```

### Other
```bash
gator help                 # Show help message
gator reset                # Reset database (deletes EVERYTHING)
```

# Tips

- Add `gator agg 1m &` to your shell rc file to auto-start the aggregator
- Create an alias like `alias gb='gator browse'` for quick access
- Default browse limit is 8 posts

# Example RSS Feeds

```
https://huggingface.co/blog/feed.xml
https://www.rockpapershotgun.com/feed
https://www.pcgamer.com/rss/
https://feeds.arstechnica.com/arstechnica/index
https://go.dev/blog/feed.atom
https://www.technologyreview.com/feed/
https://openai.com/blog/rss.xml
https://www.anthropic.com/news/rss.xml
```
