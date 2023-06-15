create table if not exists Users(
    id serial primary key,
    nickname varchar(50) unique not null,
    fullname varchar(50) not null,
    about text,
    email varchar(256) unique not null
);

create table if not exists Forum(
    id serial primary key,
    author_id int references Users,
    slug text unique not null,
    post_count int not null default 0,
    thread_count int not null default 0
);

create table if not exists Thread(
    id serial primary key,
    author_id int references Users,
    desription text not null,
    name text not null,
    forum_id int references Forum,
    slug varchar(40),
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
    thread_id int references Thread,
    created date not null default now()
);

create table if not exists Vote(
    id serial primary key,
    user_id int not null references Users,
    thread_id int not null references Thread,
    positive_voice bool not null
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

-- Indexes

