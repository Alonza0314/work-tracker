package context

import (
	"backend/constant"
	"backend/model"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/bbolt"
	bboltErrors "go.etcd.io/bbolt/errors"
)

const (
	bboltAccountBucket    = "account"
	bboltCategoryBucket   = "category"
	bboltProjectBucket    = "project"
	bboltWorkRecordBucket = "work"
	bboltTodoBucket       = "todo"
	bboltSettingBucket    = "setting"
	bboltHolidayBucket    = "holiday"

	bboltWorkSettingKey    = "work"
	bboltHolidaySyncKey    = "holidaySync"
	bboltHolidayKeyDivider = "#"
)

var bboltBuckets = []string{
	bboltAccountBucket,
	bboltCategoryBucket,
	bboltProjectBucket,
	bboltWorkRecordBucket,
	bboltTodoBucket,
	bboltSettingBucket,
	bboltHolidayBucket,
}

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
		for _, bucket := range bboltBuckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return fmt.Errorf("failed to create bucket %s: %v", bucket, err)
			}
		}
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, err
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

// generic helpers: every bucket stores one JSON value per key

func bboltGet[T any](db *bbolt.DB, bucket, key string, notFound error) (*T, error) {
	var value *T
	err := db.View(func(tx *bbolt.Tx) error {
		data := tx.Bucket([]byte(bucket)).Get([]byte(key))
		if data == nil {
			return notFound
		}
		value = new(T)
		return json.Unmarshal(data, value)
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

// bboltList returns the values in key order that satisfy keep (nil keeps all).
func bboltList[T any](db *bbolt.DB, bucket string, keep func(*T) bool) ([]*T, error) {
	var values []*T
	err := db.View(func(tx *bbolt.Tx) error {
		var err error
		values, err = bboltListTx(tx, bucket, keep)
		return err
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

func bboltListTx[T any](tx *bbolt.Tx, bucket string, keep func(*T) bool) ([]*T, error) {
	values := []*T{}
	err := tx.Bucket([]byte(bucket)).ForEach(func(_, data []byte) error {
		value := new(T)
		if err := json.Unmarshal(data, value); err != nil {
			return err
		}
		if keep == nil || keep(value) {
			values = append(values, value)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

// bboltPut stores value under key. With mustExist it fails with notFound when
// the key is missing; without it, it fails with exists (if non-nil) when the
// key is already there.
func bboltPut(tx *bbolt.Tx, bucket, key string, value any, mustExist bool, notFound, exists error) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	b := tx.Bucket([]byte(bucket))
	found := b.Get([]byte(key)) != nil
	switch {
	case mustExist && !found:
		return notFound
	case !mustExist && found && exists != nil:
		return exists
	}
	return b.Put([]byte(key), data)
}

func bboltDelete(tx *bbolt.Tx, bucket, key string, notFound error) error {
	b := tx.Bucket([]byte(bucket))
	if b.Get([]byte(key)) == nil {
		return notFound
	}
	return b.Delete([]byte(key))
}

// bboltUpdateAll rewrites every value in bucket for which change returns true
// and returns how many were rewritten. Values are collected first because a
// bucket must not be modified during ForEach.
func bboltUpdateAll[T any](tx *bbolt.Tx, bucket string, change func(*T) bool) (int, error) {
	b := tx.Bucket([]byte(bucket))
	changed := map[string]*T{}
	if err := b.ForEach(func(key, data []byte) error {
		value := new(T)
		if err := json.Unmarshal(data, value); err != nil {
			return err
		}
		if change(value) {
			changed[string(key)] = value
		}
		return nil
	}); err != nil {
		return 0, err
	}

	for key, value := range changed {
		if err := bboltPut(tx, bucket, key, value, true, nil, nil); err != nil {
			return 0, err
		}
	}
	return len(changed), nil
}

// bboltNextID returns a zero-padded sequence number, so IDs sort in creation
// order both as bytes and as strings.
func bboltNextID(tx *bbolt.Tx, bucket string) (string, error) {
	seq, err := tx.Bucket([]byte(bucket)).NextSequence()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%020d", seq), nil
}

// account

func (b *bboltDb) GetAccount(account string) (*model.Account, error) {
	return bboltGet[model.Account](b.db, bboltAccountBucket, account, ErrAccountNotFound)
}

func (b *bboltDb) ListAccounts() ([]*model.Account, error) {
	return bboltList[model.Account](b.db, bboltAccountBucket, nil)
}

func (b *bboltDb) CreateAccount(acc *model.Account) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltPut(tx, bboltAccountBucket, acc.Account, acc, false, ErrAccountNotFound, ErrAccountExists)
	})
}

func (b *bboltDb) UpdateAccount(acc *model.Account) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltPut(tx, bboltAccountBucket, acc.Account, acc, true, ErrAccountNotFound, ErrAccountExists)
	})
}

func (b *bboltDb) DeleteAccount(account string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltDelete(tx, bboltAccountBucket, account, ErrAccountNotFound)
	})
}

func (b *bboltDb) RenameAccount(oldAccount, newAccount string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		data := tx.Bucket([]byte(bboltAccountBucket)).Get([]byte(oldAccount))
		if data == nil {
			return ErrAccountNotFound
		}
		acc := &model.Account{}
		if err := json.Unmarshal(data, acc); err != nil {
			return err
		}

		acc.Account = newAccount
		if err := bboltPut(tx, bboltAccountBucket, newAccount, acc, false, nil, ErrAccountExists); err != nil {
			return err
		}
		if err := bboltDelete(tx, bboltAccountBucket, oldAccount, ErrAccountNotFound); err != nil {
			return err
		}

		if _, err := bboltUpdateAll(tx, bboltWorkRecordBucket, func(r *model.WorkRecord) bool {
			if r.Account != oldAccount {
				return false
			}
			r.Account = newAccount
			return true
		}); err != nil {
			return err
		}
		_, err := bboltUpdateAll(tx, bboltTodoBucket, func(t *model.Todo) bool {
			if t.Account != oldAccount {
				return false
			}
			t.Account = newAccount
			return true
		})
		return err
	})
}

