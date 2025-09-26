package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	assert.NotNil(t, store)
	assert.NotNil(t, store.data)
	assert.NotNil(t, store.subscribers)
}

func TestMemoryStore_SetGet(t *testing.T) {
	store := NewMemoryStore()

	t.Run("set and get value", func(t *testing.T) {
		key := "test-key"
		value := []byte("test-value")

		err := store.Set(key, value, 0)
		assert.NoError(t, err)

		result, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		result, err := store.Get("non-existent")
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("set with TTL", func(t *testing.T) {
		key := "ttl-key"
		value := []byte("ttl-value")
		ttl := 100 * time.Millisecond

		err := store.Set(key, value, ttl)
		assert.NoError(t, err)

		// Should exist immediately
		result, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// Should expire after TTL
		time.Sleep(150 * time.Millisecond)
		result, err = store.Get(key)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Nil(t, result)
	})
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()

	t.Run("delete existing key", func(t *testing.T) {
		key := "delete-key"
		value := []byte("delete-value")

		store.Set(key, value, 0)
		err := store.Delete(key)
		assert.NoError(t, err)

		_, err = store.Get(key)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})

	t.Run("delete non-existent key", func(t *testing.T) {
		err := store.Delete("non-existent")
		assert.NoError(t, err) // Should not error
	})
}

func TestMemoryStore_Del(t *testing.T) {
	store := NewMemoryStore()

	t.Run("delete multiple keys", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}

		// Set keys
		for _, key := range keys {
			store.Set(key, []byte("value"), 0)
		}

		err := store.Del(keys...)
		assert.NoError(t, err)

		// Verify all keys are deleted
		for _, key := range keys {
			_, err := store.Get(key)
			assert.Error(t, err)
			assert.Equal(t, ErrNotFound, err)
		}
	})

	t.Run("delete empty list", func(t *testing.T) {
		err := store.Del()
		assert.NoError(t, err)
	})
}

