CREATE TABLE user_group (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE approval_step (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL REFERENCES user_group(id) ON DELETE RESTRICT,
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'APPROVED', 'PROVISIONED', 'FAILED')),
    approved_by TEXT
);
