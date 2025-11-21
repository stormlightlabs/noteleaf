CREATE TABLE IF NOT EXISTS documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    doc_kind INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_documents_doc_kind ON documents(doc_kind);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at);
