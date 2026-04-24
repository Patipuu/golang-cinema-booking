-- ============================================================
-- Migration: Upgrade to Cinema Booking v1.0
-- Adds: Rooms, Promotions, Reviews, Audit Logs, etc.
-- ============================================================

-- 1. Screening Rooms (Room type: 2D, 3D, IMAX, 4DX)
CREATE TABLE IF NOT EXISTS screening_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cinema_id UUID NOT NULL REFERENCES cinemas(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    room_type VARCHAR(20) CHECK (room_type IN ('2D','3D','IMAX','4DX')),
    total_seats INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Update Seats to link to Screening Rooms
-- Moving seats from cinema-level to room-level
ALTER TABLE seats DROP CONSTRAINT IF EXISTS seats_cinema_id_fkey;
ALTER TABLE seats DROP COLUMN IF EXISTS cinema_id;
ALTER TABLE seats ADD COLUMN IF NOT EXISTS room_id UUID REFERENCES screening_rooms(id) ON DELETE CASCADE;
ALTER TABLE seats ADD COLUMN IF NOT EXISTS seat_type VARCHAR(20) CHECK (seat_type IN ('standard','vip','couple')) DEFAULT 'standard';
ALTER TABLE seats ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'maintenance'));

-- 3. Update Movies with PRD v1.0 fields
ALTER TABLE movies ADD COLUMN IF NOT EXISTS title_vi VARCHAR(300);
ALTER TABLE movies ADD COLUMN IF NOT EXISTS title_en VARCHAR(300);
ALTER TABLE movies ADD COLUMN IF NOT EXISTS director VARCHAR(200);
ALTER TABLE movies ADD COLUMN IF NOT EXISTS cast_members TEXT;
ALTER TABLE movies ADD COLUMN IF NOT EXISTS poster_url TEXT;
ALTER TABLE movies ADD COLUMN IF NOT EXISTS trailer_url TEXT;
ALTER TABLE movies ADD COLUMN IF NOT EXISTS rating_label VARCHAR(5) CHECK (rating_label IN ('P','C13','C16','C18'));
ALTER TABLE movies ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'coming_soon' CHECK (status IN ('coming_soon', 'now_showing', 'ended'));
ALTER TABLE movies ADD COLUMN IF NOT EXISTS avg_rating DECIMAL(3,2) DEFAULT 0;

-- 4. Update Showtimes to link to Rooms
ALTER TABLE showtimes ADD COLUMN IF NOT EXISTS room_id UUID REFERENCES screening_rooms(id) ON DELETE CASCADE;
ALTER TABLE showtimes ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'closed', 'cancelled'));
ALTER TABLE showtimes ADD COLUMN IF NOT EXISTS base_price DECIMAL(12,2);
ALTER TABLE showtimes ADD COLUMN IF NOT EXISTS price_modifier JSONB DEFAULT '{}';

-- 5. Promotions
CREATE TABLE IF NOT EXISTS promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    discount_type VARCHAR(10) CHECK (discount_type IN ('percent','fixed')),
    discount_value DECIMAL(10,2) NOT NULL,
    min_order_amount DECIMAL(10,2) DEFAULT 0,
    max_uses INT,
    used_count INT DEFAULT 0,
    applicable_to VARCHAR(20) DEFAULT 'all', -- 'all', 'cinema', 'movie'
    ref_id UUID,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 6. Update Bookings for PRD
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS promotion_id UUID REFERENCES promotions(id);
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS vat_amount DECIMAL(12,2) DEFAULT 0;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(12,2) DEFAULT 0;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS qr_code TEXT;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS booked_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

-- 7. Reviews
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    movie_id UUID REFERENCES movies(id) ON DELETE CASCADE,
    booking_id UUID REFERENCES bookings(id),
    rating INT CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'hidden')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 8. Audit Log
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    detail_json JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 9. Add Role to Users if not exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'customer' CHECK (role IN ('customer', 'staff', 'admin'));
