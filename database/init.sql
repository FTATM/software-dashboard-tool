-- 1. EXTENSIONS
-- Required inside the target database even if the binary is baked into the Docker image
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 2. CUSTOM TYPES
DO $$ BEGIN
    CREATE TYPE device_protocol_type AS ENUM ('MQTT', 'TCP', 'UDP', 'HTTP');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE schedule_type AS ENUM ('one_time', 'recurring');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE schedule_status AS ENUM ('active', 'completed', 'cancelled');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

-- 3. RBAC & CORE IDENTITY
CREATE TABLE IF NOT EXISTS role (
    role_id SERIAL,
    role_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pk_role_id PRIMARY KEY (role_id),
    CONSTRAINT uq_role_name UNIQUE (role_name)
);

CREATE TABLE IF NOT EXISTS "user" (
    user_id SERIAL,
    first_name TEXT,
    last_name TEXT,
    username TEXT,
    password_hash TEXT, 
    created_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    active BOOL DEFAULT FALSE,
    role_id INT,
    email TEXT,
    tel TEXT,
    line_user_token TEXT NULL,
    CONSTRAINT pk_user PRIMARY KEY (user_id),
    CONSTRAINT fk_user_role FOREIGN KEY (role_id) REFERENCES role(role_id) ON DELETE SET NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_active_username
ON "user" (username)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_line_user_token_active
ON "user" (line_user_token)
WHERE deleted_at IS NULL AND line_user_token IS NOT NULL;

CREATE TABLE IF NOT EXISTS user_notification (
    user_id INT,
    email_active BOOLEAN NOT NULL DEFAULT FALSE,
    sms_active BOOLEAN NOT NULL DEFAULT FALSE,
    line_active BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pk_user_notif PRIMARY KEY (user_id),
    CONSTRAINT fk_user_notifi_user FOREIGN KEY (user_id) REFERENCES "user" (user_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS ix_user_notif_email ON user_notification (email_active) WHERE email_active = TRUE;
CREATE INDEX IF NOT EXISTS ix_user_notif_sms ON user_notification (sms_active) WHERE sms_active = TRUE;
CREATE INDEX IF NOT EXISTS ix_user_notif_line ON user_notification (line_active) WHERE line_active = TRUE;

CREATE TABLE IF NOT EXISTS menu (
    menu_id SERIAL,
    menu_name TEXT NOT NULL,
    parent_id INT,
    CONSTRAINT pk_menu_id PRIMARY KEY (menu_id),
    CONSTRAINT uq_menu_name UNIQUE (menu_name),
    CONSTRAINT fk_menu_parent FOREIGN KEY (parent_id) REFERENCES menu(menu_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS action (
    action_id SERIAL,
    action_name TEXT NOT NULL,
    CONSTRAINT pk_action_id PRIMARY KEY (action_id),
    CONSTRAINT uq_action_name UNIQUE (action_name)
);

CREATE TABLE IF NOT EXISTS menu_action (
    menu_id INT NOT NULL,
    action_id INT NOT NULL,
    CONSTRAINT pk_menu_action PRIMARY KEY (menu_id, action_id),
    CONSTRAINT fk_ma_menu FOREIGN KEY (menu_id) REFERENCES menu(menu_id) ON DELETE CASCADE,
    CONSTRAINT fk_ma_action FOREIGN KEY (action_id) REFERENCES action(action_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS role_permission (
    role_id INT NOT NULL,
    menu_id INT NOT NULL,
    action_id INT NOT NULL,
    CONSTRAINT pk_role_permission PRIMARY KEY (role_id, menu_id, action_id),
    CONSTRAINT fk_rp_role FOREIGN KEY (role_id) REFERENCES role(role_id) ON DELETE CASCADE,
    CONSTRAINT fk_rp_valid_action FOREIGN KEY (menu_id, action_id) 
        REFERENCES menu_action(menu_id, action_id) ON DELETE CASCADE
);

-- 4. CANVAS & DASHBOARDS
CREATE TABLE IF NOT EXISTS canvas (
    canvas_id SERIAL,
    canvas_name TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    canvas_style JSONB,
    CONSTRAINT pk_canvas PRIMARY KEY (canvas_id)
);

CREATE TABLE IF NOT EXISTS canvas_role (
    canvas_id INT NOT NULL,
    role_id INT NOT NULL,
    CONSTRAINT pk_canvas_role PRIMARY KEY (canvas_id, role_id),
    CONSTRAINT fk_cr_canvas FOREIGN KEY (canvas_id) REFERENCES canvas(canvas_id) ON DELETE CASCADE,
    CONSTRAINT fk_cr_role FOREIGN KEY (role_id) REFERENCES role(role_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS widget_type (
    widget_type_id SERIAL,
    widget_type_name TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT pk_widget_type PRIMARY KEY (widget_type_id),
    CONSTRAINT uq_widget_type UNIQUE (widget_type_name)
);

-- 5. IOT INFRASTRUCTURE & DEVICES
CREATE TABLE IF NOT EXISTS device_group (
    group_id SERIAL,
    group_name TEXT NOT NULL UNIQUE,
    protocol device_protocol_type,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pk_device_group PRIMARY KEY (group_id)
);

CREATE TABLE IF NOT EXISTS device (
    device_id SERIAL,
    device_name TEXT NOT NULL,
    protocol device_protocol_type,
    value_data INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    updated_value_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    active BOOL NOT NULL DEFAULT FALSE,
    last_seen_at TIMESTAMPTZ,
    last_alert_triggered_at TIMESTAMPTZ,

    -- EU & Virtual Sensor Reference
    ref_device_id INT,
    raw_min DOUBLE PRECISION,
    raw_max DOUBLE PRECISION,
    eu_min DOUBLE PRECISION,
    eu_max DOUBLE PRECISION,

    CONSTRAINT pk_device_id PRIMARY KEY (device_id),
    CONSTRAINT fk_device_ref FOREIGN KEY (ref_device_id) 
        REFERENCES device (device_id) ON DELETE SET NULL,
    CONSTRAINT chk_no_self_ref CHECK (ref_device_id IS NULL OR ref_device_id <> device_id)
) WITH (fillfactor = 70);

CREATE UNIQUE INDEX IF NOT EXISTS uq_device_name_active 
ON device (device_name) WHERE deleted_at IS NULL;

-- Index the reference column for fast lookup
CREATE INDEX IF NOT EXISTS idx_device_ref_id ON device (ref_device_id) 
WHERE ref_device_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS device_group_map (
    group_id INT NOT NULL,
    device_id INT NOT NULL,
    CONSTRAINT pk_device_group_map PRIMARY KEY (group_id, device_id),
    CONSTRAINT fk_device_map_group FOREIGN KEY (group_id) REFERENCES device_group(group_id) ON DELETE CASCADE,
    CONSTRAINT fk_device_map_device FOREIGN KEY (device_id) REFERENCES device(device_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS ix_device_group_map_device ON device_group_map(device_id);

CREATE TABLE IF NOT EXISTS widget (
    widget_id SERIAL,
    widget_type_id INT NOT NULL,
    canvas_id INT NOT NULL,
    device_group_id INT,
    device_id_s INTEGER[],
    widget_label TEXT,
    layout_data JSONB NOT NULL,
    widget_style JSONB,
    custom_chart_data JSONB,
    CONSTRAINT pk_widget PRIMARY KEY (widget_id),
    CONSTRAINT fk_widget_canvas FOREIGN KEY (canvas_id) REFERENCES canvas(canvas_id) ON DELETE CASCADE,
    CONSTRAINT fk_widget_type FOREIGN KEY (widget_type_id) REFERENCES widget_type(widget_type_id) ON DELETE RESTRICT,
    CONSTRAINT fk_widget_group FOREIGN KEY (device_group_id) REFERENCES device_group(group_id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS ix_canvas ON widget (canvas_id);
CREATE INDEX IF NOT EXISTS ix_device_gin ON widget USING GIN (device_id_s);

CREATE TABLE IF NOT EXISTS device_rule_notification (
    rule_id SERIAL,
    device_id INT NOT NULL,
    condition TEXT NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    reason TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pk_device_rule_notif PRIMARY KEY (rule_id),
    CONSTRAINT fk_device_rule_notif_device FOREIGN KEY (device_id) REFERENCES device(device_id) ON DELETE CASCADE
) WITH (fillfactor = 70);
CREATE INDEX IF NOT EXISTS ix_device_rule_notif_device ON device_rule_notification (device_id, active);

CREATE TABLE IF NOT EXISTS schedule (
    -- gen_random_uuid() works natively out of the box
    schedule_id UUID DEFAULT gen_random_uuid(),
    device_id INT,
    device_group_id INT,
    task_action JSONB NOT NULL, 
    schedule_type schedule_type NOT NULL,
    status schedule_status NOT NULL DEFAULT 'active',
    start_time TIMESTAMPTZ NOT NULL, 
    end_time TIMESTAMPTZ, 
    cron_expression TEXT, 
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pk_schedule_id PRIMARY KEY (schedule_id),
    CONSTRAINT fk_schedule_device FOREIGN KEY (device_id) REFERENCES device(device_id) ON DELETE CASCADE,
    CONSTRAINT fk_schedule_group FOREIGN KEY (device_group_id) REFERENCES device_group(group_id) ON DELETE CASCADE
);

-- 6. AUDIT LOG (HYPERTABLE)
CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL,
    entity_type TEXT NOT NULL,
    menu_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    action TEXT NOT NULL,
    changed_by INT,
    old_data JSONB,
    new_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pk_audit_log PRIMARY KEY (id, created_at)
);

SELECT create_hypertable('audit_log', by_range('created_at'), if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS ix_audit_log_entity ON audit_log (entity_type, entity_id, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_audit_log_menu ON audit_log (menu_type, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_audit_log_action ON audit_log (action, created_at DESC);

ALTER TABLE audit_log SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'menu_type, entity_type',
    timescaledb.compress_orderby = 'created_at DESC'
);

SELECT add_compression_policy('audit_log', INTERVAL '30 days', if_not_exists => TRUE);

-- 7. DEVICE TELEMETRY LOG (HIGH-VOLUME HYPERTABLE)
CREATE TABLE IF NOT EXISTS device_data_log (
    device_id INT NOT NULL,
    value_data INT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

SELECT create_hypertable('device_data_log', by_range('received_at', INTERVAL '1 day'), if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS ix_device_data_log_device_id ON device_data_log (device_id, received_at DESC);

ALTER TABLE device_data_log SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id',
    timescaledb.compress_orderby = 'received_at DESC'
);

SELECT add_compression_policy('device_data_log', INTERVAL '7 days', if_not_exists => TRUE);