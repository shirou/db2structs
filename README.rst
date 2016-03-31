db2structs
==============

db2structs produce golang structs from running DB.


Example::

   // UserInfo
   // +model
   type UserInfo struct {
           UserID        int64     `sql:"column:user_id;primary_key;not null"`
           Email         string    `sql:"column:email;not null"`
           Password      string    `sql:"column:password;not null"`
           Name          string    `sql:"column:name;not null"`
           CreateDate    time.Time `sql:"column:create_date;not null"`
           UpdateDate    time.Time `sql:"column:update_date;not null"`
   }

install
==========

::

  $ go get github.com/shirou/db2structs


How to use
===============

Create JSON configuration file.

::

   {
     "db_type": "mysql",
     "db_user": "db",
     "db_host": "localhost",
     "db_port": 3306,
     "db_password": "",
     "db_name": "test",
     "output_file": "db_structs.go",
     "pkg_name": "models",
     "struct_tag" :"+test",
     "sql_tag": "sql"
   }

Then, just type

::

   $ db2structs -json example.json


Configuration
=================

db_type
  DB type. currently, only supports MySQL.
db_user
  DB user
db_host
  DB host
db_port
  DB port
db_password
  DB password
db_name
  DB name
output_file
  Output target file name. If omitted, printed out to StdOut.
pkg_name
  Package name of targetted file.
struct_tag
  If specified, it is inserted to a comment part of each structs.
sql_tag: sql
  If specified, `primary_key` or else are inserted to fields as SQL tag.


These environmental variables are used to override json configuration.

- MYSQL_HOST
- MYSQL_PORT
- MYSQL_DATABASE
- MYSQL_USER
- MYSQL_PASSWORD


Fork origin
==============

This package is originally developd by asdf072 at https://github.com/asdf072/struct-create. Great thanks.

License
=========

Apache License
