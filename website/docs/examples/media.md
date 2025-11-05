# Media Examples

Examples of managing your reading lists and watch queues using Noteleaf.

## Books

### Adding Books

Search and add from Open Library:

```sh
noteleaf media book add "The Name of the Wind"
noteleaf media book add "Project Hail Mary"
noteleaf media book add "Dune"
```

Add by author:

```sh
noteleaf media book add "Foundation by Isaac Asimov"
```

Add with specific year:

```sh
noteleaf media book add "1984 by George Orwell 1949"
```

### Viewing Books

List all books:

```sh
noteleaf media book list
```

Filter by status:

```sh
noteleaf media book list --status queued
noteleaf media book list --status reading
noteleaf media book list --status finished
```

### Managing Reading Status

Start reading:

```sh
noteleaf media book reading 1
```

Mark as finished:

```sh
noteleaf media book finished 1
```

### Tracking Progress

Update reading progress (percentage):

```sh
noteleaf media book progress 1 25
noteleaf media book progress 1 50
noteleaf media book progress 1 75
```

Update with page numbers:

```sh
noteleaf media book progress 1 150 --total 400
```

### Book Details

View book details:

```sh
noteleaf media book view 1
```

### Updating Book Information

Update book notes:

```sh
noteleaf media book update 1 --notes "Excellent worldbuilding and magic system"
```

Add rating:

```sh
noteleaf media book update 1 --rating 5
```

### Removing Books

Remove from list:

```sh
noteleaf media book remove 1
```

## Movies

### Adding Movies

Add movie:

```sh
noteleaf media movie add "The Matrix"
noteleaf media movie add "Inception"
noteleaf media movie add "Interstellar"
```

Add with year:

```sh
noteleaf media movie add "Blade Runner 1982"
```

### Viewing Movies

List all movies:

```sh
noteleaf media movie list
```

Filter by status:

```sh
noteleaf media movie list --status queued
noteleaf media movie list --status watched
```

### Managing Watch Status

Mark as watched:

```sh
noteleaf media movie watched 1
```

### Movie Details

View movie details:

```sh
noteleaf media movie view 1
```

### Updating Movie Information

Add notes and rating:

```sh
noteleaf media movie update 1 --notes "Mind-bending sci-fi" --rating 5
```

### Removing Movies

Remove from list:

```sh
noteleaf media movie remove 1
```

## TV Shows

### Adding TV Shows

Add TV show:

```sh
noteleaf media tv add "Breaking Bad"
noteleaf media tv add "The Wire"
noteleaf media tv add "Better Call Saul"
```

### Viewing TV Shows

List all shows:

```sh
noteleaf media tv list
```

Filter by status:

```sh
noteleaf media tv list --status queued
noteleaf media tv list --status watching
noteleaf media tv list --status watched
```

### Managing Watch Status

Start watching:

```sh
noteleaf media tv watching 1
```

Mark as finished:

```sh
noteleaf media tv watched 1
```

Put on hold:

```sh
noteleaf media tv update 1 --status on-hold
```

### TV Show Details

View show details:

```sh
noteleaf media tv view 1
```

### Updating TV Show Information

Update current episode:

```sh
noteleaf media tv update 1 --season 2 --episode 5
```

Add notes and rating:

```sh
noteleaf media tv update 1 --notes "Intense character development" --rating 5
```

### Removing TV Shows

Remove from list:

```sh
noteleaf media tv remove 1
```

## Common Workflows

### Weekend Watch List

Plan your weekend viewing:

```sh
# Add movies
noteleaf media movie add "The Shawshank Redemption"
noteleaf media movie add "Pulp Fiction"
noteleaf media movie add "Forrest Gump"

# View queue
noteleaf media movie list --status queued
```

### Reading Challenge

Track annual reading goal:

```sh
# Add books to queue
noteleaf media book add "The Lord of the Rings"
noteleaf media book add "The Hobbit"
noteleaf media book add "Mistborn"

# Check progress
noteleaf media book list --status finished
noteleaf media book list --status reading
```

### Binge Watching Tracker

Track TV series progress:

```sh
# Start series
noteleaf media tv add "Game of Thrones"
noteleaf media tv watching 1

# Update progress
noteleaf media tv update 1 --season 1 --episode 1
noteleaf media tv update 1 --season 1 --episode 2

# View current shows
noteleaf media tv list --status watching
```

### Media Recommendations

Keep track of recommendations:

```sh
# Add recommended items
noteleaf media book add "Sapiens" --notes "Recommended by John"
noteleaf media movie add "Parasite" --notes "Won Best Picture 2020"
noteleaf media tv add "Succession" --notes "From Reddit recommendations"

# View recommendations
noteleaf media book list --static | grep "Recommended"
```

### Review and Rating

After finishing, add review:

```sh
# Book review
noteleaf media book finished 1
noteleaf media book update 1 \
  --rating 5 \
  --notes "Masterful storytelling. The magic system is one of the best in fantasy."

# Movie review
noteleaf media movie watched 2
noteleaf media movie update 2 \
  --rating 4 \
  --notes "Great cinematography but slow pacing in second act."

# TV show review
noteleaf media tv watched 3
noteleaf media tv update 3 \
  --rating 5 \
  --notes "Best character development I've seen. Final season was perfect."
```

### Genre Organization

Organize by genre using notes:

```sh
noteleaf media book add "Snow Crash" --notes "Genre: Cyberpunk"
noteleaf media book add "Neuromancer" --notes "Genre: Cyberpunk"
noteleaf media book add "The Expanse" --notes "Genre: Space Opera"

# Find by genre
noteleaf media book list --static | grep "Cyberpunk"
```

### Currently Consuming

See what you're currently reading/watching:

```sh
noteleaf media book list --status reading
noteleaf media tv list --status watching
```

### Completed This Month

View completed items:

```sh
noteleaf media book list --status finished
noteleaf media movie list --status watched
noteleaf media tv list --status watched
```

### Clear Finished Items

Archive or remove completed media:

```sh
# Remove watched movies
noteleaf media movie remove 1 2 3

# Remove finished books
noteleaf media book remove 4 5 6
```

## Statistics and Reports

### Reading Statistics

Count books by status:

```sh
echo "Queued: $(noteleaf media book list --status queued --static | wc -l)"
echo "Reading: $(noteleaf media book list --status reading --static | wc -l)"
echo "Finished: $(noteleaf media book list --status finished --static | wc -l)"
```

### Viewing Habits

Track watch queue size:

```sh
echo "Movies to watch: $(noteleaf media movie list --status queued --static | wc -l)"
echo "Shows in progress: $(noteleaf media tv list --status watching --static | wc -l)"
```
