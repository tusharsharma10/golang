package cache

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/joho/godotenv"
)

var aerospike *Aerospike

func TestAerospike_GetSetJson(t *testing.T) {

	toBeCached := map[string]interface{}{
		"test-num": "5",
		"test-str": "hello how are you. ok bye",
		"test-obj": map[string]interface{}{
			"name": "dona",
			"age":  "100",
		},
	}

	var ar Aerospike

	for key, data := range toBeCached {
		err := ar.SetJson("test", key, data, 1)
		if err.Error() != "Client is nil for given Aerospike instance" {
			t.Errorf("\nError while setting cache:\nCacheKey: %s \nErr: %s\n", key, err)
		}

		err = aerospike.SetJson("test", key, data, 1)
		if err != nil {
			fmt.Println(err)
			t.Errorf("\nError while setting cache:\nCacheKey: %s \nErr: %s\n", key, err)
		}

		var got interface{}
		got, err = aerospike.GetJson("test", key, got)
		if err != nil {
			t.Errorf("\nError while reading from cache:\nCacheKey: %s \nErr: %s\n", key, err)
		}

		if !reflect.DeepEqual(got, data) {
			t.Errorf("Error while reading from cache:\n CacheKey: %s\n Got: %v\n Want: %v", key, got, data)
		}

	}
}

func Benchmark_Aerospike(b *testing.B) {
	toBeCached := map[string]interface{}{
		//"test-num": 5,
		"test-str":  "hello how are you. ok bye",
		"test-str2": "hello how are you. ok bye1",
		"test-str3": "hello how are you. ok bye2",
		"test-obj": map[string]interface{}{
			"name": "dona",
			"age":  "10",
		},
	}

	for key, data := range toBeCached {
		err := aerospike.SetJson("test", key, data, 1)
		if err != nil {
			b.Errorf("\nError while setting cache:\nCacheKey: %s \nErr: %s\n", key, err)
		}

		var got interface{}
		got, err = aerospike.GetJson("test", key, got)
		if err != nil {
			b.Errorf("\nError while reading from cache:\nCacheKey: %s \nErr: %s\n", key, err)
		}

		if !reflect.DeepEqual(got, data) {
			b.Errorf("Error while reading from cache:\n CacheKey: %s\n Got: %v\n Want: %v", key, got, data)
		}

	}
}

func init() {
	if len(os.Getenv("CACHE")) == 0 {
		_, b, _, _ := runtime.Caller(0)
		basepath := filepath.Dir(b)
		ap := path.Join(basepath, "../", ".development.env")

		if err := godotenv.Load(ap); err != nil {
			log.Fatalf("%s", err)
		}
	}

	aerospike = NewAerospikeCache()
}
