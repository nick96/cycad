-- Write your migrate up statements here

-- Use the uuid-ossp so we can generate uuids in the actual database, rather
-- than having to rely on whatever langauge we're talking to the database in.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


CREATE SCHEMA editor
  CREATE TABLE nodes(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4()

    -- At the moment names are unique by themselves because there are no users.
    -- Once there are users names would be unique to users (or teams).
    ,name VARCHAR(256) NOT NULL UNIQUE -- 256 is pretty much randomly chosen but it should be long enough...
    ,content TEXT -- Content can be null for empty documents
  )


  -- Links from one node to another. This is for forward links, backlinks are
  -- just "to" -> "from". Links aren't unique on from-to because there could be
  -- multiple links between two pages.
  CREATE TABLE links(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4()

    -- Node the link is from.
    ,fromNode uuid REFERENCES nodes(id) NOT NULL
    -- Node the link is to.
    ,toNode uuid REFERENCES nodes(id) NOT NULL
    -- Character in the from node's content that the link starts.
    ,pos int NOT NULL
  )
;

---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.

DROP TABLE IF EXISTS links;
DROP TABLE IF EXISTS linkType;
DROP TABLE IF EXISTS nodes;
DROP SCHEMA IF EXISTS editor;
