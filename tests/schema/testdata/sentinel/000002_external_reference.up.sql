ALTER TABLE sentinel_owned
	ADD COLUMN product_pid VARCHAR(30) NULL,
	ADD KEY ix_sentinel_owned_product_pid (product_pid),
	ADD CONSTRAINT fk_sentinel_owned_product
		FOREIGN KEY (product_pid)
		REFERENCES product (pid)
		ON DELETE RESTRICT
		ON UPDATE RESTRICT;
