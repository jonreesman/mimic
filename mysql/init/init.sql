use app;
CREATE TABLE users(user_id VARCHAR(255) PRIMARY KEY, source int NOT NULL);
CREATE TABLE discord_messages(message_id VARCHAR(255) PRIMARY KEY, user_id VARCHAR(255), channel_id VARCHAR(255), msg TEXT, time_stamp BIGINT, FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE);
CREATE TABLE element_messages(message_id VARCHAR(255) PRIMARY KEY, user_id VARCHAR(255), channel_id VARCHAR(255), msg TEXT, time_stamp BIGINT, FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE);