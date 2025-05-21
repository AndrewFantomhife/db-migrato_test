-- +goose Up
CREATE TABLE IF NOT EXISTS lab_adiabatic_expansion_results (
    id SERIAL PRIMARY KEY,
    h1i INTEGER NOT NULL,                   -- h1I, мм
    h2i INTEGER NOT NULL,                   -- h2I, мм
    delta_h1 INTEGER NOT NULL,              -- ∆hI = h2I - h1I
    h1ii INTEGER NOT NULL,                  -- h1II, мм
    h2ii INTEGER NOT NULL,                  -- h2II, мм
    delta_h2 INTEGER NOT NULL,              -- ∆hII = h2II - h1II
    gamma NUMERIC(4,2) NOT NULL,            -- γ — коэффициент адиабаты
    delta_gamma NUMERIC(5,3),               -- ∆γ — отклонение от среднего
    delta_gamma_squared NUMERIC(6,4),        -- (∆γ)^2
    measurement_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_measurement_time ON lab_adiabatic_expansion_results(measurement_time);
-- +goose Down
DROP TABLE IF EXISTS lab_adiabatic_expansion_results;
DROP INDEX IF EXISTS idx_measurement_time;