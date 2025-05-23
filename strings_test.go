//go:build alltest
// +build alltest

package main

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestKV(t *testing.T) {
	c := getTestConn()
	ctx := context.Background()

	if ok, err := c.Set(ctx, "a", "1234", 0).Result(); err != nil {
		t.Fatal(err)
	} else if ok != "OK" {
		t.Fatalf("set err")
	}

	if set, err := c.SetNX(ctx, "a", "123", 0).Result(); err != nil {
		t.Fatal(err)
	} else if set != false {
		t.Fatalf("setnx err")
	}

	if set, err := c.SetNX(ctx, "b", "123", 0).Result(); err != nil {
		t.Fatal(err)
	} else if set != true {
		t.Fatalf("setnx err")
	}

	if ok, err := c.SetEx(ctx, "mykey", "hello", 10*time.Second).Result(); err != nil {
		t.Fatal(err)
	} else if ok != "OK" {
		t.Fatalf("setex err")
	}
	if v, err := c.Get(ctx, "a").Result(); err != nil {
		t.Fatal(err)
	} else if v != "1234" {
		t.Fatal(v)
	}

	if v, err := c.GetSet(ctx, "a", "123").Result(); err != nil {
		t.Fatal(err)
	} else if v != "1234" {
		t.Fatal(v)
	}

	if v, err := c.Get(ctx, "a").Result(); err != nil {
		t.Fatal(err)
	} else if v != "123" {
		t.Fatal(v)
	}

	if n, err := c.Exists(ctx, "a").Result(); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatal(n)
	}

	if n, err := c.Exists(ctx, "a", "b").Result(); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := c.Exists(ctx, "a", "a", "a").Result(); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatal(n)
	}

	if n, err := c.Exists(ctx, "a", "b", "c", "d").Result(); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatal(n)
	}

	if n, err := c.Exists(ctx, "c", "d").Result(); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := c.Exists(ctx, "empty_key_test").Result(); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if _, err := c.Del(ctx, "a", "b").Result(); err != nil {
		t.Fatal(err)
	}

	if n, err := c.Exists(ctx, "a").Result(); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	if n, err := c.Exists(ctx, "b").Result(); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatal(n)
	}

	rangeKey := "range_key"
	if n, err := c.Append(ctx, rangeKey, "Hello ").Result(); err != nil {
		t.Fatal(err)
	} else if n != 6 {
		t.Fatal(n)
	}

	if n, err := c.SetRange(ctx, rangeKey, 6, "Redis").Result(); err != nil {
		t.Fatal(err)
	} else if n != 11 {
		t.Fatal(n)
	}

	if n, err := c.StrLen(ctx, rangeKey).Result(); err != nil {
		t.Fatal(err)
	} else if n != 11 {
		t.Fatal(n)
	}

	if v, err := c.GetRange(ctx, rangeKey, 0, -1).Result(); err != nil {
		t.Fatal(err)
	} else if v != "Hello Redis" {
		t.Fatal(v)
	}

	bitKey := "test:bitcount:key"

	// Clean up the key after tests
	defer c.Del(context.Background(), bitKey)

	// Set bits at positions 1, 7, and 14
	for _, pos := range []int{1, 7, 14} {
		if n, err := c.SetBit(context.Background(), bitKey, int64(pos), 1).Result(); err != nil {
			t.Fatal(err)
		} else if n != 0 {
			t.Fatal(n)
		}
	}

	// Verify the bits at positions 1, 7, and 14
	for _, pos := range []int{1, 7, 14} {
		if n, err := c.GetBit(context.Background(), bitKey, int64(pos)).Result(); err != nil {
			t.Fatal(err)
		} else if n != 1 {
			t.Fatal(n)
		}
	}

	// Test BITCOUNT with no start or end
	if n, err := c.BitCount(context.Background(), bitKey, nil).Result(); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatalf("expected 3, got %d", n)
	}

	// Test BITCOUNT with start and end in BYTE mode: first byte
	if n, err := c.BitCount(context.Background(), bitKey, &redis.BitCount{Start: 0, End: 0, Unit: "BYTE"}).Result(); err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatalf("expected 2, got %d", n)
	}

	// Test BITCOUNT with start and end in BIT mode
	if n, err := c.BitCount(context.Background(), bitKey, &redis.BitCount{Start: 0, End: 15, Unit: "BIT"}).Result(); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatalf("expected 3, got %d", n)
	}

	// Test BITCOUNT with start > end
	_, err := c.BitCount(context.Background(), bitKey, &redis.BitCount{Start: 1, End: 0, Unit: "BYTE"}).Result()
	if err == nil {
		t.Fatal("expected an error, got nil")
	} else if err.Error() != "byte invalid range: start > end" {
		t.Fatalf("expected 'ERR invalid byte range: start > end', got '%s'", err.Error())
	}

	// Test BITCOUNT with start and end out of range
	if _, err := c.BitCount(context.Background(), bitKey, &redis.BitCount{Start: 100, End: 200, Unit: "BIT"}).Result(); err == nil {
		t.Fatal("expected error, got nil")
	} else if err.Error() != "bit range out of bounds" {
		t.Fatalf("expected 'bit range out of bounds' error, got %v", err)
	}

	// Test BITPOS with only bit=1
	if n, err := c.BitPos(context.Background(), bitKey, 1).Result(); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	// Test BITPOS with start provided
	if n, err := c.BitPos(context.Background(), bitKey, 1, 0).Result(); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	// Test BITPOS with start and end provided
	if n, err := c.BitPos(context.Background(), bitKey, 1, 0, 15).Result(); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	// Test BITPOS with bit=0
	if n, err := c.BitPos(context.Background(), bitKey, 0).Result(); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}

	// Test BITPOS with start > end
	if n, err := c.BitPos(context.Background(), bitKey, 1, 15, 10).Result(); err != nil {
		t.Fatal(err)
	} else if n != -1 {
		t.Fatalf("expected -1, got %d", n)
	}

	// Test BITPOS with start beyond the highest set bit
	if n, err := c.BitPos(context.Background(), bitKey, 1, 15, 15).Result(); err != nil {
		t.Fatal(err)
	} else if n != -1 {
		t.Fatalf("expected -1, got %d", n)
	}

	// Test BitPosSpan with start and end in BYTE mode
	if n, err := c.BitPosSpan(context.Background(), bitKey, 1, 1, 1, "byte").Result(); err != nil {
		t.Fatal(err)
	} else if n != 14 {
		t.Fatalf("expected 14, got %d", n)
	}

	// Test BitPosSpan with bit=0 in BIT mode
	if n, err := c.BitPosSpan(context.Background(), bitKey, 0, 0, 15, "bit").Result(); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}

	// Test BitPosSpan with bit=0 in BYTE mode
	if n, err := c.BitPosSpan(context.Background(), bitKey, 0, 0, 1, "byte").Result(); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}

	// Test BitPosSpan with start > end in BIT mode
	if n, err := c.BitPosSpan(context.Background(), bitKey, 1, 15, 10, "bit").Result(); err != nil {
		t.Fatal(err)
	} else if n != -1 {
		t.Fatalf("expected -1, got %d", n)
	}

	// Test BitPosSpan with start > end in BYTE mode
	if n, err := c.BitPosSpan(context.Background(), bitKey, 1, 2, 1, "byte").Result(); err != nil {
		t.Fatal(err)
	} else if n != -1 {
		t.Fatalf("expected -1, got %d", n)
	}

	// Test BitPosSpan with start超出范围在 BIT mode
	if n, err := c.BitPosSpan(context.Background(), bitKey, 1, 100, 150, "bit").Result(); err != nil {
		t.Fatal(err)
	} else if n != -1 {
		t.Fatalf("expected -1, got %d", n)
	}

	// Test BitPosSpan with end超出范围在 BYTE mode
	if n, err := c.BitPosSpan(context.Background(), bitKey, 1, 0, 100, "byte").Result(); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	// Test BitPosSpan with start=0, end=-1 in BIT mode
	if n, err := c.BitPosSpan(context.Background(), bitKey, 1, 0, -1, "bit").Result(); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	// Test BitPosSpan with start=0, end=-1 in BYTE mode
	if n, err := c.BitPosSpan(context.Background(), bitKey, 1, 0, -1, "byte").Result(); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}

	c.Set(ctx, "key1", "foobar", 0)
	c.Set(ctx, "key2", "abcdef", 0)

	if n, err := c.BitOpAnd(ctx, "bit_dest_key", "key1", "key2").Result(); err != nil {
		t.Fatal(err)
	} else if n != 6 {
		t.Fatal(n)
	}

	if v, err := c.Get(ctx, "bit_dest_key").Result(); err != nil {
		t.Fatal(err)
	} else if v != "`bc`ab" {
		t.Error(v)
	}
}

