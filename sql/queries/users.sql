-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    now(),
    now(),
    $1,
    $2
)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users
WHERE TRUE
RETURNING *;

-- name: GetUser :one
select * from users 
where email = $1;

-- name: UpdateUser :one
update users
set email = $1,
    hashed_password = $2,
    updated_at = now()
where id = $3
returning *;

-- name: Chirpyred :one
update users
set is_chirpy_red = true,
    updated_at = now()
where id = $1
returning *;