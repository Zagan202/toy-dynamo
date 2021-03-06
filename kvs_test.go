package main

import (
	"strings"
	"sync"
	"testing"
	"time"
)

type testEntry struct {
	time    time.Time
	clock   map[string]int
	value   string
	version int
}

func (e *testEntry) SetVersion(v int) {
	e.version = v
}

func (e *testEntry) GetVersion() int {
	return e.version
}

func (e *testEntry) GetValue() string {
	return e.value
}

func (e *testEntry) GetTimestamp() time.Time {
	return e.time
}

func (e *testEntry) GetClock() map[string]int {
	return e.clock
}

func (e *testEntry) Update(key string, time time.Time, clock map[string]int, val string) {
	// goes nowhere does nothing
}

func (e *testEntry) Alive() bool {
	// TODO: figure this out
	return true
}

func (e *testEntry) Delete(key string, time time.Time, payload map[string]int) {
	// goes nowhere does nothing
}

// This tests for a key that does not exist in the db, the KVS should return version -1 and alive == false
func TestKVSContainsCheckIfDoesntExist(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}

	alive, version := k.Contains(keyNotHere)

	assert(t, version == 0, "Key found that does not exist.")
	assert(t, !alive, "Contains returned alive for nonexistent key.")
}

// If the key does exist, the KVS should return alive == true and its version
func TestKVSContainsCheckIfDoesExist(t *testing.T) {
	entryExists := testEntry{
		time:    time.Now(),
		clock:   nil,
		value:   valExists,
		version: 1,
	}
	db := map[string]KeyEntry{
		keyExists: &entryExists,
	}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	alive, version := k.Contains(keyExists)
	assert(t, alive, "Contains() returned !alive for existing key.")
	assert(t, version == 1, "Contains() returned incorrect version.")

}

// Get() on an existing key/value pair should return the value for the key
func TestKVSGetExistingVal(t *testing.T) {
	entryExists := testEntry{
		time:    time.Now(),
		clock:   nil,
		value:   valExists,
		version: 1,
	}
	db := map[string]KeyEntry{
		keyExists: &entryExists,
	}

	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	returned, _ := k.Get(keyExists, nil)
	equals(t, valExists, returned)
}

// Get on a non-existing key should return an empty string
func TestKVSGetValDoesntExist(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	returned, _ := k.Get(keyone, nil)
	equals(t, "", returned)
}

// Delete on an existing key/value pair should return true and further Gets() should fail
func TestKVSDeleteExistingKeyValPair(t *testing.T) {
	entryExists := testEntry{
		time:    time.Now(),
		clock:   nil,
		value:   valExists,
		version: 1,
	}
	db := map[string]KeyEntry{
		keyExists: &entryExists,
	}

	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, k.Delete(keyExists, time.Now(), map[string]int{}), "Did not delete Key Val Pair")
}

// Delete on a key that doesn't exist should return false
func TestKVSDeleteKeyDoesntExist(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, !k.Delete(keyNotHere, time.Now(), map[string]int{}), "Deleted a keyvalue pair not in data store prior")
}

// Put() with a new key should return true
func TestKVSPutNewKeyNewVal(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, k.Put(keyone, valone, time.Now(), nil), "New key and value were not added")
}

// Overwriting a value should return true
func TestKVSPutExistKeyOverwriteVal(t *testing.T) {
	entryExists := testEntry{
		time:    time.Now(),
		clock:   nil,
		value:   valExists,
		version: 1,
	}
	db := map[string]KeyEntry{
		keyExists: &entryExists,
	}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, k.Put(keyone, valtwo, time.Now(), nil), "Did not overwrite existing key's value")

}

// Put() with an invalid key should return failure
func TestKVSPutInvalidkey(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, !k.Put(invalidKey, valtwo, time.Now(), nil), "Invalid key added")
}

// Put() with an invalid value should fail
func TestKVSPutInvalidVal(t *testing.T) {
	k := NewKVS()
	var b strings.Builder
	b.Grow(1048577)
	for i := 0; i < 1048577; i++ {
		b.WriteByte(0)
	}
	invalidVal := b.String()
	assert(t, !k.Put(keyone, invalidVal, time.Now(), nil), "Invalid value added")
}