// category & project

func (b *bboltDb) createOption(bucket string, option *model.WorkOption) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		id, err := bboltNextID(tx, bucket)
		if err != nil {
			return err
		}
		option.ID = id
		return bboltPut(tx, bucket, option.ID, option, false, nil, nil)
	})
}

func (b *bboltDb) updateOption(bucket string, option *model.WorkOption, notFound error) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltPut(tx, bucket, option.ID, option, true, notFound, nil)
	})
}

// deleteOption deletes the option and clears it, via field, from every work
// record and todo.
func (b *bboltDb) deleteOption(bucket, id string, notFound error, field func(*model.WorkEntry) *string) (int, error) {
	cleared := 0
	err := b.db.Update(func(tx *bbolt.Tx) error {
		if err := bboltDelete(tx, bucket, id, notFound); err != nil {
			return err
		}

		clearIn := func(entry *model.WorkEntry) bool {
			if *field(entry) != id {
				return false
			}
			*field(entry) = ""
			return true
		}
		records, err := bboltUpdateAll(tx, bboltWorkRecordBucket, func(r *model.WorkRecord) bool {
			return clearIn(&r.WorkEntry)
		})
		if err != nil {
			return err
		}
		todos, err := bboltUpdateAll(tx, bboltTodoBucket, func(t *model.Todo) bool {
			return clearIn(&t.WorkEntry)
		})
		if err != nil {
			return err
		}

		cleared = records + todos
		return nil
	})
	if err != nil {
		return 0, err
	}
	return cleared, nil
}

func categoryField(entry *model.WorkEntry) *string {
	return &entry.CategoryID
}

func projectField(entry *model.WorkEntry) *string {
	return &entry.ProjectID
}

func (b *bboltDb) GetCategory(id string) (*model.WorkOption, error) {
	return bboltGet[model.WorkOption](b.db, bboltCategoryBucket, id, ErrCategoryNotFound)
}

func (b *bboltDb) ListCategories() ([]*model.WorkOption, error) {
	return bboltList[model.WorkOption](b.db, bboltCategoryBucket, nil)
}

