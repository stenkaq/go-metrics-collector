CREATE TABLE metrics (
    id VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    PRIMARY KEY (id, type),
    CONSTRAINT metrics_value_check CHECK (
        (type = 'counter' AND delta IS NOT NULL and value IS NULL) OR
        (type = 'gauge' AND delta IS NULL and value IS NOT NULL)
    )
)
