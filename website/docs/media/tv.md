---
title: TV Shows
sidebar_label: TV Shows
description: Track long-form series with simple queue management.
sidebar_position: 4
---

# TV Shows

TV commands live under `noteleaf media tv`. Like movies, they use Rotten Tomatoes search so you can trust the spelling and canonical links.

## Add Shows

```sh
noteleaf media tv add "Breaking Bad"
```

- Inline mode shows up to five matches and asks you to choose.
- `-i/--interactive` is wired up for the future list selector.

Every new show starts as `queued`.

## List the Queue

```sh
noteleaf media tv list                 # queued shows
noteleaf media tv list --watching      # in-progress series
noteleaf media tv list --watched       # finished shows
noteleaf media tv list --all           # everything
```

Output includes the ID, title, optional season/episode numbers (once those fields are set), status, critic-score snippet, and timestamps.

## Update Status

Use semantic verbs instead of editing the status manually:

```sh
noteleaf media tv watching 8   # Moved to “currently watching”
noteleaf media tv watched 8    # Mark completed
noteleaf media tv remove 8     # Drop from the queue entirely
```

Each transition records `last_watched` so you know when you left off. Future releases will expose explicit season/episode commands; until then store quick reminders in a linked note:

```markdown
### TV checklist
- TV #8 — resume Season 3 Episode 5
- TV #15 — waiting for new season announcement
```

## Organization Tips

- Use `noteleaf media tv list --watching | fzf` to pick tonight’s episode.
- Pipe `--all` into `rg "HBO"` to filter on the metadata snippet that contains the network/URL.
- Include `TV #ID` references in your weekly review note so you can jump back with a single ID lookup.

## What Gets Stored

- Rotten Tomatoes critic info plus canonical URL (inside the Notes column).
- Optional season/episode integers for future episode tracking (already part of the schema).
- Added timestamps and “last watched” timestamps.

Because shows can last months, keeping the queue short (just what you plan to watch soon) makes the `list` output far easier to scan.