// GetVersion of an existing entry should return the version
func TestEntryGetVersionKeyExists(t *testing.T) {
	e := Entry{
		Version:   1,
		Timestamp: time.Now(),
		Clock: map[string]int{
			keyExists: 1,
		},
		Value:     valExists,
		Tombstone: false,
	}

	assert(t, e.GetVersion() == 1, "GetVersion() returned wrong value")
}

// GetVersion of a non-existing key should return 0
func TestEntryGetVersionKeyNotExists(t *testing.T) {
	// Setting e to be an uninitialized pointer means the struct is nil
	var e *Entry
	assert(t, e.GetVersion() == 0, "GetVersion() returned wrong value")
}

// GetTimestamp should return the timestamp of the key if it exists
func TestEntryGetTimestampKeyExists(t *testing.T) {
	timestamp := time.Now()
	e := Entry{
		Version:   1,
		Timestamp: timestamp,
		Clock: map[string]int{
			keyExists: 1,
		},
		Value:     valExists,
		Tombstone: false,
	}
	assert(t, e.GetTimestamp() == timestamp, "GetTimestamp() returned wrong value")
}

// GetTimestamp should return the zero-value timestamp if the key doesn't exist
func TestEntryGetTimestampKeyNotExists(t *testing.T) {
	var e *Entry
	assert(t, e.GetTimestamp() == time.Time{}, "GetTimestamp() returned wrong value")
}

// GetClock should return the key's clock map if it exists
func TestEntryGetClockKeyExists(t *testing.T) {
	expectedClock := map[string]int{
		keyExists: 1,
	}
	e := Entry{
		Version:   1,
		Timestamp: time.Now(),
		Clock:     expectedClock,
		Value:     valExists,
		Tombstone: false,
	}
	equals(t, e.GetClock(), expectedClock)

}

// GetClock should return an empty map for a key which doesn't exist
func TestEntryGetClockKeyNotExists(t *testing.T) {
	var e *Entry
	assert(t, len(e.GetClock()) == 0, "GetClock() returned non-empty map for nonexisting key")
}

// GetValue should return the value for a key which exists
func TestEntryGetValueKeyExists(t *testing.T) {
	e := Entry{
		Version:   1,
		Timestamp: time.Now(),
		Clock:     map[string]int{keyExists: 1},
		Value:     valExists,
		Tombstone: false,
	}
	equals(t, e.GetValue(), valExists)
}

// GetValue should return an empty string for a key which doesn't exist
func TestEntryGetValueKeyNotExists(t *testing.T) {
	var e *Entry
	equals(t, e.GetValue(), "")
}

// Update should change the timestamp
func TestUpdateChangesTimestamp(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		Version:   1,
		Timestamp: start,
		Clock:     map[string]int{keyExists: 1},
		Value:     valExists,
		Tombstone: false,
	}

	e.Update(keyExists, finish, map[string]int{}, valExists)
	equals(t, e.GetTimestamp(), finish)
}

// Update should increment the version
func TestUpdateIncrementsVersion(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		Version:   1,
		Timestamp: start,
		Clock:     map[string]int{keyExists: 1},
		Value:     valExists,
		Tombstone: false,
	}

	e.Update(keyExists, finish, map[string]int{}, valExists)
	equals(t, e.GetVersion(), 2)
}

// Update should reset the tombstone
func TestUpdateClearsTombstone(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		Version:   1,
		Timestamp: start,
		Clock:     map[string]int{keyExists: 1},
		Value:     valExists,
		Tombstone: true,
	}

	e.Update(keyExists, finish, map[string]int{}, valExists)
	equals(t, e.Alive(), true)
}

// Update should change the value
func TestUpdateChangesValue(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		Version:   1,
		Timestamp: start,
		Clock:     map[string]int{keyExists: 1},
		Value:     valone,
		Tombstone: true,
	}

	e.Update(keyExists, finish, map[string]int{}, valtwo)
	equals(t, valtwo, e.GetValue())
}

// Alive should return true for existing keys
func TestAliveKeyExistsReturnsTrue(t *testing.T) {
	e := Entry{
		Version:   1,
		Timestamp: time.Now(),
		Clock:     map[string]int{keyExists: 1},
		Value:     valone,
		Tombstone: false,
	}

	equals(t, true, e.Alive())
}

// Alive should return false for non-existing keys
func TestAliveKeyNotExistsReturnsFalse(t *testing.T) {
	var e *Entry
	equals(t, false, e.Alive())
}

