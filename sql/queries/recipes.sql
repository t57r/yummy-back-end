-- name: GetRecipeByID :one
SELECT
  id,
  title,
  image,
  summary,
  source_url,
  prep_time_minutes,
  cook_time_minutes,
  serves,
  ingredients,
  preparation,
  tags
FROM recipes
WHERE id = $1;

-- name: CountRecipes :one
SELECT count(*)::bigint
FROM recipes r
WHERE (sqlc.narg(q)::text IS NULL OR r.title ILIKE ('%' || sqlc.narg(q) || '%'));

-- name: CountRecipesByCategoryName :one
SELECT count(*)::bigint
FROM recipes r
JOIN recipe_categories rc ON rc.recipe_id = r.id
JOIN categories c ON c.id = rc.category_id
WHERE c.name = $1
  AND (sqlc.narg(q)::text IS NULL OR r.title ILIKE ('%' || sqlc.narg(q) || '%'));

-- name: ListRecipes :many
SELECT
  r.id,
  r.title,
  r.image,
  r.summary,
  r.source_url,
  r.prep_time_minutes,
  r.cook_time_minutes,
  r.serves,
  r.ingredients,
  r.preparation,
  r.tags
FROM recipes r
WHERE (sqlc.narg(q)::text IS NULL OR r.title ILIKE ('%' || sqlc.narg(q) || '%'))
  AND (sqlc.narg(last_id)::bigint IS NULL OR r.id > sqlc.narg(last_id)::bigint)
ORDER BY r.id
LIMIT $1;

-- name: ListRecipesByCategoryName :many
SELECT
  r.id,
  r.title,
  r.image,
  r.summary,
  r.source_url,
  r.prep_time_minutes,
  r.cook_time_minutes,
  r.serves,
  r.ingredients,
  r.preparation,
  r.tags
FROM recipes r
JOIN recipe_categories rc ON rc.recipe_id = r.id
JOIN categories c ON c.id = rc.category_id
WHERE c.name = $1
  AND (sqlc.narg(q)::text IS NULL OR r.title ILIKE ('%' || sqlc.narg(q) || '%'))
  AND (sqlc.narg(last_id)::bigint IS NULL OR r.id > sqlc.narg(last_id)::bigint)
ORDER BY r.id
LIMIT $2;

-- name: RecipeExists :one
SELECT EXISTS(SELECT 1 FROM recipes WHERE id = $1);
