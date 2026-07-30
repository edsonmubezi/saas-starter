CREATE TABLE chat_threads (
    id              BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id),
    user_id         BIGINT NOT NULL REFERENCES users(id),
    title           VARCHAR(255) NOT NULL DEFAULT 'New Conversation',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delete_status   INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_chat_threads_org_user ON chat_threads(organization_id, user_id, delete_status);

CREATE TABLE chat_messages (
    id         BIGSERIAL PRIMARY KEY,
    thread_id  BIGINT NOT NULL REFERENCES chat_threads(id) ON DELETE CASCADE,
    role       VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_messages_thread ON chat_messages(thread_id, created_at ASC);
