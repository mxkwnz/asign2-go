CREATE TABLE payments (
                          id TEXT PRIMARY KEY,
                          order_id TEXT NOT NULL,
                          transaction_id TEXT,
                          amount BIGINT NOT NULL,
                          status TEXT NOT NULL
);