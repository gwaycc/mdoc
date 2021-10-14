package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gwaylib/database"
	"github.com/gwaylib/log"
)

func init() {
	database.REFLECT_DRV_NAME = database.DRV_NAME_SQLITE3
}

var (
	mdb        *database.DB
	mdblk      sync.Mutex
	dbGlobalLk sync.Mutex
)

func InitDB(dbFile string) {
	mdblk.Lock()
	defer mdblk.Unlock()
	if mdb != nil {
		return
	}

	dir := filepath.Dir(dbFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	log.Infof("Init sqlite db at:%s", dbFile)
	db, err := database.Open(database.DRV_NAME_SQLITE3, fmt.Sprintf("file:%s?_loc=auto&_mode=rwc&_journal_mode=WAL&cache=shared&encoding=UTF-8&_timeout=10000", dbFile))
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1) // more than one should happen:database is locked

	mdb = db

	// Init tables
	if _, err := db.Exec(tb_user_sql); err != nil {
		panic(err)
	}
}

func HasDB() bool {
	mdblk.Lock()
	defer mdblk.Unlock()
	return mdb != nil
}

func GetDB() *database.DB {
	mdblk.Lock()
	defer mdblk.Unlock()
	if mdb == nil {
		panic("Need InitDB at first")
	}
	return mdb
}
