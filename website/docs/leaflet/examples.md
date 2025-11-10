---
title: Publishing Examples
sidebar_label: Examples
description: End-to-end examples for posting, drafting, and syncing.
sidebar_position: 6
---

# Publishing Examples

## Publishing a Blog Post

**Write the post locally**:

```sh
noteleaf note create "Understanding AT Protocol" --editor
```

Write in markdown, save, and close editor.

**Preview the conversion**:

```sh
noteleaf pub post <note-id> --preview
```

Review the output to ensure formatting is correct.

**Publish**:

```sh
noteleaf pub post <note-id>
```

**Update later**:

```sh
noteleaf note edit <note-id>
# Make changes
noteleaf pub patch <note-id>
```

## Draft Workflow

**Create draft**:

```sh
noteleaf note create "Work in Progress" --editor
noteleaf pub post <note-id> --draft
```

**Iterate locally**:

```sh
noteleaf note edit <note-id>
noteleaf pub patch <note-id>  # Updates draft
```

**Publish when ready**: Use leaflet.pub web interface to change draft to published status (CLI command coming in future versions).

## Syncing Existing Content

**Pull all leaflet documents**:

```sh
noteleaf pub pull
```

**List synced documents**:

```sh
noteleaf pub list
```

**Read a synced document**:

```sh
noteleaf pub read <note-id>
```

**Edit locally and push updates**:

```sh
noteleaf note edit <note-id>
noteleaf pub patch <note-id>
```
