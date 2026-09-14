-- 1. Widget Types
INSERT INTO widget_type (widget_type_name) 
VALUES 
    ('BarChart'), ('BarLineChart'), ('BulletChart'), ('AnalogueGauge'), 
    ('DigitalGauge'), ('LineChart'), ('PieChart'), ('ScatterChart'), 
    ('BarProcess'), ('Status'), ('Table'), ('Alert'), 
    ('Text'), ('RowChart'), ('ScoreCard'), ('InputAction'), ('Visualization')
ON CONFLICT (widget_type_name) DO NOTHING;

-- 2. Menus
INSERT INTO menu (menu_name) 
VALUES 
    ('Dashboard'), ('Canvas'), ('Canvas Design'), ('Canvas Access'), 
    ('Scheduler'), ('Log Report'), ('Notification User'), 
    ('Notification Device'), ('User'), ('Role'), ('Device'), ('Device Group')
ON CONFLICT (menu_name) DO NOTHING;

-- 3. Actions
INSERT INTO action (action_name) 
VALUES 
    ('Display'), ('Create'), ('Update'), ('Delete'), 
    ('Import'), ('Export'), ('Query'), ('Command')
ON CONFLICT (action_name) DO NOTHING;

-- 4. Map Actions to Menus (Consolidated Single Statement)
WITH menu_action_mapping (menu_name, action_names) AS (
    VALUES
        ('Dashboard',           ARRAY['Display', 'Query']),
        ('Device',              ARRAY['Display', 'Create', 'Update', 'Delete', 'Import', 'Export', 'Command']),
        ('Device Group',        ARRAY['Display', 'Create', 'Update', 'Delete']),
        ('User',                ARRAY['Display', 'Create', 'Update', 'Delete']),
        ('Role',                ARRAY['Display', 'Create', 'Update', 'Delete']),
        ('Canvas',              ARRAY['Display', 'Create', 'Update', 'Delete']),
        ('Canvas Design',       ARRAY['Display', 'Update']),
        ('Canvas Access',       ARRAY['Display', 'Update']),
        ('Scheduler',           ARRAY['Display', 'Create', 'Update', 'Delete']),
        ('Log Report',          ARRAY['Display', 'Export']),
        ('Notification User',   ARRAY['Display', 'Update']),
        ('Notification Device', ARRAY['Display', 'Create', 'Update', 'Delete'])
)
INSERT INTO menu_action (menu_id, action_id)
SELECT m.menu_id, a.action_id
FROM menu_action_mapping map
JOIN menu m ON m.menu_name = map.menu_name
JOIN action a ON a.action_name = ANY(map.action_names)
ON CONFLICT (menu_id, action_id) DO NOTHING;

-- 5. Roles
INSERT INTO role (role_name) 
VALUES ('Admin') 
ON CONFLICT (role_name) DO NOTHING;

-- 6. Grant All Configured Menu Actions to Admin
INSERT INTO role_permission (role_id, menu_id, action_id)
SELECT r.role_id, ma.menu_id, ma.action_id
FROM role r
CROSS JOIN menu_action ma
WHERE r.role_name = 'Admin'
ON CONFLICT (role_id, menu_id, action_id) DO NOTHING;