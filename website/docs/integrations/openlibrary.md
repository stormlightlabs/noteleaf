---
title: Open Library API
sidebar_label: Open Library
sidebar_position: 1
description: Book metadata via Open Library API integration.
---

# Open Library API

Noteleaf integrates with [Open Library](https://openlibrary.org) to fetch book metadata, search for books, and enrich your reading list.

## Overview

Open Library provides:

- Book search by title, author, ISBN
- Work and edition metadata
- Author information
- Cover images
- Subject classifications
- Publication details

## Configuration

No API key required. Open Library is a free, open API service.

Optional user agent configuration is handled automatically:

```toml
# .noteleaf.conf.toml
# No configuration needed for Open Library
```

## Rate Limiting

Open Library enforces rate limits:

- 180 requests per minute
- 3 requests per second
- Burst limit: 5 requests

Noteleaf automatically manages rate limiting to stay within these boundaries.

## Book Search

Search for books from the command line:

```sh
noteleaf book search "Design Patterns"
noteleaf book search "Neal Stephenson"
```

Interactive selection shows:

- Title
- Author(s)
- First publication year
- Edition count
- Publisher information

## Book Metadata

When adding a book, Noteleaf fetches:

- Title
- Author names
- Publication year
- Edition information
- Subjects/genres
- Description (when available)
- Cover ID

## API Endpoints

### Search Endpoint

```
GET https://openlibrary.org/search.json
```

Parameters:

- `q`: Search query
- `offset`: Pagination offset
- `limit`: Results per page
- `fields`: Requested fields

### Work Endpoint

```
GET https://openlibrary.org/works/{work_key}.json
```

Returns detailed work information including authors, description, subjects, and covers.

## Data Mapping

Open Library data maps to Noteleaf book fields:

| Open Library | Noteleaf Field |
|--------------|----------------|
| title | Title |
| author_name | Author |
| first_publish_year | Notes (included) |
| edition_count | Notes (included) |
| publisher | Notes (included) |
| subject | Notes (included) |
| cover_i | Notes (cover ID) |

## Example API Response

Search result document:

```json
{
  "key": "/works/OL45804W",
  "title": "Design Patterns",
  "author_name": ["Erich Gamma", "Richard Helm"],
  "first_publish_year": 1994,
  "edition_count": 23,
  "isbn": ["0201633612", "9780201633610"],
  "publisher": ["Addison-Wesley"],
  "subject": ["Software design", "Object-oriented programming"],
  "cover_i": 8644882
}
```

## Limitations

### No Direct Page Count

Open Library doesn't consistently provide page counts in search results. Use the interactive editor to add page counts manually if needed.

### Author Keys vs Names

Work endpoints return author keys (`/authors/OL123A`) rather than full names. Noteleaf displays available author names from search results.

### Cover Images

Cover IDs are stored but not automatically downloaded. Future versions may support local cover image caching.

## Error Handling

### Network Issues

```sh
noteleaf book search "query"
# Error: failed to connect to Open Library
```

Check internet connection and Open Library status.

### Rate Limit Exceeded

Noteleaf automatically waits when approaching rate limits. If you see delays, this is normal behavior.

### No Results

```sh
noteleaf book search "very obscure title"
# No results found
```

Try:

- Different search terms
- Author names instead of titles
- ISBNs for specific editions

## API Service Architecture

Implementation in `internal/services/services.go`:

```go
type BookService struct {
    client  *http.Client
    limiter *rate.Limiter
    baseURL string
}
```

Features:

- Automatic rate limiting
- Context-aware requests
- Proper error handling
- Timeout management (30s)

## Custom User Agent

Noteleaf identifies itself to Open Library:

```
User-Agent: Noteleaf/v{version} (contact: info@stormlightlabs.org)
```

This helps Open Library track API usage and contact developers if needed.

## Resources

- [Open Library API Documentation](https://openlibrary.org/dev/docs/api/books)
- [Open Library Search](https://openlibrary.org/dev/docs/api/search)
- [Open Library Covers](https://openlibrary.org/dev/docs/api/covers)
- [Rate Limiting Policy](https://openlibrary.org/developers/api)
