CREATE TABLE users (
    iss TEXT NOT NULL,
    sub TEXT NOT NULL,
    username TEXT NOT NULL,
    display_name TEXT NOT NULL,
    email TEXT NOT NULL,
    last_logged_in TIMESTAMP,
    created_at TIMESTAMP,

    PRIMARY KEY (iss, sub)
);

CREATE TABLE user_groups (
    owner_iss TEXT NOT NULL,
    owner_sub TEXT NOT NULL,
    group_name TEXT NOT NULL,

    PRIMARY KEY (owner_iss, owner_sub, group_name),
    FOREIGN KEY (owner_iss, owner_sub) REFERENCES users(iss, sub) ON DELETE CASCADE
);

CREATE TABLE certificate_requests (
    id SERIAL PRIMARY KEY NOT NULL,
    owner_iss TEXT NOT NULL,
    owner_sub TEXT NOT NULL,
    message TEXT,

    -- Request details
    common_name TEXT NOT NULL,
    dns_names TEXT[] NOT NULL DEFAULT '{}',
    organizational_units TEXT[] NOT NULL DEFAULT '{}',
    validity_days INTEGER NOT NULL DEFAULT 365,

    -- Status Tracking
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending, 'approved', 'rejected', 'issued'
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Certificate details (after issued)
    issued_at TIMESTAMP,
    expires_at TIMESTAMP,
    serial_number TEXT,
    certificate_pem TEXT,

    FOREIGN KEY (owner_iss, owner_sub) REFERENCES users(iss, sub) ON DELETE RESTRICT
);

CREATE TABLE certificate_events (
    id SERIAL PRIMARY KEY NOT NULL,
    certificate_request_id INTEGER NOT NULL,
    requester_iss TEXT NOT NULL,
    requester_sub TEXT NOT NULL,
    reviewer_iss TEXT NOT NULL,
    reviewer_sub TEXT NOT NULL,

    new_status TEXT NOT NULL, -- 'pending, 'approved', 'rejected', 'issued'
    review_notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (certificate_request_id) REFERENCES certificate_requests(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_iss, reviewer_sub) REFERENCES users(iss, sub) ON DELETE RESTRICT,
    FOREIGN KEY (requester_iss, requester_sub) REFERENCES users(iss, sub) ON DELETE RESTRICT
);

CREATE INDEX idx_cert_requests_owner ON certificate_requests(owner_iss, owner_sub);
CREATE INDEX idx_cert_events_status ON certificate_requests(status);
CREATE INDEX idx_cert_events_request_id ON certificate_events(certificate_request_id);
CREATE INDEX idx_cert_events_requester ON certificate_events(requester_iss, requester_sub);
CREATE INDEX idx_cert_events_reviewer ON certificate_events(reviewer_iss, reviewer_sub);


