CREATE DATABASE prago_test CHARACTER SET utf8 DEFAULT COLLATE utf8_unicode_ci;
CREATE USER 'prago'@'localhost' IDENTIFIED BY 'prago';
GRANT ALL ON prago_test.* TO 'prago'@'localhost';
FLUSH PRIVILEGES;