-- +goose Up

-- Users

CREATE TABLE users (
    id uuid PRIMARY KEY,
    email text NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    CONSTRAINT users_email_key
        UNIQUE (email),
    CONSTRAINT users_email_check
        CHECK (email = lower(email))
);

-- Venues

CREATE TABLE venues (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE TABLE venue_sections (
    id uuid PRIMARY KEY,
    venue_id uuid REFERENCES venues (id) NOT NULL,
    name text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE INDEX venue_sections_venue_id_idx ON venue_sections (venue_id);


CREATE TABLE venue_seats (
    id uuid PRIMARY KEY,
    section_id uuid REFERENCES venue_sections (id) NOT NULL,
    row_label text NOT NULL,
    seat_number text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    CONSTRAINT venue_seats_section_id_row_label_seat_number_key
        UNIQUE (section_id, row_label, seat_number)
);

-- Events

CREATE TABLE events (
    id uuid PRIMARY KEY,
    venue_id uuid REFERENCES venues (id) NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    starts_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz
);

CREATE INDEX events_venue_id_idx ON events (venue_id);


CREATE TABLE event_seats (
    id uuid PRIMARY KEY,
    event_id uuid REFERENCES events (id) NOT NULL,
    venue_seat_id uuid REFERENCES venue_seats (id) NOT NULL,
    price_pence bigint NOT NULL,
    is_available boolean NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    CONSTRAINT event_seats_event_id_venue_seat_id_key
        UNIQUE (event_id, venue_seat_id),
    CONSTRAINT event_seats_price_pence_check
        CHECK (price_pence >= 0)
);

CREATE INDEX event_seats_venue_seat_id_idx ON event_seats (venue_seat_id);

-- Orders

CREATE TYPE order_status AS ENUM ('pending', 'paid', 'cancelled', 'refunded');

CREATE TABLE orders(
    id uuid PRIMARY KEY,
    user_id uuid REFERENCES users (id) NOT NULL,
    status order_status NOT NULL,
    total_pence bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    paid_at timestamptz,
    CONSTRAINT orders_total_pence_check
        CHECK (total_pence >= 0),
    CONSTRAINT orders_paid_at_status_check
        CHECK (
            (status in ('pending', 'cancelled') AND paid_at IS NULL)
         OR (status in ('paid', 'refunded') AND paid_at IS NOT NULL))
);

CREATE INDEX orders_user_id_idx ON orders (user_id);

-- Tickets

CREATE TABLE tickets (
    id uuid PRIMARY KEY,
    order_id uuid REFERENCES orders (id) NOT NULL,
    event_seat_id uuid REFERENCES event_seats (id) NOT NULL,
    price_pence bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,
    CONSTRAINT tickets_event_seat_id_key
        UNIQUE (event_seat_id),
    CONSTRAINT tickets_price_pence_check
        CHECK (price_pence >= 0)
);

CREATE INDEX tickets_order_id_idx ON tickets (order_id);

-- +goose Down
DROP TABLE tickets;
DROP TABLE orders;
DROP TABLE event_seats;
DROP TABLE events;
DROP TABLE venue_seats;
DROP TABLE venue_sections;
DROP TABLE venues;
DROP TABLE users;

DROP TYPE order_status;
