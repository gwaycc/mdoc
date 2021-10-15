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

## Set a admin account for login
Open another console, add a user to sqlite db.  
passwd is 'hello', see tools/auth/auth_test.go#TestHashPasswd
```
sudo apt-get install sqlite3
sqlite3 ./data/mdoc.db
INSERT INTO user_info(id,`passwd`,nick_name,kind,memo)VALUES('admin','7628d9fbecd3683d02276b6176b0ee13','admin',1,'system init');
.q

# modify the passwd
/mdoc user --admin-user=admin --admin-pwd=hello reset --username=admin --passwd=<newpasswd>

# add a new user
./mdoc user --admin-user=admin --admin-pwd=<newpasswd> add --username=newone --passwd=<newpasswd>
```
more help run "./mdoc --help"  
then open http://localhost:8080 in browser.

## For release
```shell
go build
sudo mkdir /mnt/data/markdown
sudo cp mdoc /usr/local/bin
sudo cp -r public /mnt/data/markdown
mdoc --repo=/mnt/data/markdown daemon --listen=:8080
```

## BUG:  
User need to login again by the opaque was changed when the server has been restart, maybe use redis to fixed this problem.
