---
title: Movies
sidebar_label: Movies
description: Keep track of your movie queue with Rotten Tomatoes metadata.
sidebar_position: 3
---

# Movies

Movie commands hang off `noteleaf media movie`. Results use Rotten Tomatoes search so you get consistent titles plus critic scores.

## Add Movies

```sh
noteleaf media movie add "The Matrix"
```

What happens:

1. The CLI fetches the first five Rotten Tomatoes matches.
2. You select the right one by number.
3. The chosen movie is inserted into the local queue with status `queued`.

The `-i/--interactive` flag is reserved for a future selector; currently the inline prompt is the quickest path.

## List and Filter

```sh
# Default: queued items only
noteleaf media movie list

# Include everything
noteleaf media movie list --all

# Review history
noteleaf media movie list --watched
```

Each entry shows:

- `ID` and title.
- Release year (if Rotten Tomatoes provided one).
- Status (`queued` or `watched`).
- Critic score snippet (stored inside the Notes column).
- Watched timestamp for completed items.

## Mark Movies as Watched

```sh
noteleaf media movie watched 12
```

The command sets the status to `watched` and records `watched_at` using the current timestamp. Removing an item uses the same ID:

```sh
noteleaf media movie remove 12
```

Use removal for titles you abandoned or added by mistake—the CLI deletes the database entry so your queue stays focused.

## Metadata Cheat Sheet

- **Notes field**: includes critic score, whether Rotten Tomatoes marked it “Certified Fresh,” and the canonical URL.
- **Rating column**: reserved for future personal ratings; right now it mirrors the upstream critic context.
- **Timestamps**: `added` when you saved it, `watched` when you complete it.

To keep a running diary, drop the IDs into a note:

```markdown
### Queue Ideas
- Movie #31 → Watch before sequel comes out
- Movie #12 → Pair with article #5 for cyberpunk research
```

This keeps everything searchable without having to leave the terminal.