func (b *bboltDb) CreateCategory(category *model.WorkOption) error {
	return b.createOption(bboltCategoryBucket, category)
}

func (b *bboltDb) UpdateCategory(category *model.WorkOption) error {
	return b.updateOption(bboltCategoryBucket, category, ErrCategoryNotFound)
}

func (b *bboltDb) DeleteCategory(id string) (int, error) {
	return b.deleteOption(bboltCategoryBucket, id, ErrCategoryNotFound, categoryField)
}

func (b *bboltDb) GetProject(id string) (*model.WorkOption, error) {
	return bboltGet[model.WorkOption](b.db, bboltProjectBucket, id, ErrProjectNotFound)
}

func (b *bboltDb) ListProjects() ([]*model.WorkOption, error) {
	return bboltList[model.WorkOption](b.db, bboltProjectBucket, nil)
}

func (b *bboltDb) CreateProject(project *model.WorkOption) error {
	return b.createOption(bboltProjectBucket, project)
}

func (b *bboltDb) UpdateProject(project *model.WorkOption) error {
	return b.updateOption(bboltProjectBucket, project, ErrProjectNotFound)
}

func (b *bboltDb) DeleteProject(id string) (int, error) {
	return b.deleteOption(bboltProjectBucket, id, ErrProjectNotFound, projectField)
}

// work record

func (b *bboltDb) GetWorkRecord(id string) (*model.WorkRecord, error) {
	return bboltGet[model.WorkRecord](b.db, bboltWorkRecordBucket, id, ErrWorkRecordNotFound)
}

func (b *bboltDb) ListWorkRecords(filter *model.WorkRecordFilter) ([]*model.WorkRecord, error) {
	return bboltList(b.db, bboltWorkRecordBucket, func(r *model.WorkRecord) bool {
		return (filter.Account == "" || r.Account == filter.Account) &&
			(filter.From == "" || r.Date >= filter.From) &&
			(filter.To == "" || r.Date <= filter.To) &&
			(filter.CategoryID == "" || r.CategoryID == filter.CategoryID) &&
			(filter.ProjectID == "" || r.ProjectID == filter.ProjectID)
	})
}

func createWorkRecord(tx *bbolt.Tx, record *model.WorkRecord) error {
	id, err := bboltNextID(tx, bboltWorkRecordBucket)
	if err != nil {
		return err
	}
	record.ID = id
	return bboltPut(tx, bboltWorkRecordBucket, record.ID, record, false, nil, nil)
}

func (b *bboltDb) CreateWorkRecord(record *model.WorkRecord) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return createWorkRecord(tx, record)
	})
}

func (b *bboltDb) UpdateWorkRecord(record *model.WorkRecord) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltPut(tx, bboltWorkRecordBucket, record.ID, record, true, ErrWorkRecordNotFound, nil)
	})
}

func (b *bboltDb) DeleteWorkRecord(id string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltDelete(tx, bboltWorkRecordBucket, id, ErrWorkRecordNotFound)
	})
}

// todo

func (b *bboltDb) GetTodo(id string) (*model.Todo, error) {
	return bboltGet[model.Todo](b.db, bboltTodoBucket, id, ErrTodoNotFound)
}

func (b *bboltDb) ListTodos(account string) ([]*model.Todo, error) {
	return bboltList(b.db, bboltTodoBucket, func(t *model.Todo) bool {
		return t.Account == account
	})
}

func (b *bboltDb) CreateTodo(todo *model.Todo) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		id, err := bboltNextID(tx, bboltTodoBucket)
		if err != nil {
			return err
		}
		todo.ID = id
		return bboltPut(tx, bboltTodoBucket, todo.ID, todo, false, nil, nil)
	})
}

func (b *bboltDb) UpdateTodo(todo *model.Todo) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltPut(tx, bboltTodoBucket, todo.ID, todo, true, ErrTodoNotFound, nil)
	})
}

func (b *bboltDb) DeleteTodo(id string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltDelete(tx, bboltTodoBucket, id, ErrTodoNotFound)
	})
}

