create role backendrole with login superuser password 'backendrole';


create table if not exists public.users(
    id int generated always as identity,
    login varchar not null unique,
    created_at timestamp not null default now(),
    password_hash varchar not null,
    auth_cookie varchar not null unique,
    credit_card_info varchar not null,
    cashback int,

    primary key (id)
);

create index if not exists users_login_index on public.users(login);
create index if not exists users_auth_cookie_index on public.users(auth_cookie);


create table if not exists public.images(
    id int generated always as identity,
    owner_id int,
    filename varchar not null,
    sha256 varchar not null,

    primary key(id),

    foreign key(owner_id) references public.users(id)
);

create index if not exists images_owner_id_index on public.images(owner_id);
create index if not exists images_sha256_index on public.images(sha256);
create index if not exists images_filename_index on public.images(filename);


create table if not exists public.wares(
    id int generated always as identity,
    seller_id int,
    title varchar not null,
    description varchar not null,
    price int,
    service_fee int,
    image_id int,

    primary key(id),

    foreign key(seller_id) references public.users(id),
    foreign key(image_id) references public.images(id)
);

create index if not exists wares_seller_id_index on public.wares(seller_id);


create table if not exists public.purchases(
    id int generated always as identity,
    ware_id int,
    buyer_id int,

    primary key(id),

    foreign key(ware_id) references public.wares(id),
    foreign key(buyer_id) references public.users(id)
);

create index if not exists purchases_ware_id_index on public.purchases(ware_id);
create index if not exists purchases_bayer_id_index on public.purchases(buyer_id);


grant select, insert, update on all tables in schema public to backendrole;

