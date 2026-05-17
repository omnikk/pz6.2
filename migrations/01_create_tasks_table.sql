CREATE TABLE IF NOT EXISTS tasks (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT,
    due_date    DATE,
    done        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_title ON tasks(title);

-- демо-данные для проверки SQL-инъекции
INSERT INTO tasks (id, title, description) VALUES
    ('t_demo_1', 'Buy milk',     'in the morning'),
    ('t_demo_2', 'Write report', 'practice 5 report'),
    ('t_demo_3', 'Secret task',  'must NOT leak via SQL injection')
ON CONFLICT (id) DO NOTHING;