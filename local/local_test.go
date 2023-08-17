package local_test

import (
	"encoding/json"
	"github.com/burybell/oss"
	"github.com/burybell/oss/local"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	objectStore oss.ObjectStore
	bucket      oss.Bucket
)

type Config struct {
	Local           local.Config `json:"local"`
	LocalBucketName string       `json:"local_bucket_name"`
}

func init() {
	f, err := os.Open("../config.json")
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		panic(err)
	}
	objectStore = local.MustNewObjectStore(config.Local)
	bucket = objectStore.Bucket(config.LocalBucketName)
}

func TestBucket_PutObject(t *testing.T) {
	err := bucket.PutObject("test/example.txt", strings.NewReader("some text"))
	assert.NoError(t, err)
	object, err := bucket.GetObject("test/example.txt")
	assert.NoError(t, err)
	assert.Equal(t, ".txt", object.Extension())
	assert.Equal(t, "test/example.txt", object.ObjectPath())
	bs, err := io.ReadAll(object)
	assert.NoError(t, err)
	assert.Equal(t, "some text", string(bs))
}

func TestBucket_DeleteObject(t *testing.T) {
	err := bucket.PutObject("test/example.txt", strings.NewReader("some text"))
	assert.NoError(t, err)
	err = bucket.DeleteObject("test/example.txt")
	assert.NoError(t, err)
	_, err = bucket.GetObject("test/example.txt")
	assert.ErrorIs(t, err, oss.ObjectNotFound)
}

func TestBucket_GetObject(t *testing.T) {
	TestBucket_PutObject(t)
}

func TestBucket_GetObjectSize(t *testing.T) {
	err := bucket.PutObject("test/example.txt", strings.NewReader("some text"))
	assert.NoError(t, err)
	size, err := bucket.GetObjectSize("test/example.txt")
	assert.NoError(t, err)
	assert.Equal(t, int64(9), size.Size())
}

func TestBucket_HeadObject(t *testing.T) {
	err := bucket.PutObject("test/example.txt", strings.NewReader("some text"))
	assert.NoError(t, err)
	exist, err := bucket.HeadObject("test/example.txt")
	assert.NoError(t, err)
	assert.True(t, exist)
}

func TestBucket_ListObject(t *testing.T) {
	err := bucket.PutObject("test/example.txt", strings.NewReader("some text"))
	assert.NoError(t, err)
	objects, err := bucket.ListObject("test/")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(objects))
}

func Test_bucket_SignURL(t *testing.T) {
	err := bucket.PutObject("test/example.txt", strings.NewReader("some text"))
	assert.NoError(t, err)
	url, err := bucket.SignURL("test/example.txt", http.MethodGet, time.Second*100)
	log.Println(url)
	assert.NoError(t, err)
	go func() {
		log.Fatalln(http.ListenAndServe(":8080", nil))
	}()

	select {
	case <-time.After(time.Second * 1):
		resp, err := http.Get(url)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		return
	}
}