// Delete should set a new timestamp
func TestDeleteSetsNewTimestamp(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		Version:   1,
		Timestamp: start,
		Clock:     map[string]int{keyExists: 1},
		Value:     valone,
		Tombstone: false,
	}

	e.Delete(keyExists, finish, map[string]int{})
	equals(t, finish, e.GetTimestamp())
}

// Delete should update the version
func TestDeleteUpdatesVersion(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		Version:   1,
		Timestamp: start,
		Clock:     map[string]int{keyExists: 1},
		Value:     valone,
		Tombstone: false,
	}

	e.Delete(keyExists, finish, map[string]int{})
	equals(t, 2, e.GetVersion())
}

// Delete should set the tombstone
func TestDeleteSetsTombstone(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		Version:   1,
		Timestamp: start,
		Clock:     map[string]int{keyExists: 1},
		Value:     valone,
		Tombstone: false,
	}

	e.Delete(keyExists, finish, map[string]int{})
	equals(t, false, e.Alive())
}

// Delete should clear the value
func TestDeleteClearsValue(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		Version:   1,
		Timestamp: start,
		Clock:     map[string]int{keyExists: 1},
		Value:     valone,
		Tombstone: false,
	}

	e.Delete(keyExists, finish, map[string]int{})
	equals(t, "", e.GetValue())
}

// Delete should set clock to the new version
func TestDeleteUpdatesClock(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		Version:   1,
		Timestamp: start,
		Clock:     map[string]int{keyExists: 1},
		Value:     valone,
		Tombstone: false,
	}

	e.Delete(keyExists, finish, map[string]int{})
	equals(t, map[string]int{keyExists: 2}, e.GetClock())
}

// Update should set the clock to include the key you've updated
func TestUpdateSetsInitialClock(t *testing.T) {
	e := Entry{
		Version:   1,
		Timestamp: time.Now(),
		Clock:     map[string]int{},
		Value:     valone,
		Tombstone: false,
	}
	initialClock := map[string]int{
		keyExists: 2,
	}
	e.Update(keyExists, time.Now(), initialClock, valtwo)
	equals(t, initialClock, e.GetClock())
}

// NewEntry should set an initial clock value
func TestNewEntrySetsInitialClock(t *testing.T) {
	initialClock := map[string]int{
		keyExists:        1,
		keyNotExists:     2,
		"some other key": 1,
	}
	e := NewEntry(time.Now(), initialClock, valExists, 1)

	equals(t, initialClock, e.GetClock())
}

func TestGetClockKeyExistsReturnsClock(t *testing.T) {
	initialClock := map[string]int{
		keyExists:        1,
		keyNotExists:     2,
		"some other key": 1,
	}
	var m sync.RWMutex
	e := NewEntry(time.Now(), initialClock, valExists, 1)
	d := map[string]KeyEntry{keyExists: e}
	k := KVS{db: d, mutex: &m}

	equals(t, initialClock, k.GetClock(keyExists))
}

func TestGetClockKeyNotExistsReturnsEmpty(t *testing.T) {
	k := NewKVS()
	equals(t, map[string]int{}, k.GetClock(keyExists))
}
func TestGetTimestampReturnsTimestamp(t *testing.T) {
	time := time.Now()
	initialClock := map[string]int{
		keyExists:        1,
		keyNotExists:     2,
		"some other key": 1,
	}
	var m sync.RWMutex
	e := NewEntry(time, initialClock, valExists, 1)
	d := map[string]KeyEntry{keyExists: e}
	k := KVS{db: d, mutex: &m}

	equals(t, time, k.GetTimestamp(keyExists))
}

func TestGetTimestampKeyNotExistsReturnsEmpty(t *testing.T) {
	k := NewKVS()
	equals(t, time.Time{}, k.GetTimestamp(keyExists))
}

