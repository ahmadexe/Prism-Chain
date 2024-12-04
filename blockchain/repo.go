package blockchain

import (
	"log"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	instance *Repository
	cachedChain *Blockchain
	once     sync.Once
)

type Repository struct {
	db *leveldb.DB
}

func InitializeBlockchainDatabase(dbPath string) {
	once.Do(func() {
		db, err := leveldb.OpenFile(dbPath, nil)
		if err != nil {
			log.Fatalf("Failed to open LevelDB: %v", err)
		}
		instance = &Repository{db: db}
	})
}

func GetDatabaseInstance() *Repository {
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

func (r *Repository) GetBlockchain() (*Blockchain, bool) {
	if cachedChain != nil {
		return cachedChain, true
	}

	chainRaw, err := r.db.Get([]byte("blockchain"), nil)
	if err != nil {
		return nil, false
	}

	chain := &Blockchain{}

	if err := chain.UnmarshalJSON(chainRaw); err != nil {
		log.Println("Failed to unmarshal blockchain:", err)
		return nil, false
	}

	return chain, true
}

func (r *Repository) SaveBlockchain(chain *Blockchain) {
	chainRaw, err := chain.MarshalJSON()
	if err != nil {
		return
	}

	if cachedChain == nil {
		cachedChain = chain
	} else {
		cachedChain.Chain = chain.Chain
		cachedChain.TransactionPool = chain.TransactionPool
		cachedChain.DataPool = chain.DataPool
	}

	r.db.Put([]byte("blockchain"), chainRaw, nil)
}
