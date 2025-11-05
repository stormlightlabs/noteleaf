# Note Examples

Examples of note-taking workflows using Noteleaf.

## Creating Notes

### Create Note from Command Line

```sh
noteleaf note create "Meeting Notes" "Discussed Q4 roadmap and priorities"
```

### Create Note with Tags

```sh
noteleaf note create "API Design Ideas" "REST vs GraphQL considerations" --tags api,design
```

### Create Note from File

```sh
noteleaf note create --file notes.md
```

### Create Note Interactively

Opens your editor for composition:

```sh
noteleaf note create --interactive
```

Specify editor:

```sh
EDITOR=vim noteleaf note create --interactive
```

### Create Note with Multiple Paragraphs

```sh
noteleaf note create "Project Retrospective" "
What went well:
- Good team collaboration
- Met all deadlines
- Quality code reviews

What to improve:
- Better documentation
- More automated tests
- Earlier stakeholder feedback
"
```

## Viewing Notes

### List All Notes

Interactive mode:

```sh
noteleaf note list
```

Static output:

```sh
noteleaf note list --static
```

### Filter by Tags

```sh
noteleaf note list --tags meeting
noteleaf note list --tags api,design
```

### View Archived Notes

```sh
noteleaf note list --archived
```

### Read a Note

Display note content:

```sh
noteleaf note read 1
```

### Search Notes

```sh
noteleaf note search "API design"
noteleaf note search "meeting notes"
```

## Editing Notes

### Edit Note in Editor

Opens note in your editor:

```sh
noteleaf note edit 1
```

With specific editor:

```sh
EDITOR=nvim noteleaf note edit 1
```

### Update Note Title

```sh
noteleaf note update 1 --title "Updated Meeting Notes"
```

### Add Tags to Note

```sh
noteleaf note tag 1 --add important,todo
```

### Remove Tags from Note

```sh
noteleaf note tag 1 --remove draft
```

## Organizing Notes

### Archive a Note

```sh
noteleaf note archive 1
```

### Unarchive a Note

```sh
noteleaf note unarchive 1
```

### Delete a Note

```sh
noteleaf note remove 1
```

With confirmation:

```sh
noteleaf note remove 1 --confirm
```

## Common Workflows

### Quick Meeting Notes

```sh
noteleaf note create "Team Standup $(date +%Y-%m-%d)" --interactive --tags meeting,standup
```

### Project Documentation

```sh
noteleaf note create "Project Architecture" "$(cat architecture.md)" --tags docs,architecture
```

### Research Notes

Create research note:

```sh
noteleaf note create "GraphQL Research" --interactive --tags research,api
```

List all research notes:

```sh
noteleaf note list --tags research
```

### Code Snippets

```sh
noteleaf note create "Useful Git Commands" "
# Rebase last 3 commits
git rebase -i HEAD~3

# Undo last commit
git reset --soft HEAD~1

# Show files changed in commit
git show --name-only <commit>
" --tags git,snippets,reference
```

### Daily Journal

```sh
noteleaf note create "Journal $(date +%Y-%m-%d)" --interactive --tags journal
```

### Ideas and Brainstorming

```sh
noteleaf note create "Product Ideas" --interactive --tags ideas,product
```

List all ideas:

```sh
noteleaf note list --tags ideas
```

## Exporting Notes

### Export Single Note

```sh
noteleaf note export 1 --format markdown > note.md
noteleaf note export 1 --format html > note.html
```

### Export All Notes

```sh
noteleaf note export --all --format markdown --output notes/
```

### Export Notes by Tag

```sh
noteleaf note export --tags meeting --format markdown --output meetings/
```

## Advanced Usage

### Template-based Notes

Create a note template file:

```sh
cat > ~/templates/meeting.md << 'EOF'
# Meeting: [TITLE]
Date: [DATE]
Attendees: [NAMES]

## Agenda
-

## Discussion
-

## Action Items
- [ ]

## Next Meeting
Date:
EOF
```

Use template:

```sh
noteleaf note create --file ~/templates/meeting.md
```

### Linking Notes

Reference other notes in content:

```sh
noteleaf note create "Implementation Plan" "
Based on the design in Note #5, we will:
1. Set up database schema (see Note #12)
2. Implement API endpoints
3. Add frontend components

Related: Note #5 (Design), Note #12 (Schema)
" --tags implementation,plan
```

### Note Statistics

View note count by tag:

```sh
noteleaf note list --static | grep -c "tag:meeting"
```
