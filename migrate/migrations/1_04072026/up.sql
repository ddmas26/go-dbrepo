create table IF NOT EXISTS inventory (
    id UUID primary key,
    name varchar(255) not null,
    quantity int not null,
    price decimal(10, 2) not null,
    created_at TIMESTAMPTZ default current_timestamp,
    updated_at TIMESTAMPTZ
)