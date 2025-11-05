# Article Examples

Examples of saving and managing articles using Noteleaf.

## Saving Articles

### Save Article from URL

```sh
noteleaf article add https://example.com/interesting-article
```

### Save with Custom Title

```sh
noteleaf article add https://example.com/article --title "My Custom Title"
```

### Save Multiple Articles

```sh
noteleaf article add https://example.com/article-1
noteleaf article add https://example.com/article-2
noteleaf article add https://example.com/article-3
```

## Viewing Articles

### List All Articles

```sh
noteleaf article list
```

### Filter by Author

```sh
noteleaf article list --author "Jane Smith"
noteleaf article list --author "John Doe"
```

### Filter with Query

```sh
noteleaf article list --query "golang"
noteleaf article list --query "machine learning"
```

### Limit Results

```sh
noteleaf article list --limit 10
noteleaf article list --limit 5
```

### View Article Content

Display in terminal:

```sh
noteleaf article view 1
```

Read in browser:

```sh
noteleaf article read 1
```

## Managing Articles

### Update Article Metadata

Update title:

```sh
noteleaf article update 1 --title "Updated Title"
```

Update author:

```sh
noteleaf article update 1 --author "Jane Doe"
```

Add notes:

```sh
noteleaf article update 1 --notes "Great insights on API design"
```

### Remove Article

```sh
noteleaf article remove 1
```

Remove multiple:

```sh
noteleaf article remove 1 2 3
```

## Common Workflows

### Reading List Management

Save articles to read later:

```sh
noteleaf article add https://blog.example.com/microservices
noteleaf article add https://dev.to/understanding-async
noteleaf article add https://medium.com/best-practices

# View reading list
noteleaf article list
```

### Research Collection

Collect articles for research:

```sh
# Save research articles
noteleaf article add https://arxiv.org/paper1 --notes "Research: ML optimization"
noteleaf article add https://arxiv.org/paper2 --notes "Research: Neural networks"

# Find research articles
noteleaf article list --query "Research"
```

### Technical Documentation

Archive technical articles:

```sh
noteleaf article add https://docs.example.com/api-guide --notes "Category: Documentation"
noteleaf article add https://tutorials.example.com/setup --notes "Category: Tutorial"

# Find documentation
noteleaf article list --query "Documentation"
```

### Author Following

Track articles by favorite authors:

```sh
# Save articles
noteleaf article add https://blog.author1.com/post1 --author "Martin Fowler"
noteleaf article add https://blog.author1.com/post2 --author "Martin Fowler"

# View articles by author
noteleaf article list --author "Martin Fowler"
```

### Topic Collections

Organize by topic using notes:

```sh
# Backend articles
noteleaf article add https://example.com/databases --notes "Topic: Backend, Database"
noteleaf article add https://example.com/caching --notes "Topic: Backend, Performance"

# Frontend articles
noteleaf article add https://example.com/react --notes "Topic: Frontend, React"
noteleaf article add https://example.com/css --notes "Topic: Frontend, CSS"

# Find by topic
noteleaf article list --query "Backend"
noteleaf article list --query "Frontend"
```

### Daily Reading Routine

Save articles during the day:

```sh
# Morning
noteleaf article add https://news.ycombinator.com/article1
noteleaf article add https://reddit.com/r/programming/article2

# Evening - review saved articles
noteleaf article list
noteleaf article view 1
noteleaf article view 2
```

### Offline Reading

Save articles for offline access:

```sh
# Save articles before travel
noteleaf article add https://longform.com/essay1
noteleaf article add https://magazine.com/feature

# Read offline (articles are saved locally)
noteleaf article view 1
noteleaf article view 2
```

### Archive and Cleanup

Remove read articles:

```sh
# List articles
noteleaf article list

# Remove articles you've read
noteleaf article remove 1 2 3 4 5

# Keep only recent articles (manual filtering)
noteleaf article list --limit 20
```

### Share-worthy Content

Mark articles worth sharing:

```sh
noteleaf article add https://excellent.article.com --notes "Share: Twitter, Newsletter"
noteleaf article add https://must-read.com/post --notes "Share: Team, Blog"

# Find share-worthy articles
noteleaf article list --query "Share"
```

### Learning Path

Create structured learning collections:

```sh
# Beginner articles
noteleaf article add https://tutorial.com/intro --notes "Level: Beginner, Go"
noteleaf article add https://tutorial.com/basics --notes "Level: Beginner, Go"

# Advanced articles
noteleaf article add https://advanced.com/patterns --notes "Level: Advanced, Go"

# Find by level
noteleaf article list --query "Beginner"
noteleaf article list --query "Advanced"
```

### Weekly Digests

Collect interesting articles weekly:

```sh
# Week 1
noteleaf article add https://example.com/week1-1 --notes "Week: 2024-W01"
noteleaf article add https://example.com/week1-2 --notes "Week: 2024-W01"

# Week 2
noteleaf article add https://example.com/week2-1 --notes "Week: 2024-W02"

# Review week's articles
noteleaf article list --query "2024-W01"
```

## Exporting Articles

### Export Article to File

```sh
noteleaf article export 1 --format markdown > article.md
noteleaf article export 1 --format html > article.html
```

### Export Multiple Articles

```sh
noteleaf article export --all --format markdown --output articles/
```

### Export by Query

```sh
noteleaf article export --query "golang" --format markdown --output go-articles/
```

## Integration with Notes

### Create Note from Article

After reading:

```sh
# Read article
noteleaf article view 1

# Create summary note
noteleaf note create "Article Summary: Title" "
Source: [Article URL]
Author: [Author Name]

Key Points:
- Point 1
- Point 2

My Thoughts:
- Observation 1
- Observation 2
" --tags article-summary,topic
```

### Link Article to Task

Create follow-up task:

```sh
# Save article
noteleaf article add https://example.com/implement-feature

# Create related task
noteleaf task add "Implement feature from article #1" --tags implementation
```

## Article Statistics

### Count Articles

```sh
noteleaf article list | wc -l
```

### Articles by Author

```sh
noteleaf article list --author "Author Name" | wc -l
```

### Articles by Topic

```sh
noteleaf article list --query "topic" | wc -l
```
