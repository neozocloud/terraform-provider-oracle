# Oracle Docker

Before build, need to download this - https://www.oracle.com/webapps/redirect/signon?nexturl=https://download.oracle.com/otn/linux/oracle19c/190000/LINUX.ARM64_1919000_db_home.zip 

```shell
$ mv LINUX.ARM64_1919000_db_home.zip 19.3.0/
$ cd 19.3.0
$ docker build -t oracle/database:19.3.0-ee .
$ docker run -d --name oracle19c -p 1521:1521 -p 5500:5500 -e ORACLE_PWD=MyPassword123 -e ORACLE_CHARACTERSET=AL32UTF8 oracle/database:19.3.0-ee
```