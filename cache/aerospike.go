package cache

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	as "github.com/aerospike/aerospike-client-go"
)

type Aerospike struct {
	client    *as.Client
	namespace string
}

func NewAerospikeCache() *Aerospike {
	instance := &Aerospike{}
	if os.Getenv("CACHE") == "true" {

		asHosts := getHosts()
		hosts := []*as.Host{}
		for _, host := range asHosts {
			hostAndPort := strings.Split(host, ":")
			port, err := strconv.Atoi(hostAndPort[1])
			if err == nil {
				hosts = append(hosts, as.NewHost(hostAndPort[0], port))
			}

		}

		client, err := as.NewClientWithPolicyAndHost(nil, hosts...)
		if err != nil {
			return instance
		}
		instance.namespace = os.Getenv("AEROSPIKE_NAMESPACE")
		instance.client = client
	}

	return instance
}

func (cache *Aerospike) SetJson(set string, key string, data interface{}, expiration int) error {

	if os.Getenv("CACHE") != "true" {
		return nil
	}

	if cache.client == nil {
		return errors.New("Client is nil for given Aerospike instance")
	}

	asKey, err := as.NewKey(cache.namespace, set, key)
	if err != nil {
		return err
	}
	dataEncoded, _ := json.Marshal(data)
	binVal := as.NewBin("val", dataEncoded)
	if err != nil {
		return err
	}
	//PutBin expects a WritePolicy here we can pass expiration
	// writePolicy := as.NewWritePolicy(0, 0)
	// writePolicy.Expiration = 2 (int seconds)
	// writePolicy.Expiration = 0 — Use the default TTL(Time To Live) specified on the server side on each record update.
	// writePolicy.Expiration = math.MaxUint32 — Never expire.
	var writePolicy = as.NewWritePolicy(0, 0)
	writePolicy.Expiration = uint32(expiration)

	return cache.client.PutBins(writePolicy, asKey, binVal)

}

func (cache *Aerospike) GetJson(set string, key string, container interface{}) (interface{}, error) {

	if os.Getenv("CACHE") != "true" {
		return nil, nil
	}

	if cache.client == nil {
		return nil, errors.New("Client is nil for given Aerospike instance")
	}

	asKey, err := as.NewKey(cache.namespace, set, key)
	if err != nil {
		return nil, err
	}
	data, err := cache.client.Get(nil, asKey, "val")
	if err != nil {
		return nil, err
	}

	if data != nil {
		err = json.Unmarshal(data.Bins["val"].([]byte), &container)
		if err != nil {
			return nil, err
		}

		return container, nil
	}
	return nil, nil

}

func (cache *Aerospike) Close() {
	cache.client.Close()
}

func getHosts() []string {
	hosts := strings.Split(os.Getenv("AEROSPIKE_HOSTS"), ",")
	return hosts
}
