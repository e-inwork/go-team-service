CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    email text UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    first_name char varying(100) NOT NULL,
    last_name char varying(100) NOT NULL,
    activated bool NOT NULL DEFAULT false,
    version integer NOT NULL DEFAULT 1
);