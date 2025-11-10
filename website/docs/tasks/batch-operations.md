---
title: Batch Operations
sidebar_label: Batch Ops
description: Use shell scripting patterns for mass task edits.
sidebar_position: 8
---

# Batch Operations

While Noteleaf doesn't have built-in bulk update commands, you can use shell scripting for batch operations:

**Complete all tasks in a project**:

```sh
noteleaf task list --project old-project --static | \
  awk '{print $1}' | \
  xargs -I {} noteleaf task done {}
```

**Add tag to multiple tasks**:

```sh
for id in 1 2 3 4 5; do
  noteleaf task update $id --add-tag urgent
done
```
