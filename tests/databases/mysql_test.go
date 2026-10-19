package databases_test

import (
	"errors"
	"testing"

	"github.com/spotlibs/go-lib/databases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/database/orm"
)

// =============================================================================
// Stubs
// =============================================================================

// mockTransaction implements orm.Transaction (orm.Query + Commit + Rollback).
// Only the methods exercised by withTransaction are implemented.
type mockTransaction struct {
	mock.Mock
}

func (m *mockTransaction) Commit() error {
	return m.Called().Error(0)
}

func (m *mockTransaction) Rollback() error {
	m.Called()
	return nil
}

// --- orm.Query stub methods (not called by withTransaction, panic-safe stubs) ---

func (m *mockTransaction) Association(association string) orm.Association { panic("not implemented") }
func (m *mockTransaction) Begin() (orm.Transaction, error)               { panic("not implemented") }
func (m *mockTransaction) Driver() orm.Driver                            { panic("not implemented") }
func (m *mockTransaction) Count(count *int64) error                      { panic("not implemented") }
func (m *mockTransaction) Create(value any) error                        { panic("not implemented") }
func (m *mockTransaction) Cursor() (chan orm.Cursor, error)              { panic("not implemented") }
func (m *mockTransaction) Delete(value any, conds ...any) (*orm.Result, error) {
	panic("not implemented")
}
func (m *mockTransaction) Distinct(args ...any) orm.Query            { panic("not implemented") }
func (m *mockTransaction) Exec(sql string, values ...any) (*orm.Result, error) {
	panic("not implemented")
}
func (m *mockTransaction) Exists(exists *bool) error                    { panic("not implemented") }
func (m *mockTransaction) Find(dest any, conds ...any) error            { panic("not implemented") }
func (m *mockTransaction) FindOrFail(dest any, conds ...any) error      { panic("not implemented") }
func (m *mockTransaction) First(dest any) error                         { panic("not implemented") }
func (m *mockTransaction) FirstOrCreate(dest any, conds ...any) error   { panic("not implemented") }
func (m *mockTransaction) FirstOr(dest any, callback func() error) error { panic("not implemented") }
func (m *mockTransaction) FirstOrFail(dest any) error                   { panic("not implemented") }
func (m *mockTransaction) FirstOrNew(dest any, attributes any, values ...any) error {
	panic("not implemented")
}
func (m *mockTransaction) ForceDelete(value any, conds ...any) (*orm.Result, error) {
	panic("not implemented")
}
func (m *mockTransaction) Get(dest any) error                           { panic("not implemented") }
func (m *mockTransaction) Group(name string) orm.Query                  { panic("not implemented") }
func (m *mockTransaction) Having(query any, args ...any) orm.Query      { panic("not implemented") }
func (m *mockTransaction) InRandomOrder() orm.Query                     { panic("not implemented") }
func (m *mockTransaction) Join(query string, args ...any) orm.Query     { panic("not implemented") }
func (m *mockTransaction) Limit(limit int) orm.Query                    { panic("not implemented") }
func (m *mockTransaction) Load(dest any, relation string, args ...any) error {
	panic("not implemented")
}
func (m *mockTransaction) LoadMissing(dest any, relation string, args ...any) error {
	panic("not implemented")
}
func (m *mockTransaction) LockForUpdate() orm.Query                    { panic("not implemented") }
func (m *mockTransaction) Model(value any) orm.Query                   { panic("not implemented") }
func (m *mockTransaction) Offset(offset int) orm.Query                 { panic("not implemented") }
func (m *mockTransaction) Omit(columns ...string) orm.Query            { panic("not implemented") }
func (m *mockTransaction) Order(value any) orm.Query                   { panic("not implemented") }
func (m *mockTransaction) OrderBy(column string, direction ...string) orm.Query {
	panic("not implemented")
}
func (m *mockTransaction) OrderByDesc(column string) orm.Query          { panic("not implemented") }
func (m *mockTransaction) OrWhere(query any, args ...any) orm.Query     { panic("not implemented") }
func (m *mockTransaction) OrWhereIn(column string, values []any) orm.Query {
	panic("not implemented")
}
func (m *mockTransaction) OrWhereNotIn(column string, values []any) orm.Query {
	panic("not implemented")
}
func (m *mockTransaction) OrWhereBetween(column string, x, y any) orm.Query {
	panic("not implemented")
}
func (m *mockTransaction) OrWhereNotBetween(column string, x, y any) orm.Query {
	panic("not implemented")
}
func (m *mockTransaction) OrWhereNull(column string) orm.Query          { panic("not implemented") }
func (m *mockTransaction) Paginate(page, limit int, dest any, total *int64) error {
	panic("not implemented")
}
func (m *mockTransaction) Pluck(column string, dest any) error          { panic("not implemented") }
func (m *mockTransaction) Raw(sql string, values ...any) orm.Query      { panic("not implemented") }
func (m *mockTransaction) Save(value any) error                         { panic("not implemented") }
func (m *mockTransaction) SaveQuietly(value any) error                  { panic("not implemented") }
func (m *mockTransaction) Scan(dest any) error                          { panic("not implemented") }
func (m *mockTransaction) Scopes(funcs ...func(orm.Query) orm.Query) orm.Query {
	panic("not implemented")
}
func (m *mockTransaction) Select(query any, args ...any) orm.Query      { panic("not implemented") }
func (m *mockTransaction) SharedLock() orm.Query                        { panic("not implemented") }
func (m *mockTransaction) Sum(column string, dest any) error            { panic("not implemented") }
func (m *mockTransaction) Table(name string, args ...any) orm.Query     { panic("not implemented") }
func (m *mockTransaction) ToSql() orm.ToSql                             { panic("not implemented") }
func (m *mockTransaction) ToRawSql() orm.ToSql                          { panic("not implemented") }
func (m *mockTransaction) Update(column any, value ...any) (*orm.Result, error) {
	panic("not implemented")
}
func (m *mockTransaction) UpdateOrCreate(dest any, attributes any, values any) error {
	panic("not implemented")
}
func (m *mockTransaction) Where(query any, args ...any) orm.Query       { panic("not implemented") }
func (m *mockTransaction) WhereIn(column string, values []any) orm.Query { panic("not implemented") }
func (m *mockTransaction) WhereNotIn(column string, values []any) orm.Query {
	panic("not implemented")
}
func (m *mockTransaction) WhereBetween(column string, x, y any) orm.Query {
	panic("not implemented")
}
func (m *mockTransaction) WhereNotBetween(column string, x, y any) orm.Query {
	panic("not implemented")
}
func (m *mockTransaction) WhereNull(column string) orm.Query            { panic("not implemented") }
func (m *mockTransaction) WhereNotNull(column string) orm.Query         { panic("not implemented") }
func (m *mockTransaction) WithoutEvents() orm.Query                     { panic("not implemented") }
func (m *mockTransaction) WithTrashed() orm.Query                       { panic("not implemented") }
func (m *mockTransaction) With(query string, args ...any) orm.Query     { panic("not implemented") }

