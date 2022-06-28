DELIMITER $$
CREATE PROCEDURE insertUser(IN id VARCHAR(255), IN name VARCHAR(255), IN email VARCHAR(255), IN created VARCHAR(255))
BEGIN
	INSERT INTO User (ID, Name, Email, Created) VALUES (id, name, email, created);
END $$
DELIMITER ;