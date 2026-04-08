INSERT INTO roles (code, name)
VALUES ('admin', 'Administrator');

INSERT INTO users (username, password_hash, display_name, status)
VALUES ('admin', '$2a$10$ZAYmqFjkizBLy9U4rExWruAtq1vVqFPoUicDxNhkVqzzzJPaekCqW', 'Administrator', 'active');

INSERT INTO user_roles (user_id, role_id)
VALUES (1, 1);

INSERT INTO install_tokens (token, description, max_uses, used_count)
VALUES ('install-token-demo', 'default bootstrap token', 100, 0);
