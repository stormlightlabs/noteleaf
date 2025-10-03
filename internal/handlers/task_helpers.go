package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
)

// ParsedTaskData holds extracted metadata from a task description
type ParsedTaskData struct {
	Description string
	Project     string
	Context     string
	Tags        []string
	Due         string
	Recur       string
	Until       string
	ParentUUID  string
	DependsOn   []string
}

// parseDescription extracts inline metadata from description text
// Supports: +project @context #tag due:YYYY-MM-DD recur:RULE until:DATE parent:UUID depends:UUID1,UUID2
func parseDescription(text string) *ParsedTaskData {
	parsed := &ParsedTaskData{Tags: []string{}, DependsOn: []string{}}
	words := strings.Fields(text)

	var descWords []string
	for _, word := range words {
		switch {
		case strings.HasPrefix(word, "+"):
			parsed.Project = strings.TrimPrefix(word, "+")
		case strings.HasPrefix(word, "@"):
			parsed.Context = strings.TrimPrefix(word, "@")
		case strings.HasPrefix(word, "#"):
			parsed.Tags = append(parsed.Tags, strings.TrimPrefix(word, "#"))
		case strings.HasPrefix(word, "due:"):
			parsed.Due = strings.TrimPrefix(word, "due:")
		case strings.HasPrefix(word, "recur:"):
			parsed.Recur = strings.TrimPrefix(word, "recur:")
		case strings.HasPrefix(word, "until:"):
			parsed.Until = strings.TrimPrefix(word, "until:")
		case strings.HasPrefix(word, "parent:"):
			parsed.ParentUUID = strings.TrimPrefix(word, "parent:")
		case strings.HasPrefix(word, "depends:"):
			deps := strings.TrimPrefix(word, "depends:")
			parsed.DependsOn = strings.Split(deps, ",")
		default:
			descWords = append(descWords, word)
		}
	}

	parsed.Description = strings.Join(descWords, " ")
	return parsed
}

func removeString(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func pluralize(count int) string {
	rule := plural.Cardinal.MatchPlural(language.English, count, 0, 0, 0, 0)
	switch rule {
	case plural.One:
		return ""
	default:
		return "s"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	hours := d.Hours()
	if hours < 24 {
		return fmt.Sprintf("%.1fh", hours)
	}
	days := int(hours / 24)
	if remainingHours := hours - float64(days*24); remainingHours == 0 {
		return fmt.Sprintf("%dd", days)
	} else {
		return fmt.Sprintf("%dd %.1fh", days, remainingHours)
	}
}

func printTask(task *models.Task) {
	fmt.Printf("[%d] %s", task.ID, task.Description)

	if task.Status != "pending" {
		fmt.Printf(" (%s)", task.Status)
	}

	if task.Priority != "" {
		fmt.Printf(" [%s]", task.Priority)
	}

	if task.Project != "" {
		fmt.Printf(" +%s", task.Project)
	}

	if task.Context != "" {
		fmt.Printf(" @%s", task.Context)
	}

	if len(task.Tags) > 0 {
		fmt.Printf(" #%s", strings.Join(task.Tags, " #"))
	}

	if task.Due != nil {
		fmt.Printf(" (due: %s)", task.Due.Format("2006-01-02"))
	}

	if task.Recur != "" {
		fmt.Printf(" \u21bb")
	}

	if len(task.DependsOn) > 0 {
		fmt.Printf(" \u2937%d", len(task.DependsOn))
	}

	fmt.Println()
}

func printTaskDetail(task *models.Task, noMetadata bool) {
	fmt.Printf("Task ID: %d\n", task.ID)
	fmt.Printf("UUID: %s\n", task.UUID)
	fmt.Printf("Description: %s\n", task.Description)
	fmt.Printf("Status: %s\n", task.Status)

	if task.Priority != "" {
		fmt.Printf("Priority: %s\n", task.Priority)
	}

	if task.Project != "" {
		fmt.Printf("Project: %s\n", task.Project)
	}

	if task.Context != "" {
		fmt.Printf("Context: %s\n", task.Context)
	}

	if len(task.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(task.Tags, ", "))
	}

	if task.Due != nil {
		fmt.Printf("Due: %s\n", task.Due.Format("2006-01-02 15:04"))
	}

	if task.Recur != "" {
		fmt.Printf("Recurrence: %s\n", task.Recur)
	}

	if task.Until != nil {
		fmt.Printf("Recur Until: %s\n", task.Until.Format("2006-01-02"))
	}

	if task.ParentUUID != nil {
		fmt.Printf("Parent Task: %s\n", *task.ParentUUID)
	}

	if len(task.DependsOn) > 0 {
		fmt.Printf("Depends On:\n")
		for _, dep := range task.DependsOn {
			fmt.Printf("  - %s\n", dep)
		}
	}

	if !noMetadata {
		fmt.Printf("Created: %s\n", task.Entry.Format("2006-01-02 15:04"))
		fmt.Printf("Modified: %s\n", task.Modified.Format("2006-01-02 15:04"))

		if task.Start != nil {
			fmt.Printf("Started: %s\n", task.Start.Format("2006-01-02 15:04"))
		}

		if task.End != nil {
			fmt.Printf("Completed: %s\n", task.End.Format("2006-01-02 15:04"))
		}
	}

	if len(task.Annotations) > 0 {
		fmt.Printf("Annotations:\n")
		for _, annotation := range task.Annotations {
			fmt.Printf("  - %s\n", annotation)
		}
	}
}

func printTaskJSON(task *models.Task) error {
	if data, err := json.MarshalIndent(task, "", "  "); err != nil {
		return fmt.Errorf("failed to marshal task to JSON: %w", err)
	} else {
		fmt.Println(string(data))
		return nil
	}
}
