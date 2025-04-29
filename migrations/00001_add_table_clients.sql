-- +goose Up
-- +goose StatementBegin
create table clients (
    id BIGSERIAL PRIMARY KEY,
    ip text not null,
    capacity int not NULL,
    rate_per_sec int not null
);
create INDEX indx_ip on clients(ip);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table clients;
-- +goose StatementEnd
