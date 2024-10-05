CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS es_aggregate (
  id              UUID     PRIMARY KEY,
  version         INTEGER  NOT NULL,
  aggregate_type  TEXT     NOT NULL
);

CREATE INDEX IF NOT EXISTS IDX_ES_AGGREGATE_AGGREGATE_TYPE ON es_aggregate (aggregate_type);

CREATE TABLE IF NOT EXISTS es_event (
  id              BIGSERIAL  PRIMARY KEY,
  transaction_id  XID8       NOT NULL,
  aggregate_id    UUID       NOT NULL REFERENCES es_aggregate (id),
  version         INTEGER    NOT NULL,
  event_type      TEXT       NOT NULL,
  event_data      JSONB      NOT NULL,
  created_at      TIMESTAMP  NOT NULL,
  UNIQUE (aggregate_id, version)
);

CREATE INDEX IF NOT EXISTS IDX_ES_EVENT_TRANSACTION_ID_ID ON es_event (transaction_id, id);
CREATE INDEX IF NOT EXISTS IDX_ES_EVENT_AGGREGATE_ID ON es_event (aggregate_id);
CREATE INDEX IF NOT EXISTS IDX_ES_EVENT_VERSION ON es_event (version);

CREATE TABLE IF NOT EXISTS es_aggregate_snapshot (
  aggregate_id  UUID     NOT NULL REFERENCES es_aggregate (id),
  version       INTEGER  NOT NULL,
  event_data    JSONB    NOT NULL,
  PRIMARY KEY (aggregate_id, version)
);

CREATE INDEX IF NOT EXISTS IDX_ES_AGGREGATE_SNAPSHOT_AGGREGATE_ID ON es_aggregate_snapshot (aggregate_id);
CREATE INDEX IF NOT EXISTS IDX_ES_AGGREGATE_SNAPSHOT_VERSION ON es_aggregate_snapshot (version);

CREATE TABLE IF NOT EXISTS es_event_subscription (
  subscription_name    TEXT    PRIMARY KEY,
  last_transaction_id  XID8    NOT NULL,
  last_event_id        BIGINT  NOT NULL
);
