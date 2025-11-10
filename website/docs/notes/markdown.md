---
title: Markdown Support
sidebar_label: Markdown
description: Reference for GitHub-Flavored Markdown features.
sidebar_position: 6
---

# Markdown Support

Noteleaf notes support full GitHub-Flavored Markdown:

## Headers

```markdown
# Level 1
## Level 2
### Level 3
```

## Text Formatting

```markdown
**bold**
*italic*
***bold and italic***
~~strikethrough~~
`inline code`
```

## Lists

```markdown
- Unordered item
- Another item
  - Nested item

1. Ordered item
2. Second item
   1. Nested ordered
```

## Task Lists

```markdown
- [x] Completed task
- [ ] Pending task
- [ ] Another pending
```

## Links

```markdown
[Link text](https://example.com)
[Reference link][ref]

[ref]: https://example.com
```

## Images

```markdown
![Alt text](path/to/image.png)
![Remote image](https://example.com/image.png)
```

## Code Blocks

````markdown
```python
def hello():
    print("Hello, world!")
```

```javascript
const greet = () => console.log("Hello!");
```
````

## Tables

```markdown
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |
```

## Blockquotes

```markdown
> This is a quote
> spanning multiple lines
>
> With multiple paragraphs
```

## Horizontal Rules

```markdown
---
***
___
```
