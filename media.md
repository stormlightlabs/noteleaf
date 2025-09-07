# MEDIA management feature

## Current State Analysis

- Existing Media Service (`/internal/services/media.go`):
    - Rotten Tomatoes scraping with colly for movies/TV search
    - Rich metadata extraction (Movie, TVSeries, TVSeason structs)
    - Search functionality via SearchRottenTomatoes()
    - Detailed metadata fetching via FetchMovie(), FetchTVSeries(), FetchTVSeason()

- Book Search Pattern (`/internal/handlers/books.go`):
    - Uses APIService interface with `Search()`, `Get()`, `Check()`, `Close()` methods
    - Interactive and static search modes
    - Number-based selection UX
    - Converts API results to models.Book via interface

- Models (`/internal/models/models.go`):
    - Movie and TVShow structs already implement Model interface
    - Both have proper status tracking (queued, watched, etc.)

## Media Service Refactor

### Create MovieService that implement APIService

```go
// MovieService implements APIService for Rotten Tomatoes movies
type MovieService struct {
    client  *http.Client
    limiter *rate.Limiter
}
```

### Create TVService that implement APIService

```go
// TVService implements APIService for Rotten Tomatoes TV shows
type TVService struct {
    client  *http.Client
    limiter *rate.Limiter
}
```

### Implement APIService

- `Search(ctx, query, page, limit)` - Use existing SearchRottenTomatoes() and convert results to []*models.Model
- `Get(ctx, id)` - Use existing FetchMovie() / FetchTVSeries() with Rotten Tomatoes URLs
- `Check(ctx)` - Simple connectivity test to Rotten Tomatoes
- `Close()` - Cleanup resources

### Result Conversion

- Convert services.Media search results to models.Movie / models.TVShow
- Convert detailed metadata structs to models with proper status defaults
- Extract key information (title, year, rating, description) into notes field

## Handler Implementation

### Create MovieHandler similar to BookHandler

```go
type MovieHandler struct {
    db      *store.Database
    config*store.Config
    repos   *repo.Repositories
    service*services.MovieService
}
```

### Implement search

- `SearchAndAddMovie(ctx, args, interactive)` - Mirror book search UX
- `SearchAndAddTV(ctx, args, interactive)` - Same pattern for TV shows
- Number-based selection interface identical to books

### Database Integration

- Add movie/TV repositories if not already present
- Ensure proper CRUD operations for queue management

## Commands

### Update definitions

- Replace stubbed movie commands with real implementations
- Replace stubbed TV commands with real implementations
- Connect to new handlers with proper error handling

### Structure

```sh
# Movies

media movie add [search query...] [-i for interactive]
media movie list [--all|--watched|--queued]
media movie watched <id>
media movie remove <id>

# TV Shows

media tv add [search query...] [-i for interactive]
media tv list [--all|--watched|--queued]
media tv watched <id>
media tv remove <id>
```

## UX Consistency

### Search

1. Parse search query from args
2. Show "Loading..." progress indicator
3. Display numbered results with title, year, rating
4. Prompt for selection (1-N or 0 to cancel)
5. Add selected item to queue with "queued" status
6. Confirm addition to user

### Interactivity

- Use existing TUI patterns from book/task lists
- Browse search results with keyboard navigation
- Preview detailed metadata before adding

## Key Implementation Details

_Rate Limiting_: Add rate limiter to media services (Rotten Tomatoes likely has limits)

_Error Handling_: Robust handling of scraping failures, network issues, parsing errors

_Data Mapping_:
    - Map Rotten Tomatoes critic scores to model rating fields
    - Extract genres, cast, descriptions into notes field
    - Handle missing or incomplete metadata gracefully

_Caching_: Consider caching search results to reduce API calls during selection

_Status_: Default new items to "queued" status, provide commands to update
