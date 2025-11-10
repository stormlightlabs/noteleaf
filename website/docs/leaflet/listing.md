---
title: Listing and Reading Publications
sidebar_label: Listing & Reading
description: Browse leaflet-backed notes and read content locally.
sidebar_position: 7
---

# Listing and Reading Publications

## List Published Documents

**All leaflet-synced notes**:

```sh
noteleaf pub list
```

**Only published documents**:

```sh
noteleaf pub list --published
```

**Only drafts**:

```sh
noteleaf pub list --draft
```

**Interactive browser**:

```sh
noteleaf pub list --interactive
```

Navigate with arrow keys, press Enter to read, `q` to quit.

## Reading a Publication

**Read specific document**:

```sh
noteleaf pub read 123
```

The identifier can be:

- Note ID (e.g., `123`)
- Leaflet record key (rkey, e.g., `3jxx...`)

**Read newest publication**:

```sh
noteleaf pub read
```

Omitting the identifier shows the most recently published document.