// mockQuery implements orm.Query — only Begin() is called by withTransaction.
type mockQuery struct {
	mock.Mock
}

func (m *mockQuery) Begin() (orm.Transaction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(orm.Transaction), args.Error(1)
}

func (m *mockQuery) Association(association string) orm.Association { panic("not implemented") }
func (m *mockQuery) Driver() orm.Driver                            { panic("not implemented") }
func (m *mockQuery) Count(count *int64) error                      { panic("not implemented") }
func (m *mockQuery) Create(value any) error                        { panic("not implemented") }
func (m *mockQuery) Cursor() (chan orm.Cursor, error)              { panic("not implemented") }
func (m *mockQuery) Delete(value any, conds ...any) (*orm.Result, error) {
	panic("not implemented")
}
func (m *mockQuery) Distinct(args ...any) orm.Query            { panic("not implemented") }
func (m *mockQuery) Exec(sql string, values ...any) (*orm.Result, error) {
	panic("not implemented")
}
func (m *mockQuery) Exists(exists *bool) error                    { panic("not implemented") }
func (m *mockQuery) Find(dest any, conds ...any) error            { panic("not implemented") }
func (m *mockQuery) FindOrFail(dest any, conds ...any) error      { panic("not implemented") }
func (m *mockQuery) First(dest any) error                         { panic("not implemented") }
func (m *mockQuery) FirstOrCreate(dest any, conds ...any) error   { panic("not implemented") }
func (m *mockQuery) FirstOr(dest any, callback func() error) error { panic("not implemented") }
func (m *mockQuery) FirstOrFail(dest any) error                   { panic("not implemented") }
func (m *mockQuery) FirstOrNew(dest any, attributes any, values ...any) error {
	panic("not implemented")
}
func (m *mockQuery) ForceDelete(value any, conds ...any) (*orm.Result, error) {
	panic("not implemented")
}
func (m *mockQuery) Get(dest any) error                           { panic("not implemented") }
func (m *mockQuery) Group(name string) orm.Query                  { panic("not implemented") }
func (m *mockQuery) Having(query any, args ...any) orm.Query      { panic("not implemented") }
func (m *mockQuery) InRandomOrder() orm.Query                     { panic("not implemented") }
func (m *mockQuery) Join(query string, args ...any) orm.Query     { panic("not implemented") }
func (m *mockQuery) Limit(limit int) orm.Query                    { panic("not implemented") }
func (m *mockQuery) Load(dest any, relation string, args ...any) error {
	panic("not implemented")
}
func (m *mockQuery) LoadMissing(dest any, relation string, args ...any) error {
	panic("not implemented")
}
func (m *mockQuery) LockForUpdate() orm.Query                    { panic("not implemented") }
func (m *mockQuery) Model(value any) orm.Query                   { panic("not implemented") }
func (m *mockQuery) Offset(offset int) orm.Query                 { panic("not implemented") }
func (m *mockQuery) Omit(columns ...string) orm.Query            { panic("not implemented") }
func (m *mockQuery) Order(value any) orm.Query                   { panic("not implemented") }
func (m *mockQuery) OrderBy(column string, direction ...string) orm.Query {
	panic("not implemented")
}
func (m *mockQuery) OrderByDesc(column string) orm.Query          { panic("not implemented") }
func (m *mockQuery) OrWhere(query any, args ...any) orm.Query     { panic("not implemented") }
func (m *mockQuery) OrWhereIn(column string, values []any) orm.Query {
	panic("not implemented")
}
func (m *mockQuery) OrWhereNotIn(column string, values []any) orm.Query {
	panic("not implemented")
}
func (m *mockQuery) OrWhereBetween(column string, x, y any) orm.Query {
	panic("not implemented")
}
func (m *mockQuery) OrWhereNotBetween(column string, x, y any) orm.Query {
	panic("not implemented")
}
func (m *mockQuery) OrWhereNull(column string) orm.Query          { panic("not implemented") }
func (m *mockQuery) Paginate(page, limit int, dest any, total *int64) error {
	panic("not implemented")
}
func (m *mockQuery) Pluck(column string, dest any) error          { panic("not implemented") }
func (m *mockQuery) Raw(sql string, values ...any) orm.Query      { panic("not implemented") }
func (m *mockQuery) Save(value any) error                         { panic("not implemented") }
func (m *mockQuery) SaveQuietly(value any) error                  { panic("not implemented") }
func (m *mockQuery) Scan(dest any) error                          { panic("not implemented") }
func (m *mockQuery) Scopes(funcs ...func(orm.Query) orm.Query) orm.Query {
	panic("not implemented")
}
func (m *mockQuery) Select(query any, args ...any) orm.Query      { panic("not implemented") }
func (m *mockQuery) SharedLock() orm.Query                        { panic("not implemented") }
func (m *mockQuery) Sum(column string, dest any) error            { panic("not implemented") }
func (m *mockQuery) Table(name string, args ...any) orm.Query     { panic("not implemented") }
func (m *mockQuery) ToSql() orm.ToSql                             { panic("not implemented") }
func (m *mockQuery) ToRawSql() orm.ToSql                          { panic("not implemented") }
func (m *mockQuery) Update(column any, value ...any) (*orm.Result, error) {
	panic("not implemented")
}
func (m *mockQuery) UpdateOrCreate(dest any, attributes any, values any) error {
	panic("not implemented")
}
func (m *mockQuery) Where(query any, args ...any) orm.Query       { panic("not implemented") }
func (m *mockQuery) WhereIn(column string, values []any) orm.Query { panic("not implemented") }
func (m *mockQuery) WhereNotIn(column string, values []any) orm.Query {
	panic("not implemented")
}
func (m *mockQuery) WhereBetween(column string, x, y any) orm.Query {
	panic("not implemented")
}
func (m *mockQuery) WhereNotBetween(column string, x, y any) orm.Query {
	panic("not implemented")
}
func (m *mockQuery) WhereNull(column string) orm.Query            { panic("not implemented") }
func (m *mockQuery) WhereNotNull(column string) orm.Query         { panic("not implemented") }
func (m *mockQuery) WithoutEvents() orm.Query                     { panic("not implemented") }
func (m *mockQuery) WithTrashed() orm.Query                       { panic("not implemented") }
func (m *mockQuery) With(query string, args ...any) orm.Query     { panic("not implemented") }

