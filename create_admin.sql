-- Create admin user
-- Password: admin123 (hashed)
-- Run this after creating the database schema

INSERT INTO users (
    username,
    email,
    password_hash,
    full_name,
    phone,
    role,
    is_active,
    is_verified
) VALUES (
    'admin',
    'admin@cinema.local',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj6fYzYXxGK', -- bcrypt hash for 'admin123'
    'System Administrator',
    '0123456789',
    'admin',
    true,
    true
) ON CONFLICT (email) DO NOTHING;

-- Create some sample cinemas
INSERT INTO cinemas (name, location, city, total_seats) VALUES
('CGV Vincom Center', 'Vincom Center, 123 Đường ABC', 'Hồ Chí Minh', 200),
('CGV Aeon Mall', 'Aeon Mall, 456 Đường XYZ', 'Hà Nội', 150),
('Lotte Cinema Diamond', 'Diamond Plaza, 789 Đường DEF', 'Đà Nẵng', 180)
ON CONFLICT DO NOTHING;

-- Create sample movies
INSERT INTO movies (title, genre, duration, rating) VALUES
('Avengers: Endgame', 'Action, Sci-Fi', 181, 'PG-13'),
('The Lion King', 'Animation, Adventure', 118, 'G'),
('Joker', 'Crime, Drama', 122, 'R'),
('Frozen II', 'Animation, Adventure', 103, 'PG')
ON CONFLICT DO NOTHING;