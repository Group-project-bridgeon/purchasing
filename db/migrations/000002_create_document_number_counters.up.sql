CREATE TABLE document_number_counters (
    document_type VARCHAR(50) NOT NULL,
    year INTEGER NOT NULL,
    last_number BIGINT NOT NULL DEFAULT 0,

    PRIMARY KEY (document_type, year),

    CONSTRAINT document_number_counters_year_check
        CHECK (year >= 2000),

    CONSTRAINT document_number_counters_last_number_check
        CHECK (last_number >= 0)
);