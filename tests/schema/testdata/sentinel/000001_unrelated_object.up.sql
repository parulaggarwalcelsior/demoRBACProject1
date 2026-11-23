CREATE TABLE sentinel_owned (
	id BIGINT NOT NULL AUTO_INCREMENT,
	label VARCHAR(64) NOT NULL,
	PRIMARY KEY (id),
	UNIQUE KEY uq_sentinel_owned_label (label)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;
