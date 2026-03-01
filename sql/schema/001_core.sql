CREATE TABLE IF NOT EXISTS categories (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS recipes (
  id BIGINT PRIMARY KEY,
  title TEXT NOT NULL,
  image TEXT,
  summary TEXT,
  source_url TEXT,

  prep_time_minutes INT,
  cook_time_minutes INT,
  serves INT,

  ingredients TEXT[] NOT NULL DEFAULT '{}',
  preparation TEXT[] NOT NULL DEFAULT '{}',
  tags TEXT[] NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS recipe_categories (
  recipe_id BIGINT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
  category_id INT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
  PRIMARY KEY (recipe_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_recipes_title ON recipes USING btree (title);
CREATE INDEX IF NOT EXISTS idx_recipes_prep_time ON recipes USING btree (prep_time_minutes);
CREATE INDEX IF NOT EXISTS idx_recipes_cook_time ON recipes USING btree (cook_time_minutes);
