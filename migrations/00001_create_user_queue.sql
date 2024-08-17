create table UserQueue (
    Name text primary key,
    Skill double precision not null,
    Latency double precision not null,
    QueuedAt timestamp not null,
    PosS integer not null default 0,
    PosL integer not null default 0
);