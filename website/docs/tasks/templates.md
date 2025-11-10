---
title: Task Templates
sidebar_label: Templates
description: Shell helpers for creating consistent task structures.
sidebar_position: 9
---

# Task Templates

While templates aren't built-in, you can create shell functions for common task patterns:

```sh
# In your ~/.bashrc or ~/.zshrc
bug() {
  noteleaf task add "$1" \
    --project $(git rev-parse --show-toplevel | xargs basename) \
    --tags bug \
    --priority high
}

meeting() {
  noteleaf task add "$1" \
    --project work \
    --context @office \
    --recur "FREQ=WEEKLY;BYDAY=$2"
}
```

Usage:

```sh
bug "Fix login redirect"
meeting "Sprint planning" "MO"
```
