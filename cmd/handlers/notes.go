package handlers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// NoteHandler handles all note-related commands
type NoteHandler struct {
	db               *store.Database
	config           *store.Config
	repos            *repo.Repositories
	openInEditorFunc editorFunc
}

// NewNoteHandler creates a new note handler
func NewNoteHandler() (*NoteHandler, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	repos := repo.NewRepositories(db.DB)

	return &NoteHandler{
		db:     db,
		config: config,
		repos:  repos,
	}, nil
}

// Close cleans up resources
func (h *NoteHandler) Close() error {
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

// Create handles note creation subcommands
func Create(ctx context.Context, args []string) error {
	handler, err := NewNoteHandler()
	if err != nil {
		return err
	}
	defer handler.Close()

	if len(args) == 0 {
		return handler.createInteractive(ctx)
	}

	if len(args) == 1 && isFile(args[0]) {
		return handler.createFromFile(ctx, args[0])
	}

	title := args[0]
	content := ""
	if len(args) > 1 {
		content = strings.Join(args[1:], " ")
	}

	return handler.createFromArgs(ctx, title, content)
}

// New is an alias for Create
func New(ctx context.Context, args []string) error {
	return Create(ctx, args)
}

func (h *NoteHandler) createInteractive(ctx context.Context) error {
	logger := utils.GetLogger()

	tempFile, err := os.CreateTemp("", "noteleaf-note-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	template := `# New Note

Enter your note content here...

<!-- Tags: personal, work -->
`
	if _, err := tempFile.WriteString(template); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}
	tempFile.Close()

	editor := h.getEditor()
	if editor == "" {
		return fmt.Errorf("no editor configured. Set EDITOR environment variable or configure editor in settings")
	}

	logger.Info("Opening editor", "editor", editor, "file", tempFile.Name())
	if err := h.openInEditor(editor, tempFile.Name()); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited content: %w", err)
	}

	contentStr := string(content)
	if strings.TrimSpace(contentStr) == strings.TrimSpace(template) {
		fmt.Println("Note creation cancelled (no changes made)")
		return nil
	}

	title, noteContent, tags := h.parseNoteContent(contentStr)
	if title == "" {
		title = "Untitled Note"
	}

	note := &models.Note{
		Title:   title,
		Content: noteContent,
		Tags:    tags,
	}

	id, err := h.repos.Notes.Create(ctx, note)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	fmt.Printf("Created note: %s (ID: %d)\n", title, id)
	if len(tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
	}

	return nil
}

func (h *NoteHandler) createFromFile(ctx context.Context, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	if strings.TrimSpace(contentStr) == "" {
		return fmt.Errorf("file is empty: %s", filePath)
	}

	title, noteContent, tags := h.parseNoteContent(contentStr)
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	note := &models.Note{
		Title:    title,
		Content:  noteContent,
		Tags:     tags,
		FilePath: filePath,
	}

	id, err := h.repos.Notes.Create(ctx, note)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	fmt.Printf("Created note from file: %s\n", filePath)
	fmt.Printf("Note: %s (ID: %d)\n", title, id)
	if len(tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
	}

	return nil
}

func (h *NoteHandler) createFromArgs(ctx context.Context, title, content string) error {
	note := &models.Note{
		Title:   title,
		Content: content,
	}

	id, err := h.repos.Notes.Create(ctx, note)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	fmt.Printf("Created note: %s (ID: %d)\n", title, id)

	editor := h.getEditor()
	if editor != "" {
		fmt.Print("Open in editor? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			return h.editNote(ctx, id)
		}
	}

	return nil
}

func (h *NoteHandler) editNote(ctx context.Context, id int64) error {
	note, err := h.repos.Notes.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	tempFile, err := os.CreateTemp("", fmt.Sprintf("noteleaf-note-%d-*.md", id))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	fullContent := h.formatNoteForEdit(note)
	if _, err := tempFile.WriteString(fullContent); err != nil {
		return fmt.Errorf("failed to write note content: %w", err)
	}
	tempFile.Close()

	editor := h.getEditor()
	if err := h.openInEditor(editor, tempFile.Name()); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	editedContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited content: %w", err)
	}

	editedStr := string(editedContent)
	if editedStr == fullContent {
		fmt.Println("No changes made")
		return nil
	}

	title, content, tags := h.parseNoteContent(editedStr)
	if title == "" {
		title = note.Title
	}
	note.Title = title
	note.Content = content
	note.Tags = tags

	if err := h.repos.Notes.Update(ctx, note); err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	fmt.Printf("Updated note: %s (ID: %d)\n", title, id)
	return nil
}

func (h *NoteHandler) getEditor() string {
	// TODO: Add editor to config structure
	// For now, check environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	editors := []string{"vim", "nano", "code", "emacs"}
	for _, editor := range editors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	return ""
}

type editorFunc func(editor, filePath string) error

func defaultOpenInEditor(editor, filePath string) error {
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (h *NoteHandler) openInEditor(editor, filePath string) error {
	if h.openInEditorFunc != nil {
		return h.openInEditorFunc(editor, filePath)
	}
	return defaultOpenInEditor(editor, filePath)
}

func (h *NoteHandler) parseNoteContent(content string) (title, noteContent string, tags []string) {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "<!-- Tags:") && strings.HasSuffix(line, "-->") {
			tagStr := strings.TrimPrefix(line, "<!-- Tags:")
			tagStr = strings.TrimSuffix(tagStr, "-->")
			tagStr = strings.TrimSpace(tagStr)

			if tagStr != "" {
				for _, tag := range strings.Split(tagStr, ",") {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						tags = append(tags, tag)
					}
				}
			}
		}
	}

	noteContent = content

	return title, noteContent, tags
}

func (h *NoteHandler) formatNoteForEdit(note *models.Note) string {
	var content strings.Builder

	if !strings.Contains(note.Content, "# "+note.Title) {
		content.WriteString("# " + note.Title + "\n\n")
	}

	content.WriteString(note.Content)

	if len(note.Tags) > 0 {
		if !strings.HasSuffix(note.Content, "\n") {
			content.WriteString("\n")
		}
		content.WriteString("\n<!-- Tags: " + strings.Join(note.Tags, ", ") + " -->\n")
	}

	return content.String()
}

func isFile(arg string) bool {
	if filepath.Ext(arg) != "" {
		return true
	}

	if info, err := os.Stat(arg); err == nil && !info.IsDir() {
		return true
	}

	return strings.Contains(arg, "/") || strings.Contains(arg, "\\")
}
