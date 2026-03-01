-- name: AddFavorite :exec
INSERT INTO favorites (user_id, recipe_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveFavorite :exec
DELETE FROM favorites
WHERE user_id = $1 AND recipe_id = $2;

-- name: CountFavoritesByUser :one
SELECT count(*)::bigint
FROM favorites
WHERE user_id = $1;

-- name: ListFavoriteRecipesByUser :many
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
FROM favorites f
JOIN recipes r ON r.id = f.recipe_id
WHERE f.user_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- name: IsFavorite :one
SELECT EXISTS(
  SELECT 1 FROM favorites
  WHERE user_id = $1 AND recipe_id = $2
);
