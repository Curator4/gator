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
gator reset                # Reset database (deletes EVERYTHING)
```

# extra
- i added the agg to my .zshrc, lets see how it goes.
- changed default browse limit to 8
- added 'gb' gator browse alias
- added the help command
- added some random feeds:
{Name:Hugging Face Blog Url:https://huggingface.co/blog/feed.xml Username:curator}
{Name:Rock Paper Shotgun Url:https://www.rockpapershotgun.com/feed Username:curator}
{Name:PC Gamer Url:https://www.pcgamer.com/rss/ Username:curator}
{Name:Paradox Interactive Url:https://www.paradoxinteractive.com/games/feed Username:curator}
{Name:IGN Strategy Games Url:https://feeds.ign.com/ign/games-all Username:curator}
{Name:Anime News Network Url:https://www.animenewsnetwork.com/all/rss.xml Username:curator}
{Name:Ars Technica Url:https://feeds.arstechnica.com/arstechnica/index Username:curator}
{Name:Go Blog Url:https://go.dev/blog/feed.atom Username:curator}
{Name:MIT Tech Review Url:https://www.technologyreview.com/feed/ Username:curator}
{Name:OpenAI Blog Url:https://openai.com/blog/rss.xml Username:curator}
{Name:Anthropic News Url:https://www.anthropic.com/news/rss.xml Username:curator}
{Name:Google AI Blog Url:https://ai.googleblog.com/feeds/posts/default Username:curator}
{Name:DeepMind Blog Url:https://www.deepmind.com/blog/rss.xml Username:curator}
