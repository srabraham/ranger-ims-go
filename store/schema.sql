create table SCHEMA_INFO (
    VERSION smallint not null
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

insert into SCHEMA_INFO (VERSION) values (13);


create table EVENT (
    ID   integer      not null auto_increment,
    NAME varchar(128) not null,

    primary key (ID),
    unique key (NAME)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


create table CONCENTRIC_STREET (
    EVENT integer      not null,
    ID    varchar(16)  not null,
    NAME  varchar(128) not null,

    primary key (EVENT, ID)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


create table INCIDENT_TYPE (
    ID     integer      not null auto_increment,
    NAME   varchar(128) not null,
    HIDDEN boolean      not null,

    primary key (ID),
    unique key (NAME)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

insert into INCIDENT_TYPE (NAME, HIDDEN) values ('Admin', 0);
insert into INCIDENT_TYPE (NAME, HIDDEN) values ('Junk' , 0);


create table REPORT_ENTRY (
    ID        integer     not null auto_increment,
    AUTHOR    varchar(64) not null,
    TEXT      text        not null,
    CREATED   double      not null,
    `GENERATED` boolean     not null,
    STRICKEN  boolean     not null,

    ATTACHED_FILE varchar(128),

    -- FIXME: AUTHOR is an external non-primary key.
    -- Primary key is DMS Person ID.

    primary key (ID)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


create table INCIDENT (
    EVENT    integer  not null,
    NUMBER   integer  not null,
    CREATED  double   not null,
    PRIORITY tinyint  not null,

    STATE enum(
        'new', 'on_hold', 'dispatched', 'on_scene', 'closed'
    ) not null,

    SUMMARY varchar(1024),

    LOCATION_NAME          varchar(1024),
    LOCATION_CONCENTRIC    varchar(64),
    LOCATION_RADIAL_HOUR   tinyint,
    LOCATION_RADIAL_MINUTE tinyint,
    LOCATION_DESCRIPTION   varchar(1024),

    foreign key (EVENT) references EVENT(ID),

    foreign key (EVENT, LOCATION_CONCENTRIC)
    references CONCENTRIC_STREET(EVENT, ID),

    primary key (EVENT, NUMBER)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


create table INCIDENT__RANGER (
    ID              integer     not null auto_increment,
    EVENT           integer     not null,
    INCIDENT_NUMBER integer     not null,
    RANGER_HANDLE   varchar(64) not null,

    foreign key (EVENT) references EVENT(ID),
    foreign key (EVENT, INCIDENT_NUMBER) references INCIDENT(EVENT, NUMBER),

    -- FIXME: RANGER_HANDLE is an external non-primary key.
    -- Primary key is DMS Person ID.

    primary key (ID)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

create index `INCIDENT__RANGER_EVENT_INCIDENT_NUMBER_index`
    on `INCIDENT__RANGER` (EVENT, INCIDENT_NUMBER);


create table INCIDENT__INCIDENT_TYPE (
    EVENT           integer not null,
    INCIDENT_NUMBER integer not null,
    INCIDENT_TYPE   integer not null,

    foreign key (EVENT) references EVENT(ID),
    foreign key (EVENT, INCIDENT_NUMBER) references INCIDENT(EVENT, NUMBER),
    foreign key (INCIDENT_TYPE) references INCIDENT_TYPE(ID),

    primary key (EVENT, INCIDENT_NUMBER, INCIDENT_TYPE)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


create table INCIDENT__REPORT_ENTRY (
    EVENT           integer not null,
    INCIDENT_NUMBER integer not null,
    REPORT_ENTRY    integer not null,

    foreign key (EVENT) references EVENT(ID),
    foreign key (EVENT, INCIDENT_NUMBER) references INCIDENT(EVENT, NUMBER),
    foreign key (REPORT_ENTRY) references REPORT_ENTRY(ID),

    primary key (EVENT, INCIDENT_NUMBER, REPORT_ENTRY)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


create table EVENT_ACCESS (
    ID         integer      not null auto_increment,
    EVENT      integer      not null,
    EXPRESSION varchar(128) not null,

    MODE     enum ('read', 'write', 'report') not null,
    VALIDITY enum ('always', 'onsite') not null default 'always',

    foreign key (EVENT) references EVENT(ID),

    primary key (ID)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


create table FIELD_REPORT (
    EVENT   integer  not null,
    NUMBER  integer  not null,
    CREATED double   not null,

    SUMMARY         varchar(1024),
    INCIDENT_NUMBER integer,

    foreign key (EVENT) references EVENT(ID),
    foreign key (EVENT, INCIDENT_NUMBER) references INCIDENT(EVENT, NUMBER),

    primary key (EVENT, NUMBER)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


create table FIELD_REPORT__REPORT_ENTRY (
    EVENT                  integer not null,
    FIELD_REPORT_NUMBER    integer not null,
    REPORT_ENTRY           integer not null,

    foreign key (EVENT) references EVENT(ID),
    foreign key (EVENT, FIELD_REPORT_NUMBER)
        references FIELD_REPORT(EVENT, NUMBER),
    foreign key (REPORT_ENTRY) references REPORT_ENTRY(ID),

    primary key (EVENT, FIELD_REPORT_NUMBER, REPORT_ENTRY)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
