-- +goose Up

CREATE TABLE scenarios (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    definition JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT scenarios_name_length
        CHECK (char_length(name) BETWEEN 1 AND 120),

    CONSTRAINT scenarios_definition_object
        CHECK (jsonb_typeof(definition) = 'object')
);

CREATE TABLE analyses (
    id UUID PRIMARY KEY,
    scenario_id UUID NOT NULL,
    scenario_snapshot JSONB NOT NULL,
    safe BOOLEAN NOT NULL,
    result JSONB NOT NULL,
    engine_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT analyses_scenario_fk
        FOREIGN KEY (scenario_id)
        REFERENCES scenarios(id)
        ON DELETE CASCADE,

    CONSTRAINT analyses_snapshot_object
        CHECK (jsonb_typeof(scenario_snapshot) = 'object'),

    CONSTRAINT analyses_result_object
        CHECK (jsonb_typeof(result) = 'object')
);

CREATE INDEX analyses_scenario_created_idx
    ON analyses (scenario_id, created_at DESC);

CREATE INDEX scenarios_updated_idx
    ON scenarios (updated_at DESC);

-- +goose Down

DROP TABLE analyses;
DROP TABLE scenarios;
