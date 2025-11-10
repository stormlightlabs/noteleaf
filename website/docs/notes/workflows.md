---
title: Note-Taking Workflows
sidebar_label: Workflows
description: Zettelkasten, meeting notes, daily notes, and more.
sidebar_position: 7
---

# Note-Taking Workflows

## Zettelkasten

Zettelkasten emphasizes atomic notes with heavy linking:

1. **Create atomic notes**: Each note covers one concept
2. **Add descriptive tags**: Use tags for categorization
3. **Link related notes**: Reference other notes by title
4. **Develop ideas over time**: Expand notes with new insights

Example:

```sh
noteleaf note create "Dependency Injection" --tags architecture,patterns
noteleaf note create "Inversion of Control" --tags architecture,patterns
# In each note, reference the other
```

### Research

For academic or technical research:

1. **Source note per paper/article**: Create note for each source
2. **Extract key points**: Summarize in your own words
3. **Tag by topic**: Use consistent tags across research area
4. **Link to related work**: Reference other sources

Example:

```sh
noteleaf note create "Paper: Microservices Patterns" \
  --tags research,architecture,microservices
```

### Meeting

Capture discussions and action items:

1. **Template-based**: Use meeting_note function from earlier
2. **Consistent structure**: Attendees, agenda, discussion, actions
3. **Action items**: Extract as tasks for follow-up
4. **Link to projects**: Tag with project name

Example:

```sh
meeting_note "Sprint Planning"
# Then extract action items as tasks
noteleaf task add "Implement auth endpoint" --project web-service
```

### Daily

Journal-style daily entries:

1. **Daily template**: Use daily_note function
2. **Reflect on work**: What was accomplished, what's next
3. **Capture ideas**: Random thoughts for later processing
4. **Review weekly**: Scan week's notes for patterns

Example:

```sh
daily_note
# Creates note tagged with 'daily' and today's date
```

### Personal Knowledge Base

Build a reference library:

1. **How-to guides**: Document procedures and commands
2. **Troubleshooting notes**: Solutions to problems encountered
3. **Concept explanations**: Notes on topics you're learning
4. **Snippets**: Code examples and configurations

Use tags like: `how-to`, `troubleshooting`, `reference`, `snippet`
