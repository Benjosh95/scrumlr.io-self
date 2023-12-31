CREATE TABLE passkey_sessions
(
    "user"               uuid        NOT NULL REFERENCES users ON DELETE CASCADE,
    "challenge"          varchar(64) NOT NULL,
    "user_verification"  varchar(32) NOT NULL,
    "expires"            timestamptz NOT NULL,
    "allowed_credential_i_ds" jsonb,
    "extensions"         jsonb,
    "created_at"         timestamptz NOT NULL DEFAULT NOW()
);