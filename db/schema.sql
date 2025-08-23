create table public.products
(
    id         uuid                     default uuid_generate_v4() not null
        primary key,
    name       varchar(255)                                        not null,
    price      numeric(10, 2)                                      not null,
    created_at timestamp with time zone default now()              not null,
    updated_at timestamp with time zone default now()              not null
);

create table public.users
(
    id         uuid                     default uuid_generate_v4() not null
        primary key,
    name       varchar(100)                                        not null,
    email      varchar(255)                                        not null
        unique,
    created_at timestamp with time zone default now()              not null,
    updated_at timestamp with time zone default now()              not null
);

