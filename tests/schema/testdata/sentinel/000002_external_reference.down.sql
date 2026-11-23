ALTER TABLE sentinel_owned
	DROP FOREIGN KEY fk_sentinel_owned_product,
	DROP KEY ix_sentinel_owned_product_pid,
	DROP COLUMN product_pid;
