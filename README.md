# Setup

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
gator reset                # Reset database (delete all users)
```
