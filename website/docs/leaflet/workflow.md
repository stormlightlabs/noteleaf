---
title: Publishing Workflow
sidebar_label: Workflow
description: Post, patch, drafts, pulling, and syncing documents.
sidebar_position: 4
---

# Publishing Workflow

## Converting Notes to Leaflet Documents

Noteleaf converts markdown notes to leaflet's rich text block format:

**Supported Markdown Features**:

- Headers (`#`, `##`, `###`, etc.)
- Paragraphs
- Bold (`**bold**`)
- Italic (`*italic*`)
- Code (`inline code`)
- Strikethrough (`~~text~~`)
- Links (`[text](url)`)
- Code blocks (` ```language ... ``` `)
- Blockquotes (`> quote`)
- Lists (ordered and unordered)
- Horizontal rules (`---`)

**Conversion Process**:

1. Parse markdown into AST (abstract syntax tree)
2. Convert AST nodes to leaflet block records
3. Process text formatting into facets
4. Validate document structure
5. Upload to leaflet.pub via AT Protocol

## Creating a New Document

Publish a local note as a new leaflet document:

```sh
noteleaf pub post 123
```

This:

1. Converts the note to leaflet format
2. Creates a new document on leaflet.pub
3. Links the note to the document (stores the rkey)
4. Marks the note as published

**Create as draft**:

```sh
noteleaf pub post 123 --draft
```

Drafts are saved to leaflet but not publicly visible until you publish them.

**Preview before posting**:

```sh
noteleaf pub post 123 --preview
```

Shows what the document will look like without actually posting.

**Validate conversion**:

```sh
noteleaf pub post 123 --validate
```

Checks if the markdown converts correctly to leaflet format without posting.

**Save to file**:

```sh
noteleaf pub post 123 --preview --output document.json
noteleaf pub post 123 --preview --output document.txt --plaintext
```

## Updating Published Documents

Update an existing leaflet document from a local note:

```sh
noteleaf pub patch 123
```

Requirements:

- Note must have been previously posted or pulled from leaflet
- Note must have a leaflet record key (rkey) in the database

**Preserve draft/published status**: The `patch` command maintains the document's current status. If it's published, it stays published. If it's a draft, it stays a draft.

**Preview changes**:

```sh
noteleaf pub patch 123 --preview
```

**Validate before patching**:

```sh
noteleaf pub patch 123 --validate
```

## Managing Drafts

**Create as draft**:

```sh
noteleaf pub post 123 --draft
```

**Update draft**:

```sh
noteleaf pub patch 123
```

**List drafts**:

```sh
noteleaf pub list --draft
```

**Publish a draft**: Edit the draft on leaflet.pub or use the API to change status (command support coming in future versions).

## Pulling Documents from Leaflet

Sync leaflet documents to local notes:

```sh
noteleaf pub pull
```

This:

1. Authenticates with leaflet.pub
2. Fetches all documents in your repository
3. Creates new notes for documents not yet synced
4. Updates existing notes that have changed

**Matching logic**: Notes are matched to leaflet documents by their record key (rkey) stored in the database.
If a document doesn't have a corresponding note, a new one is created. If it does, the note is updated only if the content has changed (using CID for change detection).
