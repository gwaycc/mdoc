Online server tool to made markdown document.

# Run

## No authentication
```shell
go build
./mdoc daemon --auth-mode=false # TODO
```
more help run "./mdoc --help"  
then open http://localhost:8080 in browser.

## Authentication(Default mode)
```shell
go build
./mdoc daemon --auth-mode=true
```

Open another console, add a user to sqlite db.  
passwd is 'hello', see tools/auth/auth_test.go#TestHashPasswd
```
sudo apt-get install sqlite3
sqlite3 ./data/mdoc.db
INSERT INTO user_info(id,`passwd`,nick_name,memo)VALUES('admin','7628d9fbecd3683d02276b6176b0ee13','admin','system init');
.q

# modify the passwd
./mdoc user --admin-user=admin --admin-pwd=hello reset --username=admin --passwd=<newpasswd>

# add a new user
./mdoc user --admin-user=admin --admin-pwd=<newpasswd> add --username=newone --passwd=<newpasswd>
```
more help run "./mdoc --help"  
then open http://localhost:8080 in browser.