// =============================================================================
// WithTransaction
// =============================================================================

func TestWithTransactionSuccess(t *testing.T) {
	tx := new(mockTransaction)
	tx.On("Commit").Return(nil)

	q := new(mockQuery)
	q.On("Begin").Return(tx, nil)

	err := databases.WithTransactionQuery(q, func(tx orm.Query) error {
		return nil
	})

	assert.NoError(t, err)
	tx.AssertExpectations(t)
	q.AssertExpectations(t)
}

func TestWithTransactionBeginFails(t *testing.T) {
	q := new(mockQuery)
	q.On("Begin").Return(nil, errors.New("connection refused"))

	err := databases.WithTransactionQuery(q, func(tx orm.Query) error {
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
	q.AssertExpectations(t)
}

func TestWithTransactionFnFails(t *testing.T) {
	tx := new(mockTransaction)
	tx.On("Rollback").Return(nil)

	q := new(mockQuery)
	q.On("Begin").Return(tx, nil)

	err := databases.WithTransactionQuery(q, func(tx orm.Query) error {
		return errors.New("business error")
	})

	assert.Error(t, err)
	assert.EqualError(t, err, "business error")
	tx.AssertExpectations(t)
}

func TestWithTransactionCommitFails(t *testing.T) {
	tx := new(mockTransaction)
	tx.On("Commit").Return(errors.New("commit failed"))
	tx.On("Rollback").Return(nil)

	q := new(mockQuery)
	q.On("Begin").Return(tx, nil)

	err := databases.WithTransactionQuery(q, func(tx orm.Query) error {
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit")
	tx.AssertExpectations(t)
}
