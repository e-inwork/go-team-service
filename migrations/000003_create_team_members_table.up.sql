CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    team_member_team UUID NOT NULL REFERENCES teams (id) ON DELETE CASCADE,
    team_member_user UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    UNIQUE(team_member_team, team_member_user)
);