func TestMemoryStore_Exists(t *testing.T) {
	store := NewMemoryStore()

	t.Run("existing key", func(t *testing.T) {
		key := "exists-key"
		store.Set(key, []byte("value"), 0)

		exists, err := store.Exists(key)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existent key", func(t *testing.T) {
		exists, err := store.Exists("non-existent")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("expired key", func(t *testing.T) {
		key := "expired-key"
		store.Set(key, []byte("value"), 10*time.Millisecond)

		time.Sleep(20 * time.Millisecond)
		exists, err := store.Exists(key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestMemoryStore_SetNX(t *testing.T) {
	store := NewMemoryStore()

	t.Run("set new key", func(t *testing.T) {
		key := "setnx-key"
		value := []byte("setnx-value")

		set, err := store.SetNX(key, value, 0)
		assert.NoError(t, err)
		assert.True(t, set)

		result, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("set existing key", func(t *testing.T) {
		key := "existing-key"
		store.Set(key, []byte("original"), 0)

		set, err := store.SetNX(key, []byte("new"), 0)
		assert.NoError(t, err)
		assert.False(t, set)

		// Value should remain unchanged
		result, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, []byte("original"), result)
	})
}

func TestMemoryStore_HSet_HGetAll(t *testing.T) {
	store := NewMemoryStore()

	t.Run("hash operations", func(t *testing.T) {
		key := "hash-key"
		fields := map[string]any{
			"field1": "value1",
			"field2": 42,
			"field3": true,
		}

		err := store.HSet(key, fields)
		assert.NoError(t, err)

		result, err := store.HGetAll(key)
		assert.NoError(t, err)
		assert.Equal(t, "value1", result["field1"])
		assert.Equal(t, "42", result["field2"])
		assert.Equal(t, "true", result["field3"])
	})

	t.Run("get non-existent hash", func(t *testing.T) {
		result, err := store.HGetAll("non-existent")
		assert.NoError(t, err) // Returns empty map, not error
		assert.NotNil(t, result)
		assert.Len(t, result, 0) // Empty map
	})
}

func TestMemoryStore_HIncrBy(t *testing.T) {
	store := NewMemoryStore()

	t.Run("increment existing field", func(t *testing.T) {
		key := "incr-key"
		field := "counter"

		// Set initial value
		store.HSet(key, map[string]any{field: "10"})

		result, err := store.HIncrBy(key, field, 5)
		assert.NoError(t, err)
		assert.Equal(t, int64(15), result)

		// Verify stored value
		hash, err := store.HGetAll(key)
		assert.NoError(t, err)
		assert.Equal(t, "15", hash[field])
	})

	t.Run("increment non-existent field", func(t *testing.T) {
		key := "incr-key2"
		field := "new-counter"

		result, err := store.HIncrBy(key, field, 3)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result)
	})

	t.Run("increment non-numeric field", func(t *testing.T) {
		key := "incr-key3"
		field := "text-field"

		store.HSet(key, map[string]any{field: "not-a-number"})

		result, err := store.HIncrBy(key, field, 1)
		assert.NoError(t, err) // ParseInt treats non-numeric as 0
		assert.Equal(t, int64(1), result) // 0 + 1 = 1
	})
}

func TestMemoryStore_ListOperations(t *testing.T) {
	store := NewMemoryStore()

	t.Run("LPush and Rotate", func(t *testing.T) {
		key := "list-key"

		// Push values (LPush prepends, so order is reversed)
		err := store.LPush(key, "value1", "value2", "value3")
		assert.NoError(t, err)

		// Rotate should return values in the order they were stored
		// LPush prepends, so order becomes: value3, value2, value1
		result, err := store.Rotate(key)
		assert.NoError(t, err)
		assert.Equal(t, "value3", result)

		result, err = store.Rotate(key)
		assert.NoError(t, err)
		assert.Equal(t, "value2", result)

		result, err = store.Rotate(key)
		assert.NoError(t, err)
		assert.Equal(t, "value1", result)

		// Should cycle back to first
		result, err = store.Rotate(key)
		assert.NoError(t, err)
		assert.Equal(t, "value3", result)
	})

	t.Run("rotate empty list", func(t *testing.T) {
		result, err := store.Rotate("empty-list")
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
		assert.Equal(t, "", result)
	})

	t.Run("LRem", func(t *testing.T) {
		key := "rem-list"
		store.LPush(key, "a", "b", "c", "b", "d")

		err := store.LRem(key, 0, "b") // Remove all occurrences of "b"
		assert.NoError(t, err)

		// Verify "b" is removed by rotating through the list
		values := []string{}
		for i := 0; i < 3; i++ { // Should have 3 values left
			val, err := store.Rotate(key)
			assert.NoError(t, err)
			values = append(values, val)
		}

		// Should contain "a", "c", "d" but not "b"
		assert.Contains(t, values, "a")
		assert.Contains(t, values, "c")
		assert.Contains(t, values, "d")
		assert.NotContains(t, values, "b")
	})
}

func TestMemoryStore_SetOperations(t *testing.T) {
	store := NewMemoryStore()

	t.Run("SAdd and SPopN", func(t *testing.T) {
		key := "set-key"

		err := store.SAdd(key, "member1", "member2", "member3")
		assert.NoError(t, err)

		members, err := store.SPopN(key, 2)
		assert.NoError(t, err)
		assert.Len(t, members, 2)

		// Remaining member
		remaining, err := store.SPopN(key, 5) // Ask for more than available
		assert.NoError(t, err)
		assert.Len(t, remaining, 1)

		// Set should be empty now
		empty, err := store.SPopN(key, 1)
		assert.NoError(t, err)
		assert.Len(t, empty, 0)
	})

	t.Run("SPopN from non-existent set", func(t *testing.T) {
		members, err := store.SPopN("non-existent", 1)
		assert.NoError(t, err)
		assert.Len(t, members, 0)
	})
}

func TestMemoryStore_PubSub(t *testing.T) {
	store := NewMemoryStore()

	t.Run("publish and subscribe", func(t *testing.T) {
		channel := "test-channel"
		message := []byte("test message")

		// Subscribe
		sub, err := store.Subscribe(channel)
		assert.NoError(t, err)
		assert.NotNil(t, sub)

		// Publish message
		err = store.Publish(channel, message)
		assert.NoError(t, err)

		// Receive message
		select {
		case msg := <-sub.Channel():
			assert.Equal(t, channel, msg.Channel)
			assert.Equal(t, message, msg.Payload)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Message not received within timeout")
		}

		// Clean up
		sub.Close()
	})

	t.Run("multiple subscribers", func(t *testing.T) {
		channel := "multi-channel"
		message := []byte("multi message")

		// Create multiple subscribers
		sub1, _ := store.Subscribe(channel)
		sub2, _ := store.Subscribe(channel)

		// Publish message
		store.Publish(channel, message)

		// Both should receive the message
		for i, sub := range []*memorySubscription{sub1.(*memorySubscription), sub2.(*memorySubscription)} {
			select {
			case msg := <-sub.Channel():
				assert.Equal(t, channel, msg.Channel)
				assert.Equal(t, message, msg.Payload)
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("Subscriber %d did not receive message", i+1)
			}
		}

		// Clean up
		sub1.Close()
		sub2.Close()
	})
}

func TestMemoryStore_Clear(t *testing.T) {
	store := NewMemoryStore()

	// Add some data
	store.Set("key1", []byte("value1"), 0)
	store.HSet("hash1", map[string]any{"field": "value"})
	store.LPush("list1", "item")
	store.SAdd("set1", "member")

	err := store.Clear()
	assert.NoError(t, err)

	// Verify all data is cleared
	_, err = store.Get("key1")
	assert.Equal(t, ErrNotFound, err)

	result, err := store.HGetAll("hash1")
	assert.NoError(t, err) // Returns empty map, not error
	assert.Len(t, result, 0)

	_, err = store.Rotate("list1")
	assert.Equal(t, ErrNotFound, err)

	members, _ := store.SPopN("set1", 1)
	assert.Len(t, members, 0)
}

func TestMemoryStore_Close(t *testing.T) {
	store := NewMemoryStore()

	err := store.Close()
	assert.NoError(t, err)
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()

	// Test concurrent access doesn't cause race conditions
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			store.Set("key", []byte("value"), 0)
			store.Get("key")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			store.Set("key2", []byte("value2"), 0)
			store.Get("key2")
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}
