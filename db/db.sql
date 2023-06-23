CREATE COLLATION nickname_case_insensitive(
    provider = icu,
    locale = 'und-u-ks-level2',
    deterministic = false
    );

create table if not exists Users(
    id serial primary key,
    nickname text collate "C" not null,
    fullname varchar(50) not null,
    about text,
    email varchar(256) not null
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
    slug varchar(40),
    created bigint not null,
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
    created bigint not null,
    forum_id int not null references Forum,
    path int[] not null default array[]::int[]
);

create table if not exists Vote(
    id serial primary key,
    user_id int not null references Users,
    thread_id int not null references Thread,
    positive_voice bool not null,
    unique(user_id, thread_id)
);


create table if not exists ForumUsers(
    id serial primary key,
    user_id int not null references users,
    forum_id int not null references forum
);

-- Collations:


-- Triggers:

--Triggers for vote processing
create or replace function process_thread_vote() returns trigger as $thread_vote$
    begin
        if new.positive_voice then
            update Thread set vote_count = vote_count + 1 where id = new.thread_id;
        else
            update Thread set vote_count = vote_count - 1 where id = new.thread_id;
        end if;
        return null;
    end
$thread_vote$ LANGUAGE plpgsql;

create or replace trigger trigger_thread_vote
    after insert
    on Vote
    for each row
    execute procedure process_thread_vote();

create or replace function process_thread_revote() returns trigger as $thread_revote$
    begin
        if new.positive_voice then
            update Thread set vote_count = vote_count + 2 where id = old.thread_id;
        else
            update Thread set vote_count = vote_count - 2 where id = old.thread_id;
        end if;
        return null;
    end
$thread_revote$ LANGUAGE plpgsql;

create or replace trigger trigger_thread_revote
    after update
    on Vote
    for each row
    execute procedure process_thread_revote();

-- triggers for path processing:

create or replace function process_post_insert() returns trigger as $post_insert$
    declare
        current_node int;
        parent_node record;
        new_path_array int[];
        test bool;
    begin
        new.path = (select path from post where id = new.parent_id) || NEW.id;
--         if new.parent_id is not null then
--             current_node = new.parent_id;
--
--             new_path_array = (select path from post where id=current_node);
--
--             new_path_array = array_append(new_path_array, new.id);
--
--             new.path = new_path_array;
--         else
--             new.path = [new.id];
--         end if;
        return new;
    end
$post_insert$ LANGUAGE plpgsql;

create or replace trigger trigger_post_insert
    before insert on Post
    for each row
    execute procedure process_post_insert();

create or replace function process_thread_inc() returns trigger as $thread_inc$
begin
    update forum set thread_count = thread_count + 1 where id=new.forum_id;
    return null;
end
$thread_inc$ LANGUAGE plpgsql;

create or replace trigger trigger_thread_inc
    after insert on thread
    for each row
    execute procedure process_thread_inc();

create or replace function process_post_inc() returns trigger as $post_inc$
begin
    update forum set post_count = post_count + 1 where id=new.forum_id;
    return null;
end
$post_inc$ LANGUAGE plpgsql;

create or replace trigger trigger_post_inc
    after insert on post
    for each row
execute procedure process_post_inc();



-- Indexes

create unique index on forum(lower(slug)) include(id);


create unique index on users(lower(email));
create unique index on users(lower(nickname));

create unique index on thread(lower(slug));

create index on post ((path[1]));

create index on post ((path[2:]));
