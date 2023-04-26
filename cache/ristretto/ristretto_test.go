/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ristretto

import (
	"testing"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/deislabs/ratify/cache"
	"github.com/dgraph-io/ristretto/z"
)

// TestKeytoHash_Expected tests the keyToHash function
func TestKeytoHash_Expected(t *testing.T) {
	key := "test"
	hash1, hash2 := keyToHash(nil)
	if hash1 != 0 || hash2 != 0 {
		t.Errorf("Expected hash1 and hash2 to be 0, but got %d and %d", hash1, hash2)
	}
	hash1, hash2 = keyToHash([]byte(key))
	if hash1 != 0 || hash2 != 0 {
		t.Errorf("Expected hash1 and hash2 to be 0, but got %d and %d", hash1, hash2)
	}
	hash1, hash2 = keyToHash(key)
	actualHash1 := z.MemHashString(key)
	actualHash2 := xxhash.Sum64String(key)

	if hash1 != actualHash1 || hash2 != actualHash2 {
		t.Errorf("Expected hash1 and hash2 to be %d and %d, but got %d and %d", actualHash1, actualHash2, hash1, hash2)
	}
}

// TestSet_Expected tests the Set function
func TestSet_Expected(t *testing.T) {
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider, err := cache.GetCacheProvider()
	if err != nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider("ristretto", 100, 100)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	ok := cacheProvider.Set("test", "test")
	if !ok {
		t.Errorf("Expected ok to be true")
	}
	// wait for the cache to be set
	time.Sleep(1 * time.Second)
	ristrettoCache := cacheProvider.(*ristrettoCache)
	val, found := ristrettoCache.memoryCache.Get("test")
	if val == nil || !found {
		t.Errorf("Expected value to be set in memory cache")
	}
	if val.(string) != "test" {
		t.Errorf("Expected value to be test, but got %s", val.(string))
	}
}

// TestSetWithTTL_Expected tests the SetWithTTL function
func TestSetWithTTL_Expected(t *testing.T) {
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider, err := cache.GetCacheProvider()
	if err != nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider("ristretto", 100, 100)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	ok := cacheProvider.SetWithTTL("test_ttl", "test", 3*time.Second)
	if !ok {
		t.Errorf("Expected ok to be true")
	}
	// wait for the cache to be set
	time.Sleep(1 * time.Second)
	ristrettoCache := cacheProvider.(*ristrettoCache)
	val, found := ristrettoCache.memoryCache.Get("test_ttl")
	if val == nil || !found {
		t.Errorf("Expected value to be set in memory cache")
	}
	// wait for the cache entry to expire
	time.Sleep(3 * time.Second)
	val, found = ristrettoCache.memoryCache.Get("test_ttl")
	if val != nil || found {
		t.Errorf("Expected value to be deleted from memory cache")
	}
}

// TestGet_Expected tests the Get function
func TestGet_Expected(t *testing.T) {
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider, err := cache.GetCacheProvider()
	if err != nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider("ristretto", 100, 100)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	ristrettoCache := cacheProvider.(*ristrettoCache)
	success := ristrettoCache.memoryCache.Set("test-get", "test-get-val", 1)
	if !success {
		t.Errorf("Expected success to be true")
	}
	// wait for the cache to be set
	time.Sleep(1 * time.Second)
	val, ok := cacheProvider.Get("test-get")
	if !ok {
		t.Errorf("Expected ok to be true")
	}
	if val.(string) != "test-get-val" {
		t.Errorf("Expected value to be test-get-val, but got %s", val.(string))
	}
}

// TestDelete_Expected tests the Delete function
func TestDelete_Expected(t *testing.T) {
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider, err := cache.GetCacheProvider()
	if err != nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider("ristretto", 100, 100)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	ristrettoCache := cacheProvider.(*ristrettoCache)
	success := ristrettoCache.memoryCache.Set("test-delete", "test-delete-val", 1)
	if !success {
		t.Errorf("Expected success to be true")
	}
	// wait for the cache to be set
	time.Sleep(1 * time.Second)
	cacheProvider.Delete("test-delete")
	// wait for the cache entry to delete
	time.Sleep(1 * time.Second)
	val, found := ristrettoCache.memoryCache.Get("test-delete")
	if val != nil || found {
		t.Errorf("Expected value to be deleted from memory cache")
	}
}

// TestClear_Expected tests the Clear function
func TestClear_Expected(t *testing.T) {
	// first attempt to get the cache provider if it's already been initialized
	cacheProvider, err := cache.GetCacheProvider()
	if err != nil {
		// if no cache provider has been initialized, initialize one
		cacheProvider, err = cache.NewCacheProvider("ristretto", 100, 100)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
	ristrettoCache := cacheProvider.(*ristrettoCache)
	success := ristrettoCache.memoryCache.Set("test-clear", "test-clear-val", 1)
	if !success {
		t.Errorf("Expected success to be true")
	}
	// wait for the cache to be set
	time.Sleep(1 * time.Second)
	cacheProvider.Clear()
	// wait for the cache to clear
	time.Sleep(1 * time.Second)
	val, found := ristrettoCache.memoryCache.Get("test-clear")
	if val != nil || found {
		t.Errorf("Expected all values to be flushed from memory cache")
	}
}
