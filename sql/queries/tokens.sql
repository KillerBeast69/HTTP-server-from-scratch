-- name: CreateRefreshToken :one
insert into refresh_tokens(token, created_at, updated_at, expires_at, revoked_at, user_id)
values (
    $1,
    now(),
    now(),
    $2,
    null,
    $3
)
returning *;

-- name: GetUserFromRefreshToken :one
select users.* from users
join refresh_tokens on users.id = refresh_tokens.user_id
where refresh_tokens.token = $1
and refresh_tokens.revoked_at is null
and refresh_tokens.expires_at > now();

-- name: RevokeRefreshToken :exec
update refresh_tokens
set revoked_at = now(),
    updated_at = now()
where token = $1;