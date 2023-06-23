create table if not exists Users(
    id serial primary key,
    nickname varchar(50) unique not null,
    fullname varchar(50) not null,
    about text,
    email varchar(256) unique not null
);

create table if not exists Forum(
    id serial primary key,
    author_id int not null references Users,
    slug text unique not null,
    title text not null,
    post_count int not null default 0,
    thread_count int not null default 0
);

create table if not exists Thread(
    id serial primary key,
    author_id int not null references Users,
    message text not null,
    title text not null,
    forum_id int not null references Forum,
    slug varchar(40) unique,
    created date not null default now(),
    vote_count int not null default 0,
    constraint valid_thread_slug check ( slug !~* '^[0-9]+$')
);

create table if not exists Post(
    id serial primary key,
    parent_id int references Post,
    author_id int not null references Users,
    message text not null,
    edited bool not null default false,
    thread_id int not null references Thread,
    created date not null default now(),
    forum_id int not null references Forum,
    path int[] not null default array[]::int[],
    childs int[]
);

create table if not exists Vote(
    id serial primary key,
    user_id int not null references Users,
    thread_id int not null references Thread,
    positive_voice bool not null
);


-- Collations:
CREATE COLLATION nickname_case_insensitive(
    provider = icu,
    locale = 'und-u-ks-level2',
    deterministic = false
    );

-- Triggers:

--Triggers for vote processing
create or replace function process_thread_vote() returns trigger as $thread_vote$
    begin
        if new.positive_voice then
            update Thread set vote_count = vote_count + 1 where id = new.thread_id;
        else
            update Thread set vote_count = vote_count - 1 where id = new.thread_id;
        end if;
    end
$thread_vote$ LANGUAGE plpgsql;

create or replace trigger trigger_thread_vote
    after insert
    on Vote
    for each row
    execute procedure process_thread_vote();

create or replace function process_thread_unvote() returns trigger as $thread_unvote$
    begin
        if old.positive_voice then
            update Thread set vote_count = vote_count - 1 where id = old.thread_id;
        else
            update Thread set vote_count = vote_count + 1 where id = old.thread_id;
        end if;
    end
$thread_unvote$ LANGUAGE plpgsql;

create or replace trigger trigger_thread_unvote
    after delete
    on Vote
    for each row
    execute procedure process_thread_unvote();

-- triggers for path processing:

create or replace function process_post_insert() returns trigger as $post_insert$
    declare
        current_node int;
        parent_node record;
        new_path_array int[];
        test bool;
    begin
        if new.parent_id != 0 then
            current_node = new.parent_id;

            new_path_array = (select path from post where id=current_node);

            new_path_array = array_append(new_path_array, current_node);

            new.path = new_path_array;
        end if;
        return new;
    end
$post_insert$ LANGUAGE plpgsql;

create or replace trigger trigger_post_insert
    before insert on Post
    for each row
    execute procedure process_post_insert();

-- Indexes

create index on forum(slug) include(id);
