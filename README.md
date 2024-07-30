# Prerequisites
- PostgresDB running on :5432 by default.
- Creating `greenlight` database

      postgres=# CREATE DATABASE greenlight 

- Creating `greenlight` user for database
      
      postgres=# CREATE USER greenlight;
- Creating DSN

      $ export GREENLIGHT_DB_DSN='postgres://greenlight:pa55word@localhost/greenlight'
      $ source $HOME/.profile
      $ echo $GREENLIGHT_DB_DSN
- 
- Install `citext` extension

      greenlight=# CREATE EXTENSION citext;


