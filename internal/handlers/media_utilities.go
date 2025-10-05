package handlers

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// MediaPrinter defines how to format a media item for display
type MediaPrinter[T any] func(item *T)

// ListMediaItems is a generic utility for listing media items with status filtering
func ListMediaItems[T any](
	ctx context.Context,
	status string,
	mediaType string,
	listAll func(ctx context.Context) ([]*T, error),
	listByStatus func(ctx context.Context, status string) ([]*T, error),
	printer MediaPrinter[T],
) error {
	var items []*T
	var err error

	if status == "" {
		items, err = listAll(ctx)
		if err != nil {
			return fmt.Errorf("failed to list %s: %w", mediaType, err)
		}
	} else {
		items, err = listByStatus(ctx, status)
		if err != nil {
			return fmt.Errorf("failed to get %s %s: %w", status, mediaType, err)
		}
	}

	if len(items) == 0 {
		if status == "" {
			fmt.Printf("No %s found\n", mediaType)
		} else {
			fmt.Printf("No %s %s found\n", status, mediaType)
		}
		return nil
	}

	fmt.Printf("Found %d %s:\n\n", len(items), mediaType)
	for _, item := range items {
		printer(item)
	}

	return nil
}

// PromptUserChoice prompts the user to select from a list of results
func PromptUserChoice(reader io.Reader, maxChoices int) (int, error) {
	fmt.Print("\nEnter number to add (1-", maxChoices, "), or 0 to cancel: ")

	var choice int
	if reader != nil {
		if _, err := fmt.Fscanf(reader, "%d", &choice); err != nil {
			return 0, fmt.Errorf("invalid input")
		}
	} else {
		if _, err := fmt.Scanf("%d", &choice); err != nil {
			return 0, fmt.Errorf("invalid input")
		}
	}

	if choice == 0 {
		fmt.Println("Cancelled.")
		return 0, nil
	}
	if choice < 1 || choice > maxChoices {
		return 0, fmt.Errorf("invalid choice: %d", choice)
	}
	return choice, nil
}

// ParseID converts a string ID to int64
func ParseID(id string, itemType string) (int64, error) {
	itemID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s ID: %s", itemType, id)
	}
	return itemID, nil
}

// PrintSearchResults displays search results with a type-specific formatter
func PrintSearchResults[T models.Model](results []*models.Model, formatter func(*models.Model, int)) error {
	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("Found %d result(s):\n\n", len(results))
	for i, result := range results {
		formatter(result, i+1)
	}
	return nil
}
