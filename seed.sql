DELETE FROM showtimes;
DELETE FROM movies;
DELETE FROM screening_rooms;
DELETE FROM cinemas;

-- Cinemas
INSERT INTO cinemas (id, name, location, city, total_seats) VALUES 
('497f6eca-6276-4993-bfeb-53cbbbba6f08', 'CGV Vincom Center', '72 Le Thanh Ton, Dist 1', 'HCMC', 200),
('497f6eca-6276-4993-bfeb-53cbbbba6f09', 'Galaxy Nguyen Du', '116 Nguyen Du, Dist 1', 'HCMC', 150);

-- Rooms
INSERT INTO screening_rooms (id, cinema_id, name, room_type, total_seats) VALUES
('497f6eca-6276-4993-bfeb-53cbbbba6f10', '497f6eca-6276-4993-bfeb-53cbbbba6f08', 'Room 1 (IMAX)', 'IMAX', 50),
('497f6eca-6276-4993-bfeb-53cbbbba6f11', '497f6eca-6276-4993-bfeb-53cbbbba6f08', 'Room 2 (2D)', '2D', 50);

-- Movies
INSERT INTO movies (id, title, title_vi, status, rating_label) VALUES
('497f6eca-6276-4993-bfeb-53cbbbba6f12', 'Mai', 'Mai', 'now_showing', 'C18'),
('497f6eca-6276-4993-bfeb-53cbbbba6f13', 'Meet Again My Sista', 'Gặp Lại Chị Bầu', 'now_showing', 'C13');

-- Showtimes
INSERT INTO showtimes (id, movie_id, cinema_id, room_id, show_date, show_time, price, base_price, status) VALUES
('497f6eca-6276-4993-bfeb-53cbbbba6f14', '497f6eca-6276-4993-bfeb-53cbbbba6f12', '497f6eca-6276-4993-bfeb-53cbbbba6f08', '497f6eca-6276-4993-bfeb-53cbbbba6f10', CURRENT_DATE, '19:00:00', 120000, 120000, 'open'),
('497f6eca-6276-4993-bfeb-53cbbbba6f15', '497f6eca-6276-4993-bfeb-53cbbbba6f13', '497f6eca-6276-4993-bfeb-53cbbbba6f08', '497f6eca-6276-4993-bfeb-53cbbbba6f11', CURRENT_DATE, '21:00:00', 90000, 90000, 'open');
