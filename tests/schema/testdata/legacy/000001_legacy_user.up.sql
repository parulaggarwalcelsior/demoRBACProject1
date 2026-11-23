CREATE TABLE `user` (
	email VARCHAR(60) NOT NULL,
	name VARCHAR(30) NULL DEFAULT NULL,
	mobile BIGINT NULL DEFAULT NULL,
	address VARCHAR(250) NULL DEFAULT NULL,
	pincode INT NULL DEFAULT NULL,
	password VARCHAR(20) NULL DEFAULT NULL,
	PRIMARY KEY (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;

INSERT INTO `user` (email, name, mobile, address, pincode, password) VALUES
	('legacy-one@example.test', 'Legacy One', 15550001001, '101 Synthetic Fixture Way', 10001, NULL),
	('legacy-two@example.test', NULL, NULL, NULL, NULL, NULL);
