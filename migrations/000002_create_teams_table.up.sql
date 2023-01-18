CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    team_user UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE UNIQUE,
    team_name char varying(100) NOT NULL,
    team_picture char varying(512),
    is_indexed BOOLEAN NOT NULL DEFAULT false,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    version integer NOT NULL DEFAULT 1
);