func TestMGET(t *testing.T) {
	c := getTestConn()
	ctx := context.Background()

	// Set up initial key-value pairs
	if ok, err := c.MSet(ctx, "a", "1", "b", "2").Result(); err != nil {
		t.Fatalf("MSET failed: %v", err)
	} else if ok != "OK" {
		t.Fatalf("MSET returned unexpected result: %v", ok)
	}

	// Retrieve values using MGET
	values, err := c.MGet(ctx, "a", "b", "c").Result()
	if err != nil {
		t.Fatalf("MGET failed: %v", err)
	}

	// Validate the number of returned values
	if len(values) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(values))
	}

	// Validate the value for key "a"
	if v, ok := values[0].(string); !ok || v != "1" {
		t.Errorf("Expected value '1' for key 'a', got %v", values[0])
	}

	// Validate the value for key "b"
	if v, ok := values[1].(string); !ok || v != "2" {
		t.Errorf("Expected value '2' for key 'b', got %v", values[1])
	}

	// Validate the value for non-existent key "c"
	if values[2] != nil {
		t.Errorf("Expected nil for non-existent key 'c', got %v", values[2])
	}
}

func TestKVIncrDecr(t *testing.T) {
	c := getTestConn()

	ctx := context.Background()
	if n, err := c.Incr(ctx, "n").Result(); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Error(n)
	}

	if n, err := c.Incr(ctx, "n").Result(); err != nil {
		t.Error(err)
	} else if n != 2 {
		t.Error(n)
	}

	if n, err := c.Decr(ctx, "n").Result(); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Error(n)
	}

	if n, err := c.IncrBy(ctx, "n", 10).Result(); err != nil {
		t.Error(err)
	} else if n != 11 {
		t.Error(n)
	}

	if n, err := c.DecrBy(ctx, "n", 10).Result(); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Error(n)
	}
}

