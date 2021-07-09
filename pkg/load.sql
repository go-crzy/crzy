create table if not exists friends(
   id int,
   name varchar(255)
);
truncate table friends;
insert into friends(id, name) values (1, 'Kilian');
insert into friends(id, name) values (2, 'Sacha');
