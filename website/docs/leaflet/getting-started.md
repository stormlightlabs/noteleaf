---
title: Getting Started with Leaflet
sidebar_label: Getting Started
description: Prerequisites, app passwords, and authentication commands.
sidebar_position: 2
---

# Getting Started with Leaflet

## Prerequisites

- Noteleaf installed and configured
- A BlueSky & [Leaflet](https://leaflet.pub) account (create at [bsky.app](https://bsky.app))
- App password for authentication

## Creating an App Password

For security, Noteleaf uses app passwords instead of your main BlueSky password:

1. Go to [bsky.app/settings/app-passwords](https://bsky.app/settings/app-passwords)
2. Click "Add App Password"
3. Name it "noteleaf" or similar
4. Copy the generated password (you won't see it again)

### Authentication

Authenticate with your BlueSky handle and app password:

```sh
noteleaf pub auth username.bsky.social
```

You'll be prompted for the app password. Alternatively, provide it via flag:

```sh
noteleaf pub auth username.bsky.social --password <app-password>
```

**Re-authentication**: If your session expires, run `pub auth` again. Noteleaf remembers your last authenticated handle, so you can just run:

```sh
noteleaf pub auth
```

### Check Authentication Status

Verify you're authenticated:

```sh
noteleaf pub status
```

Shows your handle and session state.
