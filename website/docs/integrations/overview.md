---
title: External Integrations
sidebar_label: Overview
sidebar_position: 1
description: Overview of external service integrations.
---

# External Integrations

Noteleaf integrates with external services to enrich your productivity workflow and extend functionality beyond local storage.

## Available Integrations

### Open Library API

Free book metadata service for building your reading list.

**Features:**

- Search books by title, author, ISBN
- Fetch metadata (author, year, subjects)
- Edition and publication information
- No API key required

**Use Cases:**

- Adding books to reading list
- Enriching book metadata
- Discovering related works

See [Open Library API](./openlibrary.md) for details.

### Leaflet.pub

Decentralized publishing platform built on AT Protocol.

**Features:**

- Publish notes as structured documents
- Pull existing documents into local notes
- Update published content
- Manage drafts and publications

**Use Cases:**

- Blog publishing from terminal
- Long-form content management
- Decentralized content ownership

See [Leaflet.pub section](../leaflet/intro.md) for details.

### AT Protocol (Bluesky)

Authentication and identity via AT Protocol network.

**Features:**

- Decentralized identity (DID)
- Session management
- Token refresh
- Secure authentication

**Use Cases:**

- Leaflet.pub authentication
- Portable identity across services
- Content verification

See [Authentication](../leaflet/authentication.md) for details.

## Integration Architecture

### Service Layer

External integrations live in `internal/services/`:

- `services.go` - Open Library API client
- `atproto.go` - AT Protocol authentication
- `http.go` - HTTP utilities and rate limiting

### Rate Limiting

All external services use rate limiting to respect API quotas:

- Open Library: 3 requests/second
- AT Protocol: Per PDS configuration

Rate limiters are built-in and automatic.

### Error Handling

Services implement consistent error handling:

- Network errors
- Rate limit exceeded
- Authentication failures
- Invalid responses

Errors propagate to user with actionable messages.

## Configuration

Integration configuration in `.noteleaf.conf.toml`:

```toml
# Open Library (no configuration needed)
# book_api_key = ""  # Reserved for future use

# AT Protocol / Leaflet.pub
atproto_handle = "username.bsky.social"
atproto_did = "did:plc:..."
atproto_pds_url = "https://bsky.social"
atproto_access_jwt = "..."
atproto_refresh_jwt = "..."
```

See [Configuration](../Configuration.md) for all options.

## Offline Support

Noteleaf works fully offline for local data. Integrations are optional enhancements:

- Books can be added manually without Open Library
- Notes exist locally without Leaflet.pub
- Tasks and media work without any external service

External services enhance but don't require connectivity.

## Privacy and Data

### Data Sent

**Open Library:**

- Search queries
- Work/edition IDs

**AT Protocol:**

- Handle/DID
- Published note content
- Authentication credentials

### Data Stored Locally

- API responses (cached)
- Session tokens
- Publication metadata

### No Tracking

Noteleaf does not:

- Track usage
- Send analytics
- Share data with third parties
- Require accounts (except for publishing)

## Resources

- [Open Library API Documentation](https://openlibrary.org/developers/api)
- [AT Protocol Docs](https://atproto.com)
- [Leaflet.pub](https://leaflet.pub)
- [Bluesky](https://bsky.app)
