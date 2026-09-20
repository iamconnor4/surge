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
    venue_id uuid NOT NULL,
    name text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,

    CONSTRAINT venue_sections_venue_id_fkey
        FOREIGN KEY (venue_id)
        REFERENCES venues (id)
);

CREATE INDEX venue_sections_venue_id_idx ON venue_sections (venue_id);


CREATE TABLE venue_seats (
    id uuid PRIMARY KEY,
    venue_section_id uuid NOT NULL,
    row_label text NOT NULL,
    seat_number text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,

    CONSTRAINT venue_seats_venue_section_id_row_label_seat_number_key
        UNIQUE (venue_section_id, row_label, seat_number),

    CONSTRAINT venue_seats_venue_section_id_fkey
        FOREIGN KEY (venue_section_id)
        REFERENCES venue_sections (id)
);

CREATE INDEX venue_seats_venue_section_id_idx ON venue_seats (venue_section_id);

-- Events

CREATE TABLE events (
    id uuid PRIMARY KEY,
    venue_id uuid NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    starts_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,

    CONSTRAINT events_venue_id_fkey
        FOREIGN KEY (venue_id)
        REFERENCES venues (id)
);

CREATE INDEX events_venue_id_idx ON events (venue_id);


CREATE TABLE event_seats (
    id uuid PRIMARY KEY,
    event_id uuid NOT NULL,
    venue_seat_id uuid NOT NULL,
    price_pence bigint NOT NULL,
    is_available boolean NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,

    CONSTRAINT event_seats_event_id_venue_seat_id_key
        UNIQUE (event_id, venue_seat_id),

    CONSTRAINT event_seats_price_pence_check
        CHECK (price_pence >= 0),

    CONSTRAINT event_seats_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES events (id),

    CONSTRAINT event_seats_venue_seat_id_fkey
        FOREIGN KEY (venue_seat_id)
        REFERENCES venue_seats (id)
);

CREATE INDEX event_seats_venue_seat_id_idx ON event_seats (venue_seat_id);

-- Orders

CREATE TYPE order_status AS ENUM ('pending', 'paid', 'cancelled', 'refunded');

CREATE TABLE orders (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
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
         OR (status in ('paid', 'refunded') AND paid_at IS NOT NULL)),

    CONSTRAINT orders_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users (id)
);

CREATE INDEX orders_user_id_idx ON orders (user_id);

-- Tickets

CREATE TABLE tickets (
    id uuid PRIMARY KEY,
    order_id uuid NOT NULL,
    event_seat_id uuid NOT NULL,
    price_pence bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz,

    CONSTRAINT tickets_event_seat_id_key
        UNIQUE (event_seat_id),

    CONSTRAINT tickets_price_pence_check
        CHECK (price_pence >= 0),

    CONSTRAINT tickets_order_id_fkey
        FOREIGN KEY (order_id)
        REFERENCES orders (id),

    CONSTRAINT tickets_event_seat_id_fkey
        FOREIGN KEY (event_seat_id)
        REFERENCES event_seats (id)
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