func (b *bboltDb) CompleteTodo(todoID string, record *model.WorkRecord) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		if err := bboltDelete(tx, bboltTodoBucket, todoID, ErrTodoNotFound); err != nil {
			return err
		}
		return createWorkRecord(tx, record)
	})
}

// setting

func (b *bboltDb) GetWorkSetting() (*model.WorkSetting, error) {
	setting, err := bboltGet[model.WorkSetting](b.db, bboltSettingBucket, bboltWorkSettingKey, errSettingNotFound)
	if err == errSettingNotFound {
		return &model.WorkSetting{}, nil
	}
	return setting, err
}

func (b *bboltDb) UpdateWorkSetting(setting *model.WorkSetting) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltPut(tx, bboltSettingBucket, bboltWorkSettingKey, setting, false, nil, nil)
	})
}

// holiday: keys are "<date>#<source>", so they sort by date then source

func bboltHolidayKey(date, source string) string {
	return date + bboltHolidayKeyDivider + source
}

func (b *bboltDb) ListHolidays(from, to string) ([]*model.Holiday, error) {
	return bboltList(b.db, bboltHolidayBucket, func(h *model.Holiday) bool {
		return h.Date >= from && h.Date <= to
	})
}

func (b *bboltDb) SaveHoliday(holiday *model.Holiday) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltPut(tx, bboltHolidayBucket, bboltHolidayKey(holiday.Date, holiday.Source), holiday, false, nil, nil)
	})
}

func (b *bboltDb) DeleteHoliday(date, source string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltDelete(tx, bboltHolidayBucket, bboltHolidayKey(date, source), ErrHolidayNotFound)
	})
}

func (b *bboltDb) ReplaceGovHolidays(year string, holidays []*model.Holiday) (int, error) {
	err := b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bboltHolidayBucket))

		var stale [][]byte
		prefix := []byte(year + "-")
		suffix := bboltHolidayKeyDivider + constant.HOLIDAY_SOURCE_GOV
		cursor := bucket.Cursor()
		for key, _ := cursor.Seek(prefix); key != nil && bytes.HasPrefix(key, prefix); key, _ = cursor.Next() {
			if strings.HasSuffix(string(key), suffix) {
				stale = append(stale, append([]byte(nil), key...))
			}
		}
		for _, key := range stale {
			if err := bucket.Delete(key); err != nil {
				return err
			}
		}

		for _, holiday := range holidays {
			if err := bboltPut(tx, bboltHolidayBucket, bboltHolidayKey(holiday.Date, holiday.Source), holiday, false, nil, nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(holidays), nil
}

type bboltHolidaySync struct {
	LastSyncedAt time.Time `json:"lastSyncedAt"`
}

func (b *bboltDb) GetHolidaySyncedAt() (time.Time, error) {
	sync, err := bboltGet[bboltHolidaySync](b.db, bboltSettingBucket, bboltHolidaySyncKey, errSettingNotFound)
	if errors.Is(err, errSettingNotFound) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return sync.LastSyncedAt, nil
}

func (b *bboltDb) SetHolidaySyncedAt(at time.Time) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return bboltPut(tx, bboltSettingBucket, bboltHolidaySyncKey, &bboltHolidaySync{LastSyncedAt: at}, false, nil, nil)
	})
}

// backup

