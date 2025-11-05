# Publication Examples

Examples of publishing notes to leaflet.pub using the AT Protocol integration.

## Overview

The publication system allows you to sync your local notes with leaflet.pub, an AT Protocol-based publishing platform. You can pull drafts from leaflet, publish local notes, and maintain a synchronized writing workflow across platforms.

## Authentication

### Initial Authentication

Authenticate with your BlueSky account:

```sh
noteleaf pub auth username.bsky.social
```

This will prompt for your app password interactively.

### Authenticate with Password Flag

Provide credentials directly:

```sh
noteleaf pub auth username.bsky.social --password "your-app-password"
```

### Creating an App Password

1. Visit [bsky.app/settings/app-passwords](https://bsky.app/settings/app-passwords)
2. Create a new app password named "noteleaf"
3. Use that password (not your main password) for authentication

### Check Authentication Status

```sh
noteleaf pub status
```

## Pulling Documents from Leaflet

### Pull All Documents

Fetch all drafts and published documents:

```sh
noteleaf pub pull
```

This will:
- Connect to your leaflet account
- Fetch all documents in your repository
- Create new notes for documents not yet synced
- Update existing notes that have changed

### After Pulling

List the synced notes:

```sh
noteleaf pub list
```

View synced notes interactively:

```sh
noteleaf pub list --interactive
```

## Publishing Local Notes

### Publish a Note

Create a new document on leaflet from a local note:

```sh
noteleaf pub post 123
```

### Publish as Draft

Create as draft instead of publishing immediately:

```sh
noteleaf pub post 123 --draft
```

### Preview Before Publishing

See what would be posted without actually posting:

```sh
noteleaf pub post 123 --preview
```

### Validate Conversion

Check if markdown conversion will work:

```sh
noteleaf pub post 123 --validate
```

## Updating Published Documents

### Update an Existing Document

Update a previously published note:

```sh
noteleaf pub patch 123
```

### Preview Update

See what would be updated:

```sh
noteleaf pub patch 123 --preview
```

### Validate Update

Check conversion before updating:

```sh
noteleaf pub patch 123 --validate
```

## Batch Operations

### Publish Multiple Notes

Create or update multiple documents at once:

```sh
noteleaf pub push 1 2 3 4 5
```

This will:
- Create new documents for notes never published
- Update existing documents for notes already on leaflet

### Batch Publish as Drafts

```sh
noteleaf pub push 10 11 12 --draft
```

## Viewing Publications

### List All Synced Notes

```sh
noteleaf pub list
```

Aliases:
```sh
noteleaf pub ls
```

### Filter by Status

Published documents only:
```sh
noteleaf pub list --published
```

Drafts only:
```sh
noteleaf pub list --draft
```

All documents:
```sh
noteleaf pub list --all
```

### Interactive Browser

Browse with TUI interface:

```sh
noteleaf pub list --interactive
noteleaf pub list -i
```

With filters:
```sh
noteleaf pub list --published --interactive
noteleaf pub list --draft -i
```

## Common Workflows

### Initial Setup and Pull

Set up leaflet integration and pull existing documents:

```sh
# Authenticate
noteleaf pub auth username.bsky.social

# Check status
noteleaf pub status

# Pull all documents
noteleaf pub pull

# View synced notes
noteleaf pub list --interactive
```

### Publishing Workflow

Write locally, then publish to leaflet:

```sh
# Create a note
noteleaf note create "My Blog Post" --interactive

# List notes to get ID
noteleaf note list

# Publish as draft first
noteleaf pub post 42 --draft

# Review draft on leaflet.pub
# Make edits locally
noteleaf note edit 42

# Update the draft
noteleaf pub patch 42

# When ready, republish without --draft flag
noteleaf pub post 42
```

### Sync Workflow

Keep local notes in sync with leaflet:

```sh
# Pull latest changes from leaflet
noteleaf pub pull

# Make local edits
noteleaf note edit 123

# Push changes back
noteleaf pub patch 123

# Check sync status
noteleaf pub list --published
```

### Draft Management

Work with drafts before publishing:

```sh
# Create drafts
noteleaf pub post 10 --draft
noteleaf pub post 11 --draft
noteleaf pub post 12 --draft

# View all drafts
noteleaf pub list --draft

# Edit a draft locally
noteleaf note edit 10

# Update on leaflet
noteleaf pub patch 10

# Promote draft to published (re-post without --draft)
noteleaf pub post 10
```

### Batch Publishing

Publish multiple notes at once:

```sh
# Create several notes
noteleaf note create "Post 1" "Content 1"
noteleaf note create "Post 2" "Content 2"
noteleaf note create "Post 3" "Content 3"

# Get note IDs
noteleaf note list --static

# Publish all at once
noteleaf pub push 50 51 52

# Or as drafts
noteleaf pub push 50 51 52 --draft
```

### Review Before Publishing

Always preview and validate before publishing:

```sh
# Validate markdown conversion
noteleaf pub post 99 --validate

# Preview the output
noteleaf pub post 99 --preview

# If everything looks good, publish
noteleaf pub post 99
```

### Cross-Platform Editing

Edit on leaflet.pub, sync to local:

```sh
# Pull changes from leaflet
noteleaf pub pull

# View what changed
noteleaf pub list --interactive

# Make additional edits locally
noteleaf note edit 123

# Push updates back
noteleaf pub patch 123
```

### Status Monitoring

Check authentication and publication status:

```sh
# Check auth status
noteleaf pub status

# List published documents
noteleaf pub list --published

# Count publications
noteleaf pub list --published --static | wc -l
```

## Troubleshooting

### Re-authenticate

If authentication expires:

```sh
noteleaf pub auth username.bsky.social
```

### Check Status

Verify connection:

```sh
noteleaf pub status
```

### Force Pull

Re-sync all documents:

```sh
noteleaf pub pull
```

### Validate Before Publishing

If publishing fails, validate first:

```sh
noteleaf pub post 123 --validate
```

Check for markdown formatting issues that might not convert properly.

## Integration with Notes

### Publishing Flow

```sh
# Create note locally
noteleaf note create "Article Title" --interactive

# Add tags for organization
noteleaf note tag 1 --add published,blog

# Publish to leaflet
noteleaf pub post 1

# Continue editing locally
noteleaf note edit 1

# Sync updates
noteleaf pub patch 1
```

### Import from Leaflet

```sh
# Pull from leaflet
noteleaf pub pull

# View imported notes
noteleaf pub list

# Edit locally
noteleaf note edit 123

# Continue working with standard note commands
noteleaf note read 123
noteleaf note tag 123 --add imported
```

## Advanced Usage

### Selective Publishing

Publish only specific notes with a tag:

```sh
# Tag notes for publication
noteleaf note tag 10 --add ready-to-publish
noteleaf note tag 11 --add ready-to-publish

# List tagged notes
noteleaf note list --tags ready-to-publish

# Publish those notes
noteleaf pub push 10 11
```

### Draft Review Cycle

```sh
# Publish drafts
noteleaf pub push 1 2 3 --draft

# Review on leaflet.pub in browser
# Make edits locally based on feedback

# Update drafts
noteleaf pub push 1 2 3

# When ready, publish (create as non-drafts)
noteleaf pub post 1
noteleaf pub post 2
noteleaf pub post 3
```

### Publication Archive

Keep track of published work:

```sh
# Tag published notes
noteleaf note tag 123 --add published,2024,blog

# List all published notes
noteleaf note list --tags published

# Archive old publications
noteleaf note archive 123
```

## Notes

- Authentication tokens are stored in the configuration file
- Notes are matched by their leaflet record key (rkey)
- The `push` command intelligently chooses between `post` and `patch`
- Draft status is preserved when patching existing documents
- Use `--preview` and `--validate` flags to test before publishing
- Pull regularly to stay synced with changes made on leaflet.pub
