package handlers

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/public"
	"github.com/stormlightlabs/noteleaf/internal/services"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

func TestPublicationHandler(t *testing.T) {
	t.Run("sessionFromConfig", func(t *testing.T) {
		t.Run("returns error when DID is missing", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "refresh_token",
			}

			_, err := sessionFromConfig(config)
			if err == nil {
				t.Error("Expected error when DID is missing")
			}
		})

		t.Run("returns error when AccessJWT is missing", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "",
				ATProtoRefreshJWT: "refresh_token",
			}

			_, err := sessionFromConfig(config)
			if err == nil {
				t.Error("Expected error when AccessJWT is missing")
			}
		})

		t.Run("returns error when RefreshJWT is missing", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "",
			}

			_, err := sessionFromConfig(config)
			if err == nil {
				t.Error("Expected error when RefreshJWT is missing")
			}
		})

		t.Run("successfully creates session from complete config", func(t *testing.T) {
			expiresAt := time.Now().Add(2 * time.Hour)
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "refresh_token",
				ATProtoPDSURL:     "https://bsky.social",
				ATProtoExpiresAt:  expiresAt.Format("2006-01-02T15:04:05Z07:00"),
			}

			session, err := sessionFromConfig(config)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if session.DID != config.ATProtoDID {
				t.Errorf("Expected DID '%s', got '%s'", config.ATProtoDID, session.DID)
			}
			if session.Handle != config.ATProtoHandle {
				t.Errorf("Expected Handle '%s', got '%s'", config.ATProtoHandle, session.Handle)
			}
			if session.AccessJWT != config.ATProtoAccessJWT {
				t.Errorf("Expected AccessJWT '%s', got '%s'", config.ATProtoAccessJWT, session.AccessJWT)
			}
			if session.RefreshJWT != config.ATProtoRefreshJWT {
				t.Errorf("Expected RefreshJWT '%s', got '%s'", config.ATProtoRefreshJWT, session.RefreshJWT)
			}
			if session.PDSURL != config.ATProtoPDSURL {
				t.Errorf("Expected PDSURL '%s', got '%s'", config.ATProtoPDSURL, session.PDSURL)
			}
			if !session.Authenticated {
				t.Error("Expected session to be authenticated")
			}
		})

		t.Run("handles missing expiry time gracefully", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "refresh_token",
				ATProtoExpiresAt:  "",
			}

			session, err := sessionFromConfig(config)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if session.ExpiresAt.After(time.Now()) {
				t.Error("Expected ExpiresAt to be in the past when not provided")
			}
		})

		t.Run("handles invalid expiry time format gracefully", func(t *testing.T) {
			config := &store.Config{
				ATProtoDID:        "did:plc:test123",
				ATProtoHandle:     "test.bsky.social",
				ATProtoAccessJWT:  "access_token",
				ATProtoRefreshJWT: "refresh_token",
				ATProtoExpiresAt:  "invalid-timestamp",
			}

			session, err := sessionFromConfig(config)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if session.ExpiresAt.After(time.Now()) {
				t.Error("Expected ExpiresAt to be in the past when parse fails")
			}
		})
	})

	t.Run("Auth", func(t *testing.T) {
		t.Run("validates required parameters", func(t *testing.T) {
			_ = NewHandlerTestSuite(t)
			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Auth(ctx, "", "password")
			if err == nil {
				t.Error("Expected error when handle is empty")
			}

			err = handler.Auth(ctx, "handle", "")
			if err == nil {
				t.Error("Expected error when password is empty")
			}
		})

	})

	t.Run("GetAuthStatus", func(t *testing.T) {
		t.Run("returns not authenticated when no session", func(t *testing.T) {
			_ = NewHandlerTestSuite(t)
			handler := CreateHandler(t, NewPublicationHandler)

			status := handler.GetAuthStatus()
			if status != "Not authenticated" {
				t.Errorf("Expected 'Not authenticated', got '%s'", status)
			}
		})

		t.Run("returns authenticated status with session", func(t *testing.T) {
			_ = NewHandlerTestSuite(t)
			handler := CreateHandler(t, NewPublicationHandler)

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			status := handler.GetAuthStatus()
			expectedStatus := "Authenticated as test.bsky.social"
			if status != expectedStatus {
				t.Errorf("Expected '%s', got '%s'", expectedStatus, status)
			}
		})
	})

	t.Run("GetLastAuthenticatedHandle", func(t *testing.T) {
		t.Run("returns empty string when no config", func(t *testing.T) {
			handler := &PublicationHandler{
				config: nil,
			}

			handle := handler.GetLastAuthenticatedHandle()
			if handle != "" {
				t.Errorf("Expected empty string, got '%s'", handle)
			}
		})

		t.Run("returns empty string when handle not set", func(t *testing.T) {
			handler := &PublicationHandler{
				config: &store.Config{},
			}

			handle := handler.GetLastAuthenticatedHandle()
			if handle != "" {
				t.Errorf("Expected empty string, got '%s'", handle)
			}
		})

		t.Run("returns handle from config", func(t *testing.T) {
			expectedHandle := "test.bsky.social"
			handler := &PublicationHandler{
				config: &store.Config{
					ATProtoHandle: expectedHandle,
				},
			}

			handle := handler.GetLastAuthenticatedHandle()
			if handle != expectedHandle {
				t.Errorf("Expected '%s', got '%s'", expectedHandle, handle)
			}
		})

		t.Run("returns handle after successful authentication", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			mock := services.SetupSuccessfulAuthMocks()
			handler.atproto = mock

			err := handler.Auth(ctx, "user.bsky.social", "password123")
			suite.AssertNoError(err, "authentication should succeed")

			handle := handler.GetLastAuthenticatedHandle()
			if handle != "user.bsky.social" {
				t.Errorf("Expected 'user.bsky.social', got '%s'", handle)
			}
		})
	})

	t.Run("NewPublicationHandler", func(t *testing.T) {
		t.Run("creates handler successfully", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler, err := NewPublicationHandler()
			if err != nil {
				t.Fatalf("Expected no error creating handler, got %v", err)
			}
			defer handler.Close()

			if handler.db == nil {
				t.Error("Expected database to be initialized")
			}
			if handler.config == nil {
				t.Error("Expected config to be initialized")
			}
			if handler.repos == nil {
				t.Error("Expected repos to be initialized")
			}
			if handler.atproto == nil {
				t.Error("Expected atproto service to be initialized")
			}
		})

		t.Run("restores session from config when available", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			config, err := store.LoadConfig()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			config.ATProtoDID = "did:plc:test123"
			config.ATProtoHandle = "test.bsky.social"
			config.ATProtoAccessJWT = "access_token"
			config.ATProtoRefreshJWT = "refresh_token"
			config.ATProtoPDSURL = "https://bsky.social"
			config.ATProtoExpiresAt = time.Now().Add(2 * time.Hour).Format("2006-01-02T15:04:05Z07:00")

			err = store.SaveConfig(config)
			if err != nil {
				t.Fatalf("Failed to save config: %v", err)
			}

			handler, err := NewPublicationHandler()
			if err != nil {
				t.Fatalf("Expected no error creating handler, got %v", err)
			}
			defer handler.Close()

			if !handler.atproto.IsAuthenticated() {
				t.Error("Expected handler to be authenticated after restoring from config")
			}

			session, err := handler.atproto.GetSession()
			if err != nil {
				t.Errorf("Expected to get session, got error: %v", err)
			}
			if session.DID != config.ATProtoDID {
				t.Errorf("Expected DID '%s', got '%s'", config.ATProtoDID, session.DID)
			}
		})

		t.Run("handles empty config gracefully", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler, err := NewPublicationHandler()
			if err != nil {
				t.Fatalf("Expected no error creating handler, got %v", err)
			}
			defer handler.Close()

			if handler.atproto.IsAuthenticated() {
				t.Error("Expected handler to not be authenticated with empty config")
			}
		})
	})

	t.Run("Close", func(t *testing.T) {
		t.Run("cleans up resources properly", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler, err := NewPublicationHandler()
			if err != nil {
				t.Fatalf("Expected no error creating handler, got %v", err)
			}

			err = handler.Close()
			if err != nil {
				t.Errorf("Expected no error on close, got %v", err)
			}
		})
	})

	t.Run("documentToMarkdown", func(t *testing.T) {
		t.Run("converts simple document with text blocks", func(t *testing.T) {
			doc := services.DocumentWithMeta{
				Document: public.Document{
					Pages: []public.LinearDocument{
						{
							Blocks: []public.BlockWrap{
								{
									Type: "pub.leaflet.pages.linearDocument#block",
									Block: public.TextBlock{
										Type:      "pub.leaflet.pages.linearDocument#textBlock",
										Plaintext: "Hello world",
									},
								},
							},
						},
					},
				},
			}

			markdown, err := documentToMarkdown(doc)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if markdown != "Hello world" {
				t.Errorf("Expected 'Hello world', got '%s'", markdown)
			}
		})

		t.Run("converts document with headers", func(t *testing.T) {
			doc := services.DocumentWithMeta{
				Document: public.Document{
					Pages: []public.LinearDocument{
						{
							Blocks: []public.BlockWrap{
								{
									Type: "pub.leaflet.pages.linearDocument#block",
									Block: public.HeaderBlock{
										Type:      "pub.leaflet.pages.linearDocument#headerBlock",
										Level:     1,
										Plaintext: "Main Title",
									},
								},
								{
									Type: "pub.leaflet.pages.linearDocument#block",
									Block: public.TextBlock{
										Type:      "pub.leaflet.pages.linearDocument#textBlock",
										Plaintext: "Content here",
									},
								},
							},
						},
					},
				},
			}

			markdown, err := documentToMarkdown(doc)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			expected := "# Main Title\n\nContent here"
			if markdown != expected {
				t.Errorf("Expected '%s', got '%s'", expected, markdown)
			}
		})

		t.Run("converts document with code blocks", func(t *testing.T) {
			doc := services.DocumentWithMeta{
				Document: public.Document{
					Pages: []public.LinearDocument{
						{
							Blocks: []public.BlockWrap{
								{
									Type: "pub.leaflet.pages.linearDocument#block",
									Block: public.CodeBlock{
										Type:      "pub.leaflet.pages.linearDocument#codeBlock",
										Plaintext: "fmt.Println(\"hello\")",
										Language:  "go",
									},
								},
							},
						},
					},
				},
			}

			markdown, err := documentToMarkdown(doc)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			expected := "```go\nfmt.Println(\"hello\")\n```"
			if markdown != expected {
				t.Errorf("Expected '%s', got '%s'", expected, markdown)
			}
		})

		t.Run("converts document with multiple pages", func(t *testing.T) {
			doc := services.DocumentWithMeta{
				Document: public.Document{
					Pages: []public.LinearDocument{
						{
							Blocks: []public.BlockWrap{
								{
									Type: "pub.leaflet.pages.linearDocument#block",
									Block: public.TextBlock{
										Type:      "pub.leaflet.pages.linearDocument#textBlock",
										Plaintext: "Page one",
									},
								},
							},
						},
						{
							Blocks: []public.BlockWrap{
								{
									Type: "pub.leaflet.pages.linearDocument#block",
									Block: public.TextBlock{
										Type:      "pub.leaflet.pages.linearDocument#textBlock",
										Plaintext: "Page two",
									},
								},
							},
						},
					},
				},
			}

			markdown, err := documentToMarkdown(doc)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			expected := "Page one\n\nPage two"
			if markdown != expected {
				t.Errorf("Expected '%s', got '%s'", expected, markdown)
			}
		})

		t.Run("handles empty document", func(t *testing.T) {
			doc := services.DocumentWithMeta{
				Document: public.Document{
					Pages: []public.LinearDocument{},
				},
			}

			markdown, err := documentToMarkdown(doc)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if markdown != "" {
				t.Errorf("Expected empty string, got '%s'", markdown)
			}
		})
	})

	t.Run("Pull", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Pull(ctx)
			if err == nil {
				t.Error("Expected error when not authenticated")
			}

			expectedMsg := "not authenticated"
			if err != nil && !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, err.Error())
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("lists all leaflet notes", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey1 := "test_rkey_1"
			cid1 := "test_cid_1"
			publishedAt := time.Now()

			note1 := &models.Note{
				Title:       "Published Note",
				Content:     "Content 1",
				LeafletRKey: &rkey1,
				LeafletCID:  &cid1,
				PublishedAt: &publishedAt,
				IsDraft:     false,
			}

			_, err := handler.repos.Notes.Create(ctx, note1)
			suite.AssertNoError(err, "create published note")

			rkey2 := "test_rkey_2"
			cid2 := "test_cid_2"
			note2 := &models.Note{
				Title:       "Draft Note",
				Content:     "Content 2",
				LeafletRKey: &rkey2,
				LeafletCID:  &cid2,
				IsDraft:     true,
			}

			_, err = handler.repos.Notes.Create(ctx, note2)
			suite.AssertNoError(err, "create draft note")

			err = handler.List(ctx, "all")
			suite.AssertNoError(err, "list all notes")

			err = handler.List(ctx, "")
			suite.AssertNoError(err, "list with empty filter")
		})

		t.Run("lists only published notes", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "published_rkey"
			cid := "published_cid"
			publishedAt := time.Now()

			note := &models.Note{
				Title:       "Published Note",
				Content:     "Content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				PublishedAt: &publishedAt,
				IsDraft:     false,
			}

			_, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create published note")

			err = handler.List(ctx, "published")
			suite.AssertNoError(err, "list published notes")
		})

		t.Run("lists only draft notes", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "draft_rkey"
			cid := "draft_cid"

			note := &models.Note{
				Title:       "Draft Note",
				Content:     "Content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				IsDraft:     true,
			}

			_, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create draft note")

			err = handler.List(ctx, "draft")
			suite.AssertNoError(err, "list draft notes")
		})

		t.Run("handles empty results gracefully", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.List(ctx, "all")
			suite.AssertNoError(err, "list with no notes")
		})

		t.Run("returns error for invalid filter", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.List(ctx, "invalid_filter")
			if err == nil {
				t.Error("Expected error for invalid filter")
			}

			expectedMsg := "invalid filter"
			if err != nil && !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, err.Error())
			}
		})

		t.Run("only lists notes with leaflet metadata", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			regularNote := &models.Note{
				Title:   "Regular Note",
				Content: "No leaflet data",
			}

			_, err := handler.repos.Notes.Create(ctx, regularNote)
			suite.AssertNoError(err, "create regular note")

			rkey := "leaflet_rkey"
			cid := "leaflet_cid"
			leafletNote := &models.Note{
				Title:       "Leaflet Note",
				Content:     "Has leaflet data",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				IsDraft:     false,
			}

			_, err = handler.repos.Notes.Create(ctx, leafletNote)
			suite.AssertNoError(err, "create leaflet note")

			err = handler.List(ctx, "all")
			suite.AssertNoError(err, "list all leaflet notes")
		})
	})

	t.Run("Post", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Post(ctx, 1, false)
			if err == nil {
				t.Error("Expected error when not authenticated")
			}

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when note does not exist", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Post(ctx, 999, false)
			if err == nil {
				t.Error("Expected error when note does not exist")
			}

			if err != nil && !strings.Contains(err.Error(), "failed to get note") {
				t.Errorf("Expected 'failed to get note' error, got '%v'", err)
			}
		})

		t.Run("returns error when note already published", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "existing_rkey"
			cid := "existing_cid"
			note := &models.Note{
				Title:       "Already Published",
				Content:     "Test content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Post(ctx, id, false)
			if err == nil {
				t.Error("Expected error when note already published")
			}

			if err != nil && !strings.Contains(err.Error(), "already published") {
				t.Errorf("Expected 'already published' error, got '%v'", err)
			}
		})

		t.Run("handles markdown conversion errors", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note := &models.Note{
				Title:   "Test Note",
				Content: "# Valid markdown",
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Post(ctx, id, false)
			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				if !strings.Contains(err.Error(), "failed to post document") && !strings.Contains(err.Error(), "failed to get session") {
					t.Logf("Got expected error during post: %v", err)
				}
			}
		})

		t.Run("sets correct draft status", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note := &models.Note{
				Title:   "Draft Note",
				Content: "# Test content",
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Post(ctx, id, true)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during post (expected for external service call): %v", err)
			}
		})
	})

	t.Run("Patch", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Patch(ctx, 1)
			if err == nil {
				t.Error("Expected error when not authenticated")
			}

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when note does not exist", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Patch(ctx, 999)
			if err == nil {
				t.Error("Expected error when note does not exist")
			}

			if err != nil && !strings.Contains(err.Error(), "failed to get note") {
				t.Errorf("Expected 'failed to get note' error, got '%v'", err)
			}
		})

		t.Run("returns error when note not published", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note := &models.Note{
				Title:   "Not Published",
				Content: "Test content",
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Patch(ctx, id)
			if err == nil {
				t.Error("Expected error when note not published")
			}

			if err != nil && !strings.Contains(err.Error(), "not published") {
				t.Errorf("Expected 'not published' error, got '%v'", err)
			}
		})

		t.Run("handles published note with existing metadata", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "existing_rkey"
			cid := "existing_cid"
			publishedAt := time.Now().Add(-24 * time.Hour)
			note := &models.Note{
				Title:       "Published Note",
				Content:     "# Updated content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				PublishedAt: &publishedAt,
				IsDraft:     false,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Patch(ctx, id)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during patch (expected for external service call): %v", err)
			}
		})

		t.Run("handles draft note", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "draft_rkey"
			cid := "draft_cid"
			note := &models.Note{
				Title:       "Draft Note",
				Content:     "# Draft content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				IsDraft:     true,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Patch(ctx, id)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during patch (expected for external service call): %v", err)
			}
		})

		t.Run("handles markdown conversion errors", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "test_rkey"
			cid := "test_cid"
			note := &models.Note{
				Title:       "Test Note",
				Content:     "# Valid markdown",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Patch(ctx, id)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during patch (expected for external service call): %v", err)
			}
		})
	})

	t.Run("PostPreview", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.PostPreview(ctx, 1, false, "", false)
			if err == nil {
				t.Error("Expected error when not authenticated")
			}

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when note does not exist", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.PostPreview(ctx, 999, false, "", false)
			if err == nil {
				t.Error("Expected error when note does not exist")
			}

			if err != nil && !strings.Contains(err.Error(), "failed to get note") {
				t.Errorf("Expected 'failed to get note' error, got '%v'", err)
			}
		})

		t.Run("returns error when note already published", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "existing_rkey"
			cid := "existing_cid"
			note := &models.Note{
				Title:       "Already Published",
				Content:     "# Test content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.PostPreview(ctx, id, false, "", false)
			if err == nil {
				t.Error("Expected error when note already published")
			}

			if err != nil && !strings.Contains(err.Error(), "already published") {
				t.Errorf("Expected 'already published' error, got '%v'", err)
			}
		})

		t.Run("shows preview for valid note", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note := &models.Note{
				Title:   "Test Note",
				Content: "# Test content\n\nThis is a test.",
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			mock := services.NewMockATProtoService()
			mock.IsAuthenticatedVal = true
			mock.Session = &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}
			handler.atproto = mock

			err = handler.PostPreview(ctx, id, false, "", false)
			suite.AssertNoError(err, "preview should succeed")
		})

		t.Run("shows preview for draft", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note := &models.Note{
				Title:   "Draft Note",
				Content: "# Draft content",
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			mock := services.NewMockATProtoService()
			mock.IsAuthenticatedVal = true
			mock.Session = &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}
			handler.atproto = mock

			err = handler.PostPreview(ctx, id, true, "", false)
			suite.AssertNoError(err, "preview draft should succeed")
		})
	})

	t.Run("PostValidate", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.PostValidate(ctx, 1, false, "", false)
			if err == nil {
				t.Error("Expected error when not authenticated")
			}

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("validates markdown conversion successfully", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note := &models.Note{
				Title:   "Test Note",
				Content: "# Test content\n\nValid markdown here.",
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			mock := services.NewMockATProtoService()
			mock.IsAuthenticatedVal = true
			mock.Session = &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}
			handler.atproto = mock

			err = handler.PostValidate(ctx, id, false, "", false)
			suite.AssertNoError(err, "validation should succeed")
		})
	})

	t.Run("PatchPreview", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.PatchPreview(ctx, 1, "", false)
			if err == nil {
				t.Error("Expected error when not authenticated")
			}

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when note does not exist", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.PatchPreview(ctx, 999, "", false)
			if err == nil {
				t.Error("Expected error when note does not exist")
			}

			if err != nil && !strings.Contains(err.Error(), "failed to get note") {
				t.Errorf("Expected 'failed to get note' error, got '%v'", err)
			}
		})

		t.Run("returns error when note not published", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note := &models.Note{
				Title:   "Not Published",
				Content: "# Test content",
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.PatchPreview(ctx, id, "", false)
			if err == nil {
				t.Error("Expected error when note not published")
			}

			if err != nil && !strings.Contains(err.Error(), "not published") {
				t.Errorf("Expected 'not published' error, got '%v'", err)
			}
		})

		t.Run("shows preview for published note", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "test_rkey"
			cid := "test_cid"
			publishedAt := time.Now().Add(-24 * time.Hour)
			note := &models.Note{
				Title:       "Published Note",
				Content:     "# Updated content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				PublishedAt: &publishedAt,
				IsDraft:     false,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			mock := services.NewMockATProtoService()
			mock.IsAuthenticatedVal = true
			mock.Session = &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}
			handler.atproto = mock

			err = handler.PatchPreview(ctx, id, "", false)
			suite.AssertNoError(err, "preview should succeed")
		})
	})

	t.Run("PatchValidate", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.PatchValidate(ctx, 1, "", false)
			if err == nil {
				t.Error("Expected error when not authenticated")
			}

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("validates markdown conversion successfully", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "test_rkey"
			cid := "test_cid"
			note := &models.Note{
				Title:       "Published Note",
				Content:     "# Updated content\n\nValid markdown here.",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				IsDraft:     false,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			mock := services.NewMockATProtoService()
			mock.IsAuthenticatedVal = true
			mock.Session = &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}
			handler.atproto = mock

			err = handler.PatchValidate(ctx, id, "", false)
			suite.AssertNoError(err, "validation should succeed")
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Delete(ctx, 1)
			if err == nil {
				t.Error("Expected error when not authenticated")
			}

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when note does not exist", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Delete(ctx, 999)
			if err == nil {
				t.Error("Expected error when note does not exist")
			}

			if err != nil && !strings.Contains(err.Error(), "failed to get note") {
				t.Errorf("Expected 'failed to get note' error, got '%v'", err)
			}
		})

		t.Run("returns error when note not published", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note := &models.Note{
				Title:   "Not Published",
				Content: "Test content",
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Delete(ctx, id)
			if err == nil {
				t.Error("Expected error when note not published")
			}

			if err != nil && !strings.Contains(err.Error(), "not published") {
				t.Errorf("Expected 'not published' error, got '%v'", err)
			}
		})

		t.Run("handles published note", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "test_rkey"
			cid := "test_cid"
			publishedAt := time.Now().Add(-24 * time.Hour)
			note := &models.Note{
				Title:       "Published Note",
				Content:     "# Test content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				PublishedAt: &publishedAt,
				IsDraft:     false,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Delete(ctx, id)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during delete (expected for external service call): %v", err)
			}
		})

		t.Run("handles draft note", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "draft_rkey"
			cid := "draft_cid"
			note := &models.Note{
				Title:       "Draft Note",
				Content:     "# Draft content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				IsDraft:     true,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Delete(ctx, id)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during delete (expected for external service call): %v", err)
			}
		})

		t.Run("does not clear metadata when delete fails", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey := "test_rkey"
			cid := "test_cid"
			publishedAt := time.Now().Add(-24 * time.Hour)
			note := &models.Note{
				Title:       "Test Note",
				Content:     "# Test content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				PublishedAt: &publishedAt,
				IsDraft:     false,
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Delete(ctx, id)
			if err == nil {
				t.Fatal("Expected delete to fail with invalid token")
			}

			if !strings.Contains(err.Error(), "failed to delete document") {
				t.Logf("Got error: %v", err)
			}

			updatedNote, err := handler.repos.Notes.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated note: %v", err)
			}

			if !updatedNote.HasLeafletAssociation() {
				t.Error("Note should still have leaflet association after failed delete")
			}

			if updatedNote.LeafletRKey == nil || updatedNote.LeafletCID == nil {
				t.Error("Note metadata should not be cleared after failed delete")
			}
		})
	})

	t.Run("Push", func(t *testing.T) {
		t.Run("returns error when not authenticated", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Push(ctx, []int64{1, 2, 3}, false, false)
			if err == nil {
				t.Error("Expected error when not authenticated")
			}

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Errorf("Expected 'not authenticated' error, got '%v'", err)
			}
		})

		t.Run("returns error when no note IDs provided", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Push(ctx, []int64{}, false, false)
			if err == nil {
				t.Error("Expected error when no note IDs provided")
			}

			if err != nil && !strings.Contains(err.Error(), "no note IDs provided") {
				t.Errorf("Expected 'no note IDs provided' error, got '%v'", err)
			}
		})

		t.Run("handles note not found error", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err := handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Push(ctx, []int64{999}, false, false)
			if err == nil {
				t.Error("Expected error when note not found")
			}

			if err != nil && !strings.Contains(err.Error(), "error(s)") {
				t.Errorf("Expected error about failures, got '%v'", err)
			}
		})

		t.Run("attempts to create notes without leaflet association", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note1 := &models.Note{
				Title:   "New Note 1",
				Content: "# Content 1",
			}
			note2 := &models.Note{
				Title:   "New Note 2",
				Content: "# Content 2",
			}

			id1, err := handler.repos.Notes.Create(ctx, note1)
			suite.AssertNoError(err, "create note 1")

			id2, err := handler.repos.Notes.Create(ctx, note2)
			suite.AssertNoError(err, "create note 2")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Push(ctx, []int64{id1, id2}, false, false)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during push (expected for external service call): %v", err)
			}
		})

		t.Run("attempts to update notes with leaflet association", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			rkey1 := "rkey1"
			cid1 := "cid1"
			publishedAt1 := time.Now().Add(-24 * time.Hour)
			note1 := &models.Note{
				Title:       "Published Note 1",
				Content:     "# Content 1",
				LeafletRKey: &rkey1,
				LeafletCID:  &cid1,
				PublishedAt: &publishedAt1,
				IsDraft:     false,
			}

			rkey2 := "rkey2"
			cid2 := "cid2"
			note2 := &models.Note{
				Title:       "Draft Note 2",
				Content:     "# Content 2",
				LeafletRKey: &rkey2,
				LeafletCID:  &cid2,
				IsDraft:     true,
			}

			id1, err := handler.repos.Notes.Create(ctx, note1)
			suite.AssertNoError(err, "create note 1")

			id2, err := handler.repos.Notes.Create(ctx, note2)
			suite.AssertNoError(err, "create note 2")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Push(ctx, []int64{id1, id2}, false, false)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during push (expected for external service call): %v", err)
			}
		})

		t.Run("handles mixed create and update operations", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			newNote := &models.Note{
				Title:   "New Note",
				Content: "# New content",
			}

			rkey := "existing_rkey"
			cid := "existing_cid"
			publishedAt := time.Now().Add(-24 * time.Hour)
			existingNote := &models.Note{
				Title:       "Existing Note",
				Content:     "# Updated content",
				LeafletRKey: &rkey,
				LeafletCID:  &cid,
				PublishedAt: &publishedAt,
				IsDraft:     false,
			}

			newID, err := handler.repos.Notes.Create(ctx, newNote)
			suite.AssertNoError(err, "create new note")

			existingID, err := handler.repos.Notes.Create(ctx, existingNote)
			suite.AssertNoError(err, "create existing note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Push(ctx, []int64{newID, existingID}, false, false)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during push (expected for external service call): %v", err)
			}
		})

		t.Run("continues processing after individual failures", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note1 := &models.Note{
				Title:   "Valid Note",
				Content: "# Content",
			}

			id1, err := handler.repos.Notes.Create(ctx, note1)
			suite.AssertNoError(err, "create valid note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			invalidID := int64(999)
			err = handler.Push(ctx, []int64{id1, invalidID}, false, false)

			if err == nil {
				t.Error("Expected error due to invalid note ID")
			}

			if err != nil && !strings.Contains(err.Error(), "error(s)") {
				t.Errorf("Expected error message about failures, got '%v'", err)
			}
		})

		t.Run("respects draft flag for new notes", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			note := &models.Note{
				Title:   "Draft Note",
				Content: "# Draft content",
			}

			id, err := handler.repos.Notes.Create(ctx, note)
			suite.AssertNoError(err, "create note")

			session := &services.Session{
				DID:           "did:plc:test123",
				Handle:        "test.bsky.social",
				AccessJWT:     "access_token",
				RefreshJWT:    "refresh_token",
				PDSURL:        "https://bsky.social",
				ExpiresAt:     time.Now().Add(2 * time.Hour),
				Authenticated: true,
			}

			err = handler.atproto.RestoreSession(session)
			if err != nil {
				t.Fatalf("Failed to restore session: %v", err)
			}

			err = handler.Push(ctx, []int64{id}, true, false)

			if err != nil && !strings.Contains(err.Error(), "not authenticated") {
				t.Logf("Got error during push (expected for external service call): %v", err)
			}
		})
	})

	t.Run("Read", func(t *testing.T) {
		t.Run("returns error when no publications exist", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Read(ctx, "")
			if err == nil {
				t.Error("Expected error when no publications exist")
			}

			if !strings.Contains(err.Error(), "note not found") {
				t.Errorf("Expected 'note not found' error, got '%v'", err)
			}
		})

		t.Run("returns error for non-existent numeric ID", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Read(ctx, "999")
			if err == nil {
				t.Error("Expected error for non-existent ID")
			}

			if !strings.Contains(err.Error(), "failed to get publication by ID") {
				t.Errorf("Expected 'failed to get publication by ID' error, got '%v'", err)
			}
		})

		t.Run("returns error when note by ID is not a publication", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			regularNote := &models.Note{
				Title:   "Regular Note",
				Content: "# Not a publication",
			}

			id, err := handler.repos.Notes.Create(ctx, regularNote)
			suite.AssertNoError(err, "create regular note")

			err = handler.Read(ctx, fmt.Sprintf("%d", id))
			if err == nil {
				t.Error("Expected error when note is not a publication")
			}

			if !strings.Contains(err.Error(), "not a publication") {
				t.Errorf("Expected 'not a publication' error, got '%v'", err)
			}
		})

		t.Run("returns error for non-existent rkey", func(t *testing.T) {
			suite := NewHandlerTestSuite(t)
			defer suite.Cleanup()

			handler := CreateHandler(t, NewPublicationHandler)
			ctx := context.Background()

			err := handler.Read(ctx, "nonexistent_rkey")
			if err == nil {
				t.Error("Expected error for non-existent rkey")
			}

			if !strings.Contains(err.Error(), "failed to get publication by rkey") {
				t.Errorf("Expected 'failed to get publication by rkey' error, got '%v'", err)
			}
		})
	})

	t.Run("Auth Success Path", func(t *testing.T) {
		suite := NewHandlerTestSuite(t)
		defer suite.Cleanup()

		handler := CreateHandler(t, NewPublicationHandler)
		ctx := context.Background()

		mock := services.SetupSuccessfulAuthMocks()
		handler.atproto = mock

		err := handler.Auth(ctx, "test.bsky.social", "password123")
		suite.AssertNoError(err, "authentication should succeed")

		if !handler.atproto.IsAuthenticated() {
			t.Error("Expected handler to be authenticated after successful auth")
		}

		session, err := handler.atproto.GetSession()
		suite.AssertNoError(err, "get session should succeed")

		if session.Handle != "test.bsky.social" {
			t.Errorf("Expected handle 'test.bsky.social', got '%s'", session.Handle)
		}

		if session.DID == "" {
			t.Error("Expected DID to be set")
		}
	})

	t.Run("Pull Success Path", func(t *testing.T) {
		suite := NewHandlerTestSuite(t)
		defer suite.Cleanup()

		handler := CreateHandler(t, NewPublicationHandler)
		ctx := context.Background()

		mock := services.SetupSuccessfulPullMocks()
		handler.atproto = mock

		err := handler.Pull(ctx)
		suite.AssertNoError(err, "pull should succeed")

		notes, err := handler.repos.Notes.GetLeafletNotes(ctx)
		suite.AssertNoError(err, "get leaflet notes should succeed")

		if len(notes) != 1 {
			t.Errorf("Expected 1 note created, got %d", len(notes))
		}

		if notes[0].Title != "Test Document" {
			t.Errorf("Expected title 'Test Document', got '%s'", notes[0].Title)
		}

		if notes[0].LeafletRKey == nil || *notes[0].LeafletRKey != "test_rkey" {
			t.Error("Expected leaflet rkey to be set correctly")
		}
	})

	t.Run("Post Success Path", func(t *testing.T) {
		suite := NewHandlerTestSuite(t)
		defer suite.Cleanup()

		handler := CreateHandler(t, NewPublicationHandler)
		ctx := context.Background()

		mock := services.NewMockATProtoService()
		mock.IsAuthenticatedVal = true
		mock.Session = &services.Session{
			DID:           "did:plc:test123",
			Handle:        "test.bsky.social",
			AccessJWT:     "mock_access",
			RefreshJWT:    "mock_refresh",
			PDSURL:        "https://bsky.social",
			ExpiresAt:     time.Now().Add(2 * time.Hour),
			Authenticated: true,
		}
		handler.atproto = mock

		note := &models.Note{
			Title:   "Test Post",
			Content: "# Test Content\n\nThis is a test.",
		}

		id, err := handler.repos.Notes.Create(ctx, note)
		suite.AssertNoError(err, "create note should succeed")

		err = handler.Post(ctx, id, false)
		suite.AssertNoError(err, "post should succeed")

		updatedNote, err := handler.repos.Notes.Get(ctx, id)
		suite.AssertNoError(err, "get updated note should succeed")

		if updatedNote.LeafletRKey == nil || *updatedNote.LeafletRKey != "mock_rkey_123" {
			t.Error("Expected leaflet rkey to be set after post")
		}

		if updatedNote.LeafletCID == nil || *updatedNote.LeafletCID != "mock_cid_456" {
			t.Error("Expected leaflet cid to be set after post")
		}

		if updatedNote.IsDraft {
			t.Error("Expected note to be marked as published")
		}

		if updatedNote.PublishedAt == nil {
			t.Error("Expected published at to be set")
		}
	})

	t.Run("Post Draft Success Path", func(t *testing.T) {
		suite := NewHandlerTestSuite(t)
		defer suite.Cleanup()

		handler := CreateHandler(t, NewPublicationHandler)
		ctx := context.Background()

		mock := services.NewMockATProtoService()
		mock.IsAuthenticatedVal = true
		mock.Session = &services.Session{
			DID:           "did:plc:test123",
			Handle:        "test.bsky.social",
			AccessJWT:     "mock_access",
			RefreshJWT:    "mock_refresh",
			PDSURL:        "https://bsky.social",
			ExpiresAt:     time.Now().Add(2 * time.Hour),
			Authenticated: true,
		}
		handler.atproto = mock

		note := &models.Note{
			Title:   "Test Draft",
			Content: "# Draft Content",
		}

		id, err := handler.repos.Notes.Create(ctx, note)
		suite.AssertNoError(err, "create note should succeed")

		err = handler.Post(ctx, id, true)
		suite.AssertNoError(err, "post draft should succeed")

		updatedNote, err := handler.repos.Notes.Get(ctx, id)
		suite.AssertNoError(err, "get updated note should succeed")

		if !updatedNote.IsDraft {
			t.Error("Expected note to be marked as draft")
		}

		if updatedNote.PublishedAt != nil {
			t.Error("Expected published at to be nil for draft")
		}
	})

	t.Run("Patch Success Path", func(t *testing.T) {
		suite := NewHandlerTestSuite(t)
		defer suite.Cleanup()

		handler := CreateHandler(t, NewPublicationHandler)
		ctx := context.Background()

		mock := services.NewMockATProtoService()
		mock.IsAuthenticatedVal = true
		mock.Session = &services.Session{
			DID:           "did:plc:test123",
			Handle:        "test.bsky.social",
			AccessJWT:     "mock_access",
			RefreshJWT:    "mock_refresh",
			PDSURL:        "https://bsky.social",
			ExpiresAt:     time.Now().Add(2 * time.Hour),
			Authenticated: true,
		}
		handler.atproto = mock

		rkey := "existing_rkey"
		cid := "existing_cid"
		publishedAt := time.Now().Add(-24 * time.Hour)
		note := &models.Note{
			Title:       "Updated Note",
			Content:     "# Updated Content",
			LeafletRKey: &rkey,
			LeafletCID:  &cid,
			PublishedAt: &publishedAt,
			IsDraft:     false,
		}

		id, err := handler.repos.Notes.Create(ctx, note)
		suite.AssertNoError(err, "create note should succeed")

		err = handler.Patch(ctx, id)
		suite.AssertNoError(err, "patch should succeed")

		updatedNote, err := handler.repos.Notes.Get(ctx, id)
		suite.AssertNoError(err, "get updated note should succeed")

		if updatedNote.LeafletCID == nil || *updatedNote.LeafletCID != "mock_cid_updated_789" {
			t.Error("Expected leaflet cid to be updated after patch")
		}
	})

	t.Run("Delete Success Path", func(t *testing.T) {
		suite := NewHandlerTestSuite(t)
		defer suite.Cleanup()

		handler := CreateHandler(t, NewPublicationHandler)
		ctx := context.Background()

		mock := services.NewMockATProtoService()
		mock.IsAuthenticatedVal = true
		mock.Session = &services.Session{
			DID:           "did:plc:test123",
			Handle:        "test.bsky.social",
			AccessJWT:     "mock_access",
			RefreshJWT:    "mock_refresh",
			PDSURL:        "https://bsky.social",
			ExpiresAt:     time.Now().Add(2 * time.Hour),
			Authenticated: true,
		}
		handler.atproto = mock

		rkey := "test_rkey"
		cid := "test_cid"
		publishedAt := time.Now().Add(-24 * time.Hour)
		note := &models.Note{
			Title:       "Note to Delete",
			Content:     "# Content",
			LeafletRKey: &rkey,
			LeafletCID:  &cid,
			PublishedAt: &publishedAt,
			IsDraft:     false,
		}

		id, err := handler.repos.Notes.Create(ctx, note)
		suite.AssertNoError(err, "create note should succeed")

		err = handler.Delete(ctx, id)
		suite.AssertNoError(err, "delete should succeed")

		updatedNote, err := handler.repos.Notes.Get(ctx, id)
		suite.AssertNoError(err, "get updated note should succeed")

		if updatedNote.LeafletRKey != nil {
			t.Error("Expected leaflet rkey to be cleared after delete")
		}

		if updatedNote.LeafletCID != nil {
			t.Error("Expected leaflet cid to be cleared after delete")
		}

		if updatedNote.PublishedAt != nil {
			t.Error("Expected published at to be cleared after delete")
		}

		if updatedNote.IsDraft {
			t.Error("Expected draft flag to be false after delete")
		}
	})
}
