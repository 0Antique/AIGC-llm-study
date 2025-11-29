package cos

import (
	"bytes"
	"context"
	"drawserver/log"
	"errors"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
)

var (
	ErrorClientNull = errors.New("client: not found instance")
)

var defaultCosClient atomic.Value

// Secret ...
type Secret struct {
	SecretId   string `json:"secret_id"`
	SecretKey  string `json:"secret_key"`
	Encryption string `json:"encryption"`
}

// Bucket ...
type Bucket struct {
	Bucket string `json:"bucket"`
	Region string `json:"region"`
}

// Config ...
type Config struct {
	Secret Secret `json:"cos_secret_config"`
	Bucket Bucket `json:"cos_bucket_config"`
}

// SetConfig ...
func SetConfig(cfg *Config) {
	u, err := url.Parse(fmt.Sprintf("https://%v.cos.%v.myqcloud.com", cfg.Bucket.Bucket, cfg.Bucket.Region))
	if err != nil {
		log.Log("parse url failed, %+v", err)
		return
	}
	b := &cos.BaseURL{BucketURL: u}
	client := http.Client{}
	client.Transport = &cos.AuthorizationTransport{
		SecretID:  cfg.Secret.SecretId,
		SecretKey: cfg.Secret.SecretKey,
	}
	defaultCosClient.Store(cos.NewClient(b, &client))
}

// OptionFunc ...
type OptionFunc func(opt *cos.ObjectPutOptions)

// WithContentType ...
func WithContentType(val string) OptionFunc {
	return func(opt *cos.ObjectPutOptions) {
		opt.ContentType = val
	}
}

// /client
func client() *cos.Client {
	v, ok := defaultCosClient.Load().(*cos.Client)
	if !ok {
		return nil
	}
	return v
}

// GetBuckets ...
func GetBuckets(ctx context.Context) ([]cos.Bucket, error) {
	c := client()
	if c == nil {
		log.Log("get client is nil")
		return nil, ErrorClientNull
	}
	result, _, err := c.Service.Get(ctx)
	if err != nil {
		log.Log("Put COS, failed %+v", err)
		return nil, err
	}
	return result.Buckets, nil
}

// GetObjects ...
func GetObjects(ctx context.Context) ([]cos.Object, error) {
	c := client()
	if c == nil {
		log.Log("get client is nil")
		return nil, ErrorClientNull
	}
	result, _, err := c.Bucket.Get(ctx, &cos.BucketGetOptions{})
	if err != nil {
		log.Log("get COS, failed %+v", err)
		return nil, err
	}
	return result.Contents, err
}

// PutObject ...
func PutObject(ctx context.Context, filepath string, data []byte, opts ...OptionFunc) (string, error) {
	c := client()
	if c == nil {
		log.Log("get client is nil")
		return "", ErrorClientNull
	}

	opt := &cos.ObjectPutOptions{
		ACLHeaderOptions:       &cos.ACLHeaderOptions{},
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{},
	}
	for _, o := range opts {
		o(opt)
	}

	buff := bytes.NewBuffer(data)
	_, err := c.Object.Put(ctx, filepath, buff, opt)
	if err != nil {
		log.Log("Put COS, failed %+v", err)
		return "", err
	}
	ret := c.BaseURL.BucketURL.String() + EncodeURI(filepath)
	return ret, nil
}

// ObjectExist ...
func ObjectExist(path string) bool {
	_, err := client().Object.Head(context.Background(), path, &cos.ObjectHeadOptions{})
	if err == nil {
		return true // exist
	}
	if cos.IsNotFoundError(err) {
		return false
	}
	return false
}

// GetInstance ...
func GetInstance() *cos.Client {
	return client()
}

// EncodeURIComponent like same function in javascript
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/encodeURIComponent
// http://www.ecma-international.org/ecma-262/6.0/#sec-uri-syntax-and-semantics
// From TencentCloud COS SDK
func EncodeURIComponent(s string, excluded ...[]byte) string {
	var b bytes.Buffer
	written := 0

	for i, n := 0, len(s); i < n; i++ {
		c := s[i]

		switch c {
		case '-', '_', '.', '!', '~', '*', '\'', '(', ')':
			continue
		default:
			// Unreserved according to RFC 3986 sec 2.3
			if isNumOrLetter(c) {
				continue
			}

			if len(excluded) > 0 && isExcluded(c, excluded...) {
				continue
			}
		}

		b.WriteString(s[written:i])
		fmt.Fprintf(&b, "%%%02X", c)
		written = i + 1
	}

	if written == 0 {
		return s
	}
	b.WriteString(s[written:])
	return b.String()
}

func isNumOrLetter(c uint8) (ret bool) {
	if 'a' <= c && c <= 'z' {
		return true
	}
	if 'A' <= c && c <= 'Z' {
		return true
	}
	if '0' <= c && c <= '9' {
		return true
	}

	return false
}

func isExcluded(c uint8, excluded ...[]byte) (ret bool) {
	for _, ch := range excluded[0] {
		if ch == c {
			return true
		}
	}

	return
}

// EncodeURI ...
func EncodeURI(uri string) string {
	uri = "/" + strings.TrimLeft(uri, "/")
	uri = EncodeURIComponent(uri, []byte("/"))
	return uri
}
