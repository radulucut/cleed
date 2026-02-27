
<div align="right">
  <details>
    <summary >🌐 Language</summary>
    <div>
      <div align="center">
        <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=en">English</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=zh-CN">简体中文</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=zh-TW">繁體中文</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=ja">日本語</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=ko">한국어</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=hi">हिन्दी</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=th">ไทย</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=fr">Français</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=de">Deutsch</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=es">Español</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=it">Italiano</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=ru">Русский</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=pt">Português</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=nl">Nederlands</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=pl">Polski</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=ar">العربية</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=fa">فارسی</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=tr">Türkçe</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=vi">Tiếng Việt</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=id">Bahasa Indonesia</a>
        | <a href="https://openaitx.github.io/view.html?user=radulucut&project=cleed&lang=as">অসমীয়া</
      </div>
    </div>
  </details>
</div>

# cleed

Simple feed reader for the command line.

![Test](https://github.com/radulucut/cleed/actions/workflows/tests.yml/badge.svg)

![Screenshot](./screenshot.png)

## Installation

#### MacOS - Homebrew

```bash
brew tap radulucut/cleed
brew install cleed
```

#### Windows - Scoop

```bash
scoop bucket add cleed https://github.com/radulucut/scoop-cleed
scoop install cleed
```

#### From source

```bash
go run main.go
```

#### Binary

Download the latest binary from the [releases](https://github.com/radulucut/cleed/releases) page.

## Usage

#### Follow a feed

```bash
# Add a feed to the default list
cleed follow https://example.com/feed.xml

# Add multiple feeds to a list
cleed follow https://example.com/feed.xml https://example2.com/feed --list mylist
```

#### Display feeds

```bash
# Display feeds from all lists
cleed

# Display feeds from a specific list
cleed --list my-list

# Display feeds since a specific date
cleed --since "2024-01-01 12:03:04"

# Display feeds since a specific period
cleed --since "1d"

# Display feeds since the last run
cleed --since last

# Display feeds from a specific list and limit the number of items
cleed --list my-list --limit 10

# Search for items
cleed --search "keyword" --limit 10

# Search for items in cached feeds
cleed --search "keyword" -C

# Search for items using a regular expression (https://github.com/google/re2/wiki/Syntax)
cleed --searchr "^keyword"

# Search for items using a regular expression case insensitive
cleed --searchr "(?i)keyword"

# Using a proxy
cleed --proxy socks5://user:password@proxy.example.com:8080
```

#### Unfollow a feed

```bash
# Remove a feed from the default list
cleed unfollow https://example.com/feed.xml

# Remove multiple feeds from a list
cleed unfollow https://example.com/feed.xml https://example2.com/feed --list mylist
```

#### List feeds

```bash
# Show all lists
cleed list

# Show all feeds in a list
cleed list mylist

# Rename a list
cleed list mylist --rename newlist

# Merge a list. Move all feeds from anotherlist to mylist and remove anotherlist
cleed list mylist --merge anotherlist

# Remove a list
cleed list mylist --remove

# Import feeds from a file
cleed list mylist --import-from-file feeds.txt

# Import feeds from an OPML file
cleed list mylist --import-from-opml feeds.opml

# Export feeds to a file
cleed list mylist --export-to-file feeds.txt

# Export feeds to an OPML file
cleed list mylist --export-to-opml feeds.opml

# Export only cached feeds to an OPML file
cleed list --export-to-opml feeds.opml -C
```

#### Configuration

```bash
# Display configuration
cleed config

# Set the user agent
cleed config --user-agent="My User Agent"

# Disable styling
cleed config --styling=2

# Map color 0 to 230 and color 1 to 213
cleed config --map-colors=0:230,1:213

# Remove color mapping for color 0
cleed config --map-colors=0:

# Clear all color mappings
cleed config --map-colors=

# Display color range. Useful for finding colors to map
cleed config --color-range

# Enable run summary
cleed config --summary=1

# Set the miniflux token
cleed config --miniflux-token="your_token_here"`
```

> **Color mapping**
>
> You can map the colors used in the feed reader to any color you want. This is useful if certain colors are not visible in your terminal based on the color scheme that you are using.
>
> Run `cleed config --color-range` to see the color range and map the colors that you want using the `cleed config --map-colors` command.

#### Miniflux

```bash
# Push feeds to Miniflux
cleed miniflux push

# Pull feeds from Miniflux
cleed miniflux pull
```

#### Explore feeds

```bash
# Explore feeds from the default repository (https://github.com/radulucut/cleed-explore)
cleed explore

# Fetch the latest changes and explore feeds from the default repository
cleed explore --update

# Explore feeds from a repository
cleed explore https://github.com/radulucut/cleed-explore.git

# Limit the number of items to display from each list
cleed explore --limit 5

# Search for items (title, description)
cleed explore --search "news"

# Import all feeds into my feeds
cleed explore --import --limit 0

# Import feeds from search results
cleed explore --import --search "news"

# Remove a repository
cleed explore https://github.com/radulucut/cleed-explore.git --remove
```

> **Note**
>
> The explore command expects git to be installed in order to fetch the repository, and it will only look at `.opml` files when exploring a repository.

#### Custom data path

You can set the `CLEED_DATA_PATH` environment variable to specify a custom path for the data directory. All data will be stored in this directory (config, cache, etc).

#### Help

```bash
cleed --help
```
