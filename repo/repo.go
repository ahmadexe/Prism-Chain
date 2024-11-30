package repo

import (
	"log"
	"sync"

	"github.com/ahmadexe/prism_chain/blockchain"
	"github.com/syndtr/goleveldb/leveldb"
)


var (
	instance *Repository
	once     sync.Once
)

type Repository struct {
	db *leveldb.DB
}

func Initialize(dbPath string) {
	once.Do(func() {
		db, err := leveldb.OpenFile(dbPath, nil)
		if err != nil {
			log.Fatalf("Failed to open LevelDB: %v", err)
		}
		instance = &Repository{db: db}
	})
}

func GetInstance() *Repository {
	if instance == nil {
		log.Fatalf("Repository instance is not initialized. Call Initialize() first.")
	}
	return instance
}

func (r *Repository) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

func (r *Repository) GetBlockchain() (*blockchain.Blockchain, bool) {
	chainRaw, err := r.db.Get([]byte("blockchain"), nil)
	if err != nil {
		return nil, false
	}

	chain := &blockchain.Blockchain{}

	if err := chain.UnmarshalJSON(chainRaw); err != nil {
		log.Println("Failed to unmarshal blockchain:", err)
		return nil, false
	}

	return chain, true
}

func (r *Repository) SaveBlockchain(chain *blockchain.Blockchain) {
	chainRaw, err := chain.MarshalJSON()
	if err != nil {
		return
	}

	r.db.Put([]byte("blockchain"), chainRaw, nil)
}