func TestKVErrorParams(t *testing.T) {
	c := getTestConn()
	ctx := context.Background()

	if _, err := c.Do(ctx, "get", "a", "b", "c").Result(); err == nil {
		t.Errorf("invalid err %v", err)
	}

	if _, err := c.Do(ctx, "set", "a", "b", "c").Result(); err == nil {
		t.Errorf("invalid err %v", err)
	}

	if _, err := c.Do(ctx, "getset", "a", "b", "c").Result(); err == nil {
		t.Errorf("invalid err %v", err)
	}

	if _, err := c.Do(ctx, "setnx", "a", "b", "c").Result(); err == nil {
		t.Errorf("invalid err %v", err)
	}

	if _, err := c.Do(ctx, "incr", "a", "b").Result(); err == nil {
		t.Errorf("invalid err %v", err)
	}

	if _, err := c.Do(ctx, "incrby", "a").Result(); err == nil {
		t.Errorf("invalid err %v", err)
	}

	if _, err := c.Do(ctx, "decrby", "a").Result(); err == nil {
		t.Errorf("invalid err %v", err)
	}

	if _, err := c.Do(ctx, "del").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "mset").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "mset", "a", "b", "c").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "mget").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "expire").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "expire", "a", "b").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "expireat").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "expireat", "a", "b").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "ttl").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "persist").Result(); err == nil {
		t.Errorf("invalid err of %v", err)
	}

	if _, err := c.Do(ctx, "setex", "a", "blah", "hello world").Result(); err == nil {
		t.Errorf("invalid err %v", err)
	}
}
