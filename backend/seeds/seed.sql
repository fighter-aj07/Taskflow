-- Seed data for TaskFlow
-- User password: password123 (bcrypt cost 12)
-- This seed is meant to be run idempotently: only inserts if no users exist.

DO $$
BEGIN
    IF (SELECT COUNT(*) FROM users) = 0 THEN
        INSERT INTO users (id, name, email, password) VALUES
            ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Test User', 'test@example.com',
             '$2a$12$6nuR9bIwsesEOdj2BIfc9uBuHdOTOkn1ME0kyYYVjgsk3y1k2U8mC');

        INSERT INTO projects (id, name, description, owner_id) VALUES
            ('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Website Redesign', 'Q2 redesign project',
             'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11');

        INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, created_by, due_date) VALUES
            ('Design homepage', 'Create new homepage mockups', 'todo', 'high',
             'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
             'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2026-04-20'),
            ('Set up CI pipeline', 'Configure GitHub Actions', 'in_progress', 'medium',
             'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
             'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2026-04-15'),
            ('Write API docs', 'Document all endpoints', 'done', 'low',
             'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', NULL,
             'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NULL);
    END IF;
END $$;
