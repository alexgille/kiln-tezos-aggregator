
CREATE TABLE delegation (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  inserted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  block_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
  operation_id BIGINT NOT NULL UNIQUE,
  amount BIGINT NOT NULL,
  level INTEGER NOT NULL,
  sender TEXT NOT NULL,
  block_hash TEXT NOT NULL
);

COMMENT ON COLUMN delegation.id IS 'Unique ID of a delegation';
COMMENT ON COLUMN delegation.inserted_at IS 'Timestamp with time zone of insertion in database';
COMMENT ON COLUMN delegation.block_timestamp IS 'Timestamp with time zone of the block of the delegation operation';
COMMENT ON COLUMN delegation.operation_id IS 'Unique ID of the operation, stored in the TzKT indexer database';
COMMENT ON COLUMN delegation.amount IS 'Sender balance at the time of delegation operation (aka delegation amount)';
COMMENT ON COLUMN delegation.level IS 'The height of the block from the genesis block, in which the operation was included';
COMMENT ON COLUMN delegation.sender IS 'Account address (public key hash) of the delegated account';
COMMENT ON COLUMN delegation.block_hash IS 'Hash of the block, in which the operation was included';

---- create above / drop below ----

DROP TABLE delegation CASCADE;
