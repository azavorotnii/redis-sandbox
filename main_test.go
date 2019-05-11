package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func setupClient(tb testing.TB) (*redis.Client, func()) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	tearDown := func() {
		if err := client.Close(); err != nil {
			tb.Error(err)
		}
	}
	defer func() {
		if tb.Failed() {
			tearDown()
		}
	}()

	if result, err := client.Ping().Result(); err != nil {
		tb.Fatal(result, err)
	}
	return client, tearDown
}

func TestGetNil(t *testing.T) {
	// just to test connection
	client, tearDown := setupClient(t)
	defer tearDown()

	_, err := client.Get("dont_exists").Result()
	require.Equal(t, err, redis.Nil)
}
func randomKey(keyLen int) string {
	buf := make([]byte, (keyLen*3+1)/4)
	if n, err := rand.Read(buf); err != nil || n != len(buf) {
		panic(fmt.Errorf("%v %v", n, err))
	}
	key := base64.RawURLEncoding.EncodeToString(buf)
	//if len(key) > keyLen + 3 || len(key) < keyLen - 3 {
	//	panic(fmt.Errorf("%v %v", len(key), keyLen))
	//}
	return key[:keyLen]
}

func TestRandomKey(t *testing.T) {
	for i := 1; i < 100; i++ {
		log.Println(randomKey(i))
	}
}

func BenchmarkGet(b *testing.B) {
	client, tearDown := setupClient(b)
	defer tearDown()

	cases := []struct {
		name   string
		keyLen int
	}{
		{"short_key", 4},
		{"long_key", 100},
	}
	for _, c := range cases {
		keyLen := c.keyLen

		setupStart := time.Now()
		const keysCount = 1000000
		var keys []string
		for i := 0; i < keysCount; i++ {
			key := randomKey(keyLen)
			err := client.Set(key, key, 0).Err()
			require.Nil(b, err)
			keys = append(keys, key)
		}
		b.Log("setup", c.name, keysCount, time.Since(setupStart))

		b.Run(c.name+"_same_key", func(b *testing.B) {
			key := keys[rand.Intn(len(keys))]
			for i := 0; i < b.N; i++ {
				value, err := client.Get(key).Result()
				require.Nil(b, err)
				require.Equal(b, value, key)
			}
		})
		b.Run(c.name+"_random_key", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				key := keys[rand.Intn(len(keys))]

				value, err := client.Get(key).Result()
				require.Nil(b, err)
				require.Equal(b, value, key)
			}
		})

		teardownStart := time.Now()
		err := client.Del(keys...).Err()
		require.Nil(b, err)
		b.Log("tearDown", c.name, len(keys), time.Since(teardownStart))
	}
}

func BenchmarkHGet(b *testing.B) {
	client, tearDown := setupClient(b)
	defer tearDown()

	const hashKey = "key"

	cases := []struct {
		name   string
		keyLen int
	}{
		{"short_field", 4},
		{"long_field", 100},
	}
	for _, c := range cases {
		keyLen := c.keyLen

		setupStart := time.Now()
		const keysCount = 1000000
		var fields []string
		for i := 0; i < keysCount; i++ {
			field := randomKey(keyLen)
			err := client.HSet(hashKey, field, field).Err()
			require.Nil(b, err)
			fields = append(fields, field)
		}
		b.Log("setup", c.name, keysCount, time.Since(setupStart))

		b.Run(c.name+"_same_field", func(b *testing.B) {
			field := fields[rand.Intn(len(fields))]
			for i := 0; i < b.N; i++ {
				value, err := client.HGet(hashKey, field).Result()
				require.Nil(b, err)
				require.Equal(b, value, field)
			}
		})
		b.Run(c.name+"_random_field", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				field := fields[rand.Intn(len(fields))]

				value, err := client.HGet(hashKey, field).Result()
				require.Nil(b, err)
				require.Equal(b, value, field)
			}
		})

		teardownStart := time.Now()
		err := client.Del(hashKey).Err()
		require.Nil(b, err)
		b.Log("tearDown", c.name, len(fields), time.Since(teardownStart))
	}
}
