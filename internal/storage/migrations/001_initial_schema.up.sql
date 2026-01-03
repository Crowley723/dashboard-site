CREATE TABLE users(
    iss TEXT NOT NULL,
    sub TEXT NOT NULL,
    username TEXT NOT NULL,
    display_name TEXT NOT NULL,
    email TEXT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    last_logged_in TIMESTAMP,
    created_at TIMESTAMP,

    PRIMARY KEY (iss, sub)
);

CREATE UNIQUE INDEX idx_users_system ON users(is_system) WHERE is_system = TRUE;

CREATE TABLE user_groups(
    owner_iss TEXT NOT NULL,
    owner_sub TEXT NOT NULL,
    group_name TEXT NOT NULL,

    PRIMARY KEY (owner_iss, owner_sub, group_name),
    FOREIGN KEY (owner_iss, owner_sub) REFERENCES users(iss, sub) ON DELETE CASCADE
);

CREATE TABLE certificate_requests(
    id SERIAL PRIMARY KEY NOT NULL,
    owner_iss TEXT NOT NULL,
    owner_sub TEXT NOT NULL,
    message TEXT,

    -- Request details
    common_name TEXT NOT NULL,
    dns_names TEXT[] DEFAULT '{}',
    organizational_units TEXT[] DEFAULT '{}',
    validity_days INTEGER NOT NULL DEFAULT 365,

    -- Status Tracking
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending, 'approved', 'rejected', 'issued'
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Kubernetes Certificate metadata
    k8s_certificate_name TEXT,
    k8s_namespace TEXT,
    k8s_secret_name TEXT,

    -- Certificate details (after issued)
    issued_at TIMESTAMP,
    expires_at TIMESTAMP,
    serial_number TEXT,
    certificate_pem TEXT,

    FOREIGN KEY (owner_iss, owner_sub) REFERENCES users(iss, sub) ON DELETE RESTRICT
);

CREATE TABLE certificate_downloads(
    id SERIAL PRIMARY KEY NOT NULL,
    certificate_request_id INTEGER NOT NULL,
    downloader_sub TEXT NOT NULL,
    downloader_iss TEXT NOT NULL,

    ip_address INET NOT NULL,

    user_agent TEXT,
    browser_name TEXT,
    browser_version TEXT,
    os_name TEXT,
    os_version TEXT,
    device_type TEXT,

    downloaded_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (downloader_iss, downloader_sub) REFERENCES users(iss, sub) ON DELETE RESTRICT,
    FOREIGN KEY (certificate_request_id) REFERENCES certificate_requests(id) ON DELETE RESTRICT
);
CREATE INDEX idx_cert_downloads_owner ON certificate_downloads(downloader_iss, downloader_sub);
CREATE INDEX idx_cert_downloads_ip ON certificate_downloads(ip_address);
CREATE INDEX idx_cert_downloads_cert_id ON certificate_downloads(certificate_request_id);


CREATE TABLE certificate_events(
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
CREATE INDEX idx_cert_requests_status ON certificate_requests(status);
CREATE INDEX idx_cert_requests_k8s_name ON certificate_requests(k8s_certificate_name, k8s_namespace);
CREATE INDEX idx_cert_events_request_id ON certificate_events(certificate_request_id);
CREATE INDEX idx_cert_events_requester ON certificate_events(requester_iss, requester_sub);
CREATE INDEX idx_cert_events_reviewer ON certificate_events(reviewer_iss, reviewer_sub);


