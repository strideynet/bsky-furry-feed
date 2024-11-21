CREATE TABLE follow_tasks (
    id BIGSERIAL PRIMARY KEY,
    actor_did TEXT NOT NULL REFERENCES candidate_actors (did),
    next_try_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    tries INT DEFAULT 0 NOT NULL,
    should_unfollow BOOLEAN NOT NULL,
    finished_at TIMESTAMPTZ,
    last_error TEXT
);

CREATE INDEX follow_tasks_dates_idx ON public.follow_tasks (next_try_at, finished_at);
