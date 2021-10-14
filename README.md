Online server tool to made markdown document.

# Run

## No authentication
```shell
go build
./mdoc --auth-mode=false # TODO
```
then open http://localhost:8080 in browser.

## Authentication(Default mode)
```shell
go build
./mdoc --auth-mode=false --db-file=./data/mdoc.db
```

Open another console, add a user to sqlite db.
```
sudo apt-get install sqlite3
sqlite3 ./data/mdoc.db
INSERT INTO user_info(id,`passwd`,nick_name,memo)VALUES('admin','yourpasswd','admin','system init');
.q
```
then open http://localhost:8080 in browser.
