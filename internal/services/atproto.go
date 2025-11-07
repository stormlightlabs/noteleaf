// Package services provides AT Protocol integration for leaflet.pub
//
// Document Flow:
//   - Pull: Fetch pub.leaflet.document records from AT Protocol repository
//   - Post: Create new pub.leaflet.document records in AT Protocol repository
//   - Push: Update existing pub.leaflet.document records in AT Protocol repository
//   - Delete: Remove pub.leaflet.document records from AT Protocol repository
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/stormlightlabs/noteleaf/internal/public"
)

// DocumentWithMeta combines a document with its repository metadata
type DocumentWithMeta struct {
	Document public.Document
	Meta     public.DocumentMeta
}

// convertCBORToJSONCompatible recursively converts CBOR data structures to JSON-compatible types
//
// This converts map[any]any to map[string]any to allow usage of [json.Marshal]
func convertCBORToJSONCompatible(data any) any {
	switch v := data.(type) {
	case map[any]any:
		result := make(map[string]any, len(v))
		for key, value := range v {
			strKey := fmt.Sprintf("%v", key)
			result[strKey] = convertCBORToJSONCompatible(value)
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, value := range v {
			result[key] = convertCBORToJSONCompatible(value)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = convertCBORToJSONCompatible(item)
		}
		return result
	default:
		return v
	}
}

// convertJSONToCBORCompatible recursively converts JSON-compatible data structures to CBOR types
//
// This converts map[string]any to map[any]any to allow proper CBOR encoding for AT Protocol
func convertJSONToCBORCompatible(data any) any {
	switch v := data.(type) {
	case map[string]any:
		result := make(map[any]any, len(v))
		for key, value := range v {
			result[key] = convertJSONToCBORCompatible(value)
		}
		return result
	case map[any]any:
		result := make(map[any]any, len(v))
		for key, value := range v {
			result[key] = convertJSONToCBORCompatible(value)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = convertJSONToCBORCompatible(item)
		}
		return result
	default:
		return v
	}
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

// ATProtoClient defines the interface for AT Protocol operations
type ATProtoClient interface {
	Authenticate(ctx context.Context, handle, password string) error
	GetSession() (*Session, error)
	IsAuthenticated() bool
	RestoreSession(session *Session) error
	PullDocuments(ctx context.Context) ([]DocumentWithMeta, error)
	PostDocument(ctx context.Context, doc public.Document, isDraft bool) (*DocumentWithMeta, error)
	PatchDocument(ctx context.Context, rkey string, doc public.Document, isDraft bool) (*DocumentWithMeta, error)
	DeleteDocument(ctx context.Context, rkey string, isDraft bool) error
	UploadBlob(ctx context.Context, data []byte, mimeType string) (public.Blob, error)
	GetDefaultPublication(ctx context.Context) (string, error)
	Close() error
}

// ATProtoService provides AT Protocol operations for leaflet integration
type ATProtoService struct {
	handle   string
	password string
	session  *Session
	pdsURL   string // Personal Data Server URL
	client   *xrpc.Client

	// TODO: Future enhancement - integrate OS keychain for secure password storage
	// Consider using keyring libraries like:
	//   - github.com/zalando/go-keyring (cross-platform)
	//   - keychain access on macOS (Security.framework)
	//   - Windows Credential Manager (credman)
	//   - Linux Secret Service API (libsecret)
	// This would allow storing app passwords securely in the system keychain
	// instead of requiring re-authentication every time JWTs expire.
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
// and automatically refreshes the token if expired
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

	// Check if token is expired or about to expire (within 5 minutes)
	if time.Now().Add(5 * time.Minute).After(session.ExpiresAt) {
		ctx := context.Background()
		if err := s.RefreshToken(ctx); err != nil {
			// Token refresh failed - session may be invalid
			// User will need to re-authenticate
			return fmt.Errorf("session expired and refresh failed: %w", err)
		}
	}

	return nil
}

// RefreshToken refreshes the access token using the refresh token
// This extends the session without requiring the user to re-authenticate
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

	// TODO: Consider increasing token lifetime for better UX
	// Current: 2 hours - requires frequent re-authentication
	// Consider: Store in OS keychain to enable longer sessions without security risk
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

	documentCount := 0
	err = r.ForEach(ctx, prefix, func(k string, v cid.Cid) error {
		documentCount++

		_, recordBytes, err := r.GetRecordBytes(ctx, k)
		if err != nil {
			return fmt.Errorf("failed to get record bytes for %s: %w", k, err)
		}

		var cborData any
		if err := cbor.Unmarshal(*recordBytes, &cborData); err != nil {
			return fmt.Errorf("failed to decode CBOR for document %s: %w", k, err)
		}

		jsonCompatible := convertCBORToJSONCompatible(cborData)

		jsonBytes, err := json.MarshalIndent(jsonCompatible, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to convert CBOR to JSON for document %s: %w", k, err)
		}

		parts := strings.Split(k, "/")
		rkey := ""
		if len(parts) > 0 {
			rkey = parts[len(parts)-1]
		}

		var typeCheck public.TypeCheck

		if err := json.Unmarshal(jsonBytes, &typeCheck); err != nil {
			return fmt.Errorf("failed to check $type for %s: %w", k, err)
		}

		if typeCheck.Type != public.TypeDocument {
			return nil
		}

		var doc public.Document
		if err := json.Unmarshal(jsonBytes, &doc); err != nil {
			return fmt.Errorf("failed to unmarshal JSON to Document for %s: %w", k, err)
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

		var cborData any
		if err := cbor.Unmarshal(*recordBytes, &cborData); err != nil {
			return fmt.Errorf("failed to decode CBOR for document %s: %w", k, err)
		}

		jsonCompatible := convertCBORToJSONCompatible(cborData)

		jsonBytes, err := json.MarshalIndent(jsonCompatible, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to convert CBOR to JSON for document %s: %w", k, err)
		}

		parts := strings.Split(k, "/")
		rkey := ""
		if len(parts) > 0 {
			rkey = parts[len(parts)-1]
		}

		var pub public.Publication
		if err := json.Unmarshal(jsonBytes, &pub); err != nil {
			return fmt.Errorf("failed to unmarshal publication %s: %w", k, err)
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

// GetDefaultPublication returns the URI of the first available publication for the authenticated user
//
// Returns an error if no publications exist
func (s *ATProtoService) GetDefaultPublication(ctx context.Context) (string, error) {
	publications, err := s.ListPublications(ctx)
	if err != nil {
		return "", err
	}

	if len(publications) == 0 {
		return "", fmt.Errorf("no publications found - create a publication on leaflet.pub first")
	}

	return publications[0].URI, nil
}

// PostDocument creates a new document in the user's repository
func (s *ATProtoService) PostDocument(ctx context.Context, doc public.Document, isDraft bool) (*DocumentWithMeta, error) {
	if !s.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	if doc.Title == "" {
		return nil, fmt.Errorf("document title is required")
	}

	collection := public.TypeDocument
	if isDraft {
		collection = public.TypeDocumentDraft
	}

	doc.Type = collection
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	var m map[string]any
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	m["$type"] = collection

	output, err := repoCreateRecord(ctx, s.client, s.session.DID, collection, m)
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	parts := strings.Split(output.Uri, "/")
	rkey := parts[len(parts)-1]

	meta := public.DocumentMeta{
		RKey:      rkey,
		CID:       output.Cid,
		URI:       output.Uri,
		IsDraft:   isDraft,
		FetchedAt: time.Now(),
	}

	return &DocumentWithMeta{
		Document: doc,
		Meta:     meta,
	}, nil
}

// PatchDocument updates an existing document in the user's repository
func (s *ATProtoService) PatchDocument(ctx context.Context, rkey string, doc public.Document, isDraft bool) (*DocumentWithMeta, error) {
	if !s.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	if rkey == "" {
		return nil, fmt.Errorf("rkey is required")
	}

	if doc.Title == "" {
		return nil, fmt.Errorf("document title is required")
	}

	collection := public.TypeDocument
	if isDraft {
		collection = public.TypeDocumentDraft
	}

	doc.Type = collection

	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document to JSON: %w", err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	cborCompatible := convertJSONToCBORCompatible(jsonData)

	cborBytes, err := cbor.Marshal(cborCompatible)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to CBOR: %w", err)
	}

	record := &lexutil.LexiconTypeDecoder{}
	if err := cbor.Unmarshal(cborBytes, record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CBOR to lexicon type: %w", err)
	}

	input := &atproto.RepoPutRecord_Input{
		Repo:       s.session.DID,
		Collection: collection,
		Rkey:       rkey,
		Record:     record,
	}

	output, err := atproto.RepoPutRecord(ctx, s.client, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	uri := fmt.Sprintf("at://%s/%s/%s", s.session.DID, collection, rkey)

	meta := public.DocumentMeta{
		RKey:      rkey,
		CID:       output.Cid,
		URI:       uri,
		IsDraft:   isDraft,
		FetchedAt: time.Now(),
	}

	return &DocumentWithMeta{
		Document: doc,
		Meta:     meta,
	}, nil
}

// DeleteDocument removes a document from the user's repository
func (s *ATProtoService) DeleteDocument(ctx context.Context, rkey string, isDraft bool) error {
	if !s.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	if rkey == "" {
		return fmt.Errorf("rkey is required")
	}

	collection := public.TypeDocument
	if isDraft {
		collection = public.TypeDocumentDraft
	}

	input := &atproto.RepoDeleteRecord_Input{
		Repo:       s.session.DID,
		Collection: collection,
		Rkey:       rkey,
	}

	_, err := atproto.RepoDeleteRecord(ctx, s.client, input)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

// UploadBlob uploads binary data as a blob to AT Protocol
func (s *ATProtoService) UploadBlob(ctx context.Context, data []byte, mimeType string) (public.Blob, error) {
	if !s.IsAuthenticated() {
		return public.Blob{}, fmt.Errorf("not authenticated")
	}

	if len(data) == 0 {
		return public.Blob{}, fmt.Errorf("data cannot be empty")
	}

	if mimeType == "" {
		return public.Blob{}, fmt.Errorf("mimeType is required")
	}

	output, err := atproto.RepoUploadBlob(ctx, s.client, bytes.NewReader(data))
	if err != nil {
		return public.Blob{}, fmt.Errorf("failed to upload blob: %w", err)
	}

	blob := public.Blob{
		Type:     public.TypeBlob,
		Ref:      public.CID{Link: output.Blob.Ref.String()},
		MimeType: output.Blob.MimeType,
		Size:     int(output.Blob.Size),
	}
	return blob, nil
}

// Close cleans up resources
func (s *ATProtoService) Close() error {
	s.session = nil
	return nil
}

type RepoCreateRecordOutput struct {
	Cid string `json:"cid"`
	Uri string `json:"uri"`
}

func repoCreateRecord(ctx context.Context, client *xrpc.Client, repo, collection string, record map[string]any) (*RepoCreateRecordOutput, error) {
	body := map[string]any{
		"repo":       repo,
		"collection": collection,
		"record":     record,
	}

	var out RepoCreateRecordOutput
	if err := client.LexDo(
		ctx,
		lexutil.Procedure,
		"application/json",
		"com.atproto.repo.createRecord",
		nil,
		body,
		&out,
	); err != nil {
		return nil, fmt.Errorf("repoCreateRecord failed: %w", err)
	}
	return &out, nil
}
