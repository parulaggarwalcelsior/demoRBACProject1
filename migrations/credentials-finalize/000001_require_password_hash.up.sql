CREATE PROCEDURE schema_guard_finalize_credentials()
BEGIN
    IF EXISTS (
        SELECT 1 FROM `user`
        WHERE password_hash IS NULL OR must_reset_password <> 0
    ) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'refuse credential finalization: password_hash NULL or must_reset_password truthy';
    END IF;
    ALTER TABLE `user` MODIFY COLUMN password_hash VARCHAR(255) NOT NULL;
END;
CALL schema_guard_finalize_credentials();
DROP PROCEDURE schema_guard_finalize_credentials;
