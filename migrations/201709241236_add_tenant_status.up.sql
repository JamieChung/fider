ALTER TABLE tenants ADD status INT NOT NULL DEFAULT 1;
ALTER TABLE tenants ALTER COLUMN status DROP DEFAULT;