func TestOverwriteKeyExists(t *testing.T) {
	// Define a starter entry
	firstTime := time.Now()
	firstClock := map[string]int{}
	firstVal := valExists
	firstVersion := 1

	first := Entry{
		Value:     firstVal,
		Clock:     firstClock,
		Timestamp: firstTime,
		Tombstone: false,
		Version:   firstVersion,
	}

	// Define a second entry to overwrite the first
	secondTime := time.Now()
	secondVal := valNotExists
	secondClock := map[string]int{keyExists: 2}
	secondVersion := 3
	second := Entry{
		Value:     secondVal,
		Clock:     secondClock,
		Timestamp: secondTime,
		Tombstone: false,
		Version:   secondVersion,
	}

	// Make a KVS
	var m sync.RWMutex
	k := KVS{
		db: map[string]KeyEntry{
			keyExists: &first,
		},
		mutex: &m,
	}

	// Overwrite the entry and verify result
	k.OverwriteEntry(keyExists, &second)
	equals(t, &second, k.db[keyExists])
}

func TestOverwriteEntryNotExists(t *testing.T) {
	// Define a second entry to overwrite the first
	secondTime := time.Now()
	secondVal := valNotExists
	secondClock := map[string]int{keyExists: 2}
	secondVersion := 3
	second := Entry{
		Value:     secondVal,
		Clock:     secondClock,
		Timestamp: secondTime,
		Tombstone: false,
		Version:   secondVersion,
	}

	// Make a KVS
	var m sync.RWMutex
	k := KVS{
		db:    map[string]KeyEntry{},
		mutex: &m,
	}

	// Overwrite the entry and verify result
	k.OverwriteEntry(keyExists, &second)
	equals(t, &second, k.db[keyExists])
}
func TestGetTimeGlobKeyExists(t *testing.T) {
	timestamp := time.Now()
	initialClock := map[string]int{
		keyExists:        1,
		keyNotExists:     2,
		"some other key": 1,
	}
	var m sync.RWMutex
	e := NewEntry(timestamp, initialClock, valExists, 1)
	d := map[string]KeyEntry{keyExists: e}
	k := KVS{db: d, mutex: &m}

	g := k.GetTimeGlob()
	h := make(map[string]time.Time)
	h[keyExists] = timestamp
	j := timeGlob{List: h}
	equals(t, j, g)
}

func TestGetTimeGlobKeyNotExists(t *testing.T) {
	k := NewKVS()
	h := make(map[string]time.Time)
	j := timeGlob{List: h}
	equals(t, j, k.GetTimeGlob())
}

func TestGetEntryGlobEmptyTimeGlob(t *testing.T) {
	k := NewKVS()
	g := timeGlob{}
	h := k.GetEntryGlob(g)
	j := entryGlob{Keys: map[string]Entry{}}
	equals(t, j, h)
}

func TestGetEntryGlobKeyExist(t *testing.T) {
	timestamp := time.Now()
	initialClock := map[string]int{
		keyExists:        1,
		keyNotExists:     2,
		"some other key": 1,
	}
	var m sync.RWMutex
	e := NewEntry(timestamp, initialClock, valExists, 1)
	d := map[string]KeyEntry{keyExists: e}
	k := KVS{db: d, mutex: &m}

	g := k.GetTimeGlob()
	h := k.GetEntryGlob(g)
	j := make(map[string]Entry)
	j[keyExists] = *e
	l := entryGlob{Keys: j}
	equals(t, l, h)
}

func TestSetVersionSetsVersion(t *testing.T) {
	e := NewEntry(time.Now(), map[string]int{}, valExists, 1)
	e.SetVersion(2)
	assert(t, e.Version == 2, "SetVersion didn't work")
}

func TestSetVersionEntryNilDoesntExplode(t *testing.T) {
	var e *Entry
	e.SetVersion(2)
	assert(t, true, "SetVersion exploded")
}

// TODO: TESTS NEEDED
// Need a test that runs Put on a key which does exist so it tests lines 274-280
//
// Need to test mergeClocks with no client clock
// Need to test mergeClocks iwth no server clock
// Need to test GetTimeGlob with a nil key
// Need to test GetEntryGlob with a nil key

// These functions were taken from Ben Johnson's post here: https://medium.com/@benbjohnson/structuring-tests-in-go-46ddee7a25c

// // assert fails the test if the condition is false.
// func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
// 	if !condition {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
// 		tb.FailNow()
// 	}
// }

// // ok fails the test if an err is not nil.
// func ok(tb testing.TB, err error) {
// 	if err != nil {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
// 		tb.FailNow()
// 	}
// }

// // equals fails the test if exp is not equal to act.
// func equals(tb testing.TB, exp, act interface{}) {
// 	if !reflect.DeepEqual(exp, act) {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
// 		tb.FailNow()
// 	}
// }
