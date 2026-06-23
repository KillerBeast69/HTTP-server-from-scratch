-- +goose Up
create table refresh_tokens(
    token text primary key,
    created_at timestamp default current_timestamp not null,
    updated_at timestamp default current_timestamp not null,
    expires_at timestamp not null,
    revoked_at timestamp,
    user_id uuid not null,
    constraint user_id_constraint foreign key (user_id) references users(id) on delete cascade 
);

-- +goose Down
drop table refresh_tokens;