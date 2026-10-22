-- Migration: Add Maintenance & Operating Hours Notification Feature
-- Run this against an existing database if init.sql / preload.sql have already been executed.

-- 1. Create maintenance_notification table
CREATE TABLE IF NOT EXISTS maintenance_notification (
    maintenance_id SERIAL,
    device_id INT,
    device_group_id INT,
    start_date TIMESTAMPTZ NOT NULL DEFAULT now(),
    maintenance_date TIMESTAMPTZ NOT NULL,
    interval_days INT,
    day_count INT NOT NULL DEFAULT 0,
    reason TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    notified BOOLEAN NOT NULL DEFAULT FALSE,
    notif_sent_count INT NOT NULL DEFAULT 0,
    last_counted_at TIMESTAMPTZ,
    last_notified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pk_maintenance_notif PRIMARY KEY (maintenance_id),
    CONSTRAINT fk_maintenance_notif_device FOREIGN KEY (device_id) REFERENCES device(device_id) ON DELETE CASCADE,
    CONSTRAINT fk_maintenance_notif_group FOREIGN KEY (device_group_id) REFERENCES device_group(group_id) ON DELETE CASCADE,
    CONSTRAINT chk_maintenance_target CHECK (
        (device_id IS NOT NULL AND device_group_id IS NULL) OR
        (device_id IS NULL AND device_group_id IS NOT NULL)
    )
) WITH (fillfactor = 70);

ALTER TABLE maintenance_notification ADD COLUMN IF NOT EXISTS interval_days INT;
ALTER TABLE maintenance_notification ADD COLUMN IF NOT EXISTS notif_sent_count INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS ix_maintenance_notif_active ON maintenance_notification (active, maintenance_date);

-- 2. Register Notification Maintenance menu and RBAC permissions
INSERT INTO menu (menu_name)
VALUES ('Notification Maintenance')
ON CONFLICT (menu_name) DO NOTHING;

WITH menu_action_mapping (menu_name, action_names) AS (
    VALUES ('Notification Maintenance', ARRAY['Display', 'Create', 'Update', 'Delete'])
)
INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu_action_mapping map
JOIN menu m ON m.menu_name = map.menu_name
JOIN action a ON a.action_name = ANY(map.action_names)
ON CONFLICT (menu_id, action_id) DO NOTHING;

INSERT INTO role_permission (role_id, menu_id, action_id)
SELECT r.role_id, ma.menu_id, ma.action_id
FROM role r
CROSS JOIN menu_action ma
JOIN menu m ON ma.menu_id = m.menu_id
WHERE r.role_name = 'Admin' AND m.menu_name = 'Notification Maintenance'
ON CONFLICT (role_id, menu_id, action_id) DO NOTHING;