func (b *bboltDb) Dump() (*model.Backup, error) {
	backup := &model.Backup{}
	err := b.db.View(func(tx *bbolt.Tx) error {
		var err error
		if backup.Accounts, err = dumpBucket[model.Account](tx, bboltAccountBucket); err != nil {
			return err
		}
		if backup.Categories, err = dumpBucket[model.WorkOption](tx, bboltCategoryBucket); err != nil {
			return err
		}
		if backup.Projects, err = dumpBucket[model.WorkOption](tx, bboltProjectBucket); err != nil {
			return err
		}
		if backup.WorkRecords, err = dumpBucket[model.WorkRecord](tx, bboltWorkRecordBucket); err != nil {
			return err
		}
		if backup.Todos, err = dumpBucket[model.Todo](tx, bboltTodoBucket); err != nil {
			return err
		}
		if backup.Holidays, err = dumpBucket[model.Holiday](tx, bboltHolidayBucket); err != nil {
			return err
		}

		settings := tx.Bucket([]byte(bboltSettingBucket))
		if data := settings.Get([]byte(bboltWorkSettingKey)); data != nil {
			if err := json.Unmarshal(data, &backup.WorkSetting); err != nil {
				return err
			}
		}
		if data := settings.Get([]byte(bboltHolidaySyncKey)); data != nil {
			sync := bboltHolidaySync{}
			if err := json.Unmarshal(data, &sync); err != nil {
				return err
			}
			backup.HolidaySyncedAt = sync.LastSyncedAt
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return backup, nil
}

func dumpBucket[T any](tx *bbolt.Tx, bucket string) ([]T, error) {
	values, err := bboltListTx[T](tx, bucket, nil)
	if err != nil {
		return nil, err
	}
	result := make([]T, 0, len(values))
	for _, value := range values {
		result = append(result, *value)
	}
	return result, nil
}

func (b *bboltDb) Restore(backup *model.Backup) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		if err := bboltRecreateBuckets(tx); err != nil {
			return err
		}

		for i := range backup.Accounts {
			if err := bboltPut(tx, bboltAccountBucket, backup.Accounts[i].Account, &backup.Accounts[i], false, nil, ErrAccountExists); err != nil {
				return err
			}
		}
		for _, h := range backup.Holidays {
			if err := bboltPut(tx, bboltHolidayBucket, bboltHolidayKey(h.Date, h.Source), h, false, nil, nil); err != nil {
				return err
			}
		}

		if err := restoreByID(tx, bboltCategoryBucket, backup.Categories, func(o model.WorkOption) string { return o.ID }); err != nil {
			return err
		}
		if err := restoreByID(tx, bboltProjectBucket, backup.Projects, func(o model.WorkOption) string { return o.ID }); err != nil {
			return err
		}
		if err := restoreByID(tx, bboltWorkRecordBucket, backup.WorkRecords, func(r model.WorkRecord) string { return r.ID }); err != nil {
			return err
		}
		if err := restoreByID(tx, bboltTodoBucket, backup.Todos, func(t model.Todo) string { return t.ID }); err != nil {
			return err
		}

		if err := bboltPut(tx, bboltSettingBucket, bboltWorkSettingKey, backup.WorkSetting, false, nil, nil); err != nil {
			return err
		}
		if !backup.HolidaySyncedAt.IsZero() {
			return bboltPut(tx, bboltSettingBucket, bboltHolidaySyncKey, &bboltHolidaySync{LastSyncedAt: backup.HolidaySyncedAt}, false, nil, nil)
		}
		return nil
	})
}

// restoreByID stores values keyed by their sequence ID and moves the bucket's
// sequence past the largest one, so new IDs continue after them.
func restoreByID[T any](tx *bbolt.Tx, bucket string, values []T, id func(T) string) error {
	var last uint64
	for _, value := range values {
		key := id(value)
		if err := bboltPut(tx, bucket, key, value, false, nil, nil); err != nil {
			return err
		}
		if seq, err := strconv.ParseUint(key, 10, 64); err == nil && seq > last {
			last = seq
		}
	}
	return tx.Bucket([]byte(bucket)).SetSequence(last)
}

func (b *bboltDb) Reset() error {
	return b.db.Update(bboltRecreateBuckets)
}

// bboltRecreateBuckets empties every bucket and restarts its ID sequence.
func bboltRecreateBuckets(tx *bbolt.Tx) error {
	for _, bucket := range bboltBuckets {
		if err := tx.DeleteBucket([]byte(bucket)); err != nil && !errors.Is(err, bboltErrors.ErrBucketNotFound) {
			return fmt.Errorf("failed to delete bucket %s: %v", bucket, err)
		}
		if _, err := tx.CreateBucket([]byte(bucket)); err != nil {
			return fmt.Errorf("failed to create bucket %s: %v", bucket, err)
		}
	}
	return nil
}
