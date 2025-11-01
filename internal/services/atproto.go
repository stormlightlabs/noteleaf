// Package services provides AT Protocol integration for leaflet.pub
//
// Document Flow:
//   - Pull: Fetch pub.leaflet.document records from AT Protocol repository
//   - Post: Create new pub.leaflet.document records in AT Protocol repository
//   - Push: Update existing pub.leaflet.document records in AT Protocol repository
//
// Publishing Workflow (TODO):
//  1. Post - Create new document:
//     - Convert note to pub.leaflet.document format
//     - Upload any embedded images as blobs
//     - Create record with com.atproto.repo.createRecord
//     - Store returned rkey and cid in note metadata
//  2. Push - Update existing document:
//     - Check if note has leaflet_rkey (indicates previously published)
//     - Convert updated note to pub.leaflet.document format
//     - Upload any new images as blobs
//     - Update record with com.atproto.repo.putRecord
//     - Update stored cid in note metadata
//  3. Delete - Remove published document:
//     - Use com.atproto.repo.deleteRecord with stored rkey
//     - Clear leaflet metadata from note
//
// Blob Upload (TODO):
//  1. Use com.atproto.repo.uploadBlob for images
//  2. Returns blob reference with CID
//  3. Include blob ref in ImageBlock structures
//
// Draft vs Published (TODO):
//  1. Draft documents stored in collection: pub.leaflet.document.draft
//  2. Published documents in: pub.leaflet.document
//  3. Moving from draft to published requires:
//     - Delete from draft collection
//     - Create in published collection
//     - Update note metadata with new rkey
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/ipfs/go-cid"
	"github.com/stormlightlabs/noteleaf/internal/public"
)

// DocumentWithMeta combines a document with its repository metadata
type DocumentWithMeta struct {
	Document public.Document
	Meta     public.DocumentMeta
}

// PublicationWithMeta combines a publication with its metadata
type PublicationWithMeta struct {
	Publication public.Publication
	RKey        string
	CID         string
	URI         string
}

// Session holds authentication session information
type Session struct {
	DID           string    // Decentralized Identifier
	Handle        string    // User handle (e.g., username.bsky.social)
	AccessJWT     string    // Access token
	RefreshJWT    string    // Refresh token
	PDSURL        string    // Personal Data Server URL
	ExpiresAt     time.Time // When access token expires
	Authenticated bool      // Whether session is valid
}

// ATProtoService provides AT Protocol operations for leaflet integration
type ATProtoService struct {
	handle   string
	password string
	session  *Session
	pdsURL   string // Personal Data Server URL
	client   *xrpc.Client
}

// NewATProtoService creates a new AT Protocol service
func NewATProtoService() *ATProtoService {
	pdsURL := "https://bsky.social"
	return &ATProtoService{
		pdsURL: pdsURL,
		client: &xrpc.Client{
			Host: pdsURL,
		},
	}
}

// Authenticate logs in with BlueSky/AT Protocol credentials
func (s *ATProtoService) Authenticate(ctx context.Context, handle, password string) error {
	if handle == "" || password == "" {
		return fmt.Errorf("handle and password are required")
	}

	s.handle = handle
	s.password = password

	input := &atproto.ServerCreateSession_Input{
		Identifier: handle,
		Password:   password,
	}

	output, err := atproto.ServerCreateSession(ctx, s.client, input)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	expiresAt := time.Now().Add(2 * time.Hour)

	s.session = &Session{
		DID:           output.Did,
		Handle:        output.Handle,
		AccessJWT:     output.AccessJwt,
		RefreshJWT:    output.RefreshJwt,
		PDSURL:        s.pdsURL,
		ExpiresAt:     expiresAt,
		Authenticated: true,
	}

	s.client.Auth = &xrpc.AuthInfo{
		AccessJwt:  output.AccessJwt,
		RefreshJwt: output.RefreshJwt,
		Handle:     output.Handle,
		Did:        output.Did,
	}

	return nil
}

// GetSession returns the current session information
func (s *ATProtoService) GetSession() (*Session, error) {
	if s.session == nil || !s.session.Authenticated {
		return nil, fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}
	return s.session, nil
}

// IsAuthenticated checks if the service has a valid session
func (s *ATProtoService) IsAuthenticated() bool {
	return s.session != nil && s.session.Authenticated
}

