CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS orders (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(), 
  version     INTEGER NOT NULL,
  name        TEXT NOT NULL,
  order_items JSONB NOT NULL,
  is_submitted BOOLEAN NOT NULL DEFAULT false,
  created_at  TIMESTAMP NOT NULL DEFAULT now(),
  updated_at  TIMESTAMP
) 
