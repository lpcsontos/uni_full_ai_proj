-- users
CREATE TABLE IF NOT EXISTS users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- refresh tokens (hashed)
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  token_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE (user_id, token_hash)
);

-- ingredients (normalized)
CREATE TABLE IF NOT EXISTS ingredients (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE
);

-- recipes
CREATE TABLE IF NOT EXISTS recipes (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  servings INT,
  prep_time_minutes INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- recipe_ingredients
CREATE TABLE IF NOT EXISTS recipe_ingredients (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  recipe_id BIGINT NOT NULL,
  ingredient_id BIGINT NOT NULL,
  quantity DOUBLE NOT NULL,
  unit VARCHAR(50),
  FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE,
  FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE,
  UNIQUE (recipe_id, ingredient_id)
);

-- inventory items
CREATE TABLE IF NOT EXISTS inventory_items (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  ingredient_id BIGINT NOT NULL,
  quantity DOUBLE NOT NULL,
  unit VARCHAR(50),
  expiration_date DATE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (ingredient_id) REFERENCES ingredients(id),
  UNIQUE (user_id, ingredient_id)
);

-- shopping lists
CREATE TABLE IF NOT EXISTS shopping_lists (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  name VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS shopping_list_items (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  list_id BIGINT NOT NULL,
  ingredient_id BIGINT NOT NULL,
  quantity DOUBLE NOT NULL,
  unit VARCHAR(50),
  purchased BOOLEAN DEFAULT FALSE,
  FOREIGN KEY (list_id) REFERENCES shopping_lists(id) ON DELETE CASCADE,
  FOREIGN KEY (ingredient_id) REFERENCES ingredients(id)
);

-- indexes
CREATE INDEX idx_recipes_title ON recipes(title);
CREATE INDEX idx_recipe_ingredient ON recipe_ingredients(ingredient_id);
CREATE INDEX idx_inventory_user_ing ON inventory_items(user_id, ingredient_id);
