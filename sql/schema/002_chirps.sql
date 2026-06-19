-- +goose Up
create table chirps (
    id uuid primary key default gen_random_uuid(),
    created_at timestamp default current_timestamp not null,
    updated_at timestamp default current_timestamp not null,
    body text not null,
    user_id uuid not null,
    constraint user_id_constraint foreign key (user_id) references users (id) on delete cascade
);

-- +goose Down
drop table chirps;