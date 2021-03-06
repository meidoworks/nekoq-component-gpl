package lmdb

import (
	"goimport.moetang.info/nekoq-api/component/db"
	"goimport.moetang.info/nekoq-api/component/db/manager"
	"goimport.moetang.info/nekoq-api/errorutil"

	"github.com/szferi/gomdb"
)

const (
	simpleDbNameString = "db:__simple_db__"
	atomicDbNameString = "db:__atomic_db__"
)

type lmdbDbApiImpl struct {
	env         *mdb.Env
	simpleDbDbi mdb.DBI
	atomicDbDbi mdb.DBI
}

var _ manager.DbApi = new(lmdbDbApiImpl)

func createDbApi(config map[string]string) (manager.DbApi, error) {
	dbDir, ok := config[CONFIG_DATABASE_DIR_PATH]
	if !ok {
		return nil, errorutil.New("no database dir -> lmdb -> db -> nekoq-component")
	}

	lmdbImpl := new(lmdbDbApiImpl)

	env, err := mdb.NewEnv()
	if err != nil {
		return nil, errorutil.NewNested("new env error -> lmdb -> db -> nekoq-component", err)
	}

	lmdbImpl.env = env
	env.SetMapSize(1 << 40) //1TB
	env.SetMaxDBs(16)       // 16 dbs

	err = env.Open(dbDir, 0, 0644)
	if err != nil {
		return nil, errorutil.NewNested("open env error -> lmdb -> db -> nekoq-component", err)
	}

	//========
	txn, err := env.BeginTxn(nil, 0)
	if err != nil {
		env.Close()
		return nil, errorutil.NewNested("initing error: begin txn for simpledb -> lmdb -> db -> nekoq-component", err)
	}
	var nameStr = simpleDbNameString
	dbi, err := txn.DBIOpen(&nameStr, mdb.CREATE)
	if err != nil {
		env.Close()
		return nil, errorutil.NewNested("initing error: dbiOpen for simpledb -> lmdb -> db -> nekoq-component", err)
	}
	txn.Commit()
	lmdbImpl.simpleDbDbi = dbi
	//========
	txn, err = env.BeginTxn(nil, 0)
	if err != nil {
		env.Close()
		return nil, errorutil.NewNested("initing error: begin txn for atomicdb -> lmdb -> db -> nekoq-component", err)
	}
	nameStr = atomicDbNameString
	dbi, err = txn.DBIOpen(&nameStr, mdb.CREATE)
	if err != nil {
		env.Close()
		return nil, errorutil.NewNested("initing error: dbiOpen for atomicdb -> lmdb -> db -> nekoq-component", err)
	}
	txn.Commit()
	lmdbImpl.atomicDbDbi = dbi
	//========

	return lmdbImpl, nil
}

func (this *lmdbDbApiImpl) GetSimpleDb() (db.SimpleDB, error) {
	return createSimpleDb(this)
}

func (this *lmdbDbApiImpl) CloseDbApi() error {
	return this.env.Close()
}

func (this *lmdbDbApiImpl) GetAtomicDb() (db.AtomicDB, error) {
	return createAtomicDb(this)
}
