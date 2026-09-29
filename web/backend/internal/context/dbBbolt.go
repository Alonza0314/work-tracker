package context

import (
	"backend/model"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

const (
	bboltAccountBucket = "account"
)

type bboltDb struct {
	db *bbolt.DB
}

func newBboltDb(dbPath string) (*bboltDb, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create DB directory: %v", err)
	}

	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %v", err)
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bboltAccountBucket))
		return err
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create bucket %s: %v", bboltAccountBucket, err)
	}

	return &bboltDb{
		db: db,
	}, nil
}

func (b *bboltDb) Release() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

func (b *bboltDb) GetAccount(account string) (*model.Account, error) {
	var acc *model.Account
	err := b.db.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket([]byte(bboltAccountBucket)).Get([]byte(account))
		if data == nil {
			return ErrAccountNotFound
		}
		acc = &model.Account{}
		return json.Unmarshal(data, acc)
	})
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (b *bboltDb) ListAccounts() ([]*model.Account, error) {
	accounts := []*model.Account{}
	err := b.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(bboltAccountBucket)).ForEach(func(_, data []byte) error {
			acc := &model.Account{}
			if err := json.Unmarshal(data, acc); err != nil {
				return err
			}
			accounts = append(accounts, acc)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (b *bboltDb) CreateAccount(acc *model.Account) error {
	return b.putAccount(acc, false)
}

func (b *bboltDb) UpdateAccount(acc *model.Account) error {
	return b.putAccount(acc, true)
}

// putAccount stores acc; mustExist selects update (true) or create (false)
// semantics, checked inside the same transaction as the write.
func (b *bboltDb) putAccount(acc *model.Account, mustExist bool) error {
	data, err := json.Marshal(acc)
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bboltAccountBucket))
		exists := bucket.Get([]byte(acc.Account)) != nil
		switch {
		case mustExist && !exists:
			return ErrAccountNotFound
		case !mustExist && exists:
			return ErrAccountExists
		}
		return bucket.Put([]byte(acc.Account), data)
	})
}

func (b *bboltDb) DeleteAccount(account string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bboltAccountBucket))
		if bucket.Get([]byte(account)) == nil {
			return ErrAccountNotFound
		}
		return bucket.Delete([]byte(account))
	})
}