// RestoreSession restores a previously authenticated session from stored credentials
func (s *ATProtoService) RestoreSession(session *Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if session.DID == "" || session.AccessJWT == "" || session.RefreshJWT == "" {
		return fmt.Errorf("session missing required fields (DID, AccessJWT, RefreshJWT)")
	}

	s.session = session

	s.client.Auth = &xrpc.AuthInfo{
		AccessJwt:  session.AccessJWT,
		RefreshJwt: session.RefreshJWT,
		Handle:     session.Handle,
		Did:        session.DID,
	}

	if session.PDSURL != "" {
		s.pdsURL = session.PDSURL
		s.client.Host = session.PDSURL
	}

	return nil
}

// RefreshToken refreshes the access token using the refresh token
func (s *ATProtoService) RefreshToken(ctx context.Context) error {
	if s.session == nil || s.session.RefreshJWT == "" {
		return fmt.Errorf("no session available to refresh")
	}

	s.client.Auth = &xrpc.AuthInfo{
		AccessJwt:  s.session.AccessJWT,
		RefreshJwt: s.session.RefreshJWT,
		Handle:     s.session.Handle,
		Did:        s.session.DID,
	}

	output, err := atproto.ServerRefreshSession(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	expiresAt := time.Now().Add(2 * time.Hour)
	s.session.AccessJWT = output.AccessJwt
	s.session.RefreshJWT = output.RefreshJwt
	s.session.ExpiresAt = expiresAt
	s.session.Authenticated = true

	s.client.Auth.AccessJwt = output.AccessJwt
	s.client.Auth.RefreshJwt = output.RefreshJwt

	return nil
}

// PullDocuments fetches all leaflet documents from the user's repository
func (s *ATProtoService) PullDocuments(ctx context.Context) ([]DocumentWithMeta, error) {
	if !s.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	carBytes, err := atproto.SyncGetRepo(ctx, s.client, s.session.DID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	r, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(carBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse CAR file: %w", err)
	}

	var documents []DocumentWithMeta
	prefix := public.TypeDocument

	err = r.ForEach(ctx, prefix, func(k string, v cid.Cid) error {
		_, recordBytes, err := r.GetRecordBytes(ctx, k)
		if err != nil {
			return fmt.Errorf("failed to get record bytes for %s: %w", k, err)
		}

		var doc public.Document
		if err := json.Unmarshal(*recordBytes, &doc); err != nil {
			return fmt.Errorf("failed to unmarshal document %s: %w", k, err)
		}

		parts := strings.Split(k, "/")
		rkey := ""
		if len(parts) > 0 {
			rkey = parts[len(parts)-1]
		}

		uri := fmt.Sprintf("at://%s/%s", s.session.DID, k)

		meta := public.DocumentMeta{
			RKey:      rkey,
			CID:       v.String(),
			URI:       uri,
			IsDraft:   false,
			FetchedAt: time.Now(),
		}

		documents = append(documents, DocumentWithMeta{
			Document: doc,
			Meta:     meta,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate over documents: %w", err)
	}

	return documents, nil
}

// ListPublications fetches available publications for the authenticated user
func (s *ATProtoService) ListPublications(ctx context.Context) ([]PublicationWithMeta, error) {
	if !s.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	carBytes, err := atproto.SyncGetRepo(ctx, s.client, s.session.DID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	r, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(carBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse CAR file: %w", err)
	}

	var publications []PublicationWithMeta
	prefix := public.TypePublication

	err = r.ForEach(ctx, prefix, func(k string, v cid.Cid) error {
		_, recordBytes, err := r.GetRecordBytes(ctx, k)
		if err != nil {
			return fmt.Errorf("failed to get record bytes for %s: %w", k, err)
		}

		var pub public.Publication
		if err := json.Unmarshal(*recordBytes, &pub); err != nil {
			return fmt.Errorf("failed to unmarshal publication %s: %w", k, err)
		}

		parts := strings.Split(k, "/")
		rkey := ""
		if len(parts) > 0 {
			rkey = parts[len(parts)-1]
		}

		uri := fmt.Sprintf("at://%s/%s", s.session.DID, k)

		publications = append(publications, PublicationWithMeta{
			Publication: pub,
			RKey:        rkey,
			CID:         v.String(),
			URI:         uri,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate over publications: %w", err)
	}

	return publications, nil
}

// Close cleans up resources
func (s *ATProtoService) Close() error {
	s.session = nil
	return nil
}
