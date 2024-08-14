# cleed

Simple feed reader for the command line.

![Test](https://github.com/radulucut/cleed/actions/workflows/tests.yml/badge.svg)

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

# Display feeds from the last 1 day
cleed --last 1d

# Display feeds since a specific date
cleed --since "2024-01-01 12:03:04"

# Display feeds from a specific list and limit the number of feeds
cleed --list my-list --limit 10
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
```
