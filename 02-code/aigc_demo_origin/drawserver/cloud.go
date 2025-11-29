package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"drawserver/config"
	"drawserver/log"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func getTimestamp() uint64 {
	return uint64(time.Now().Unix())
}

type Meta struct {
	Version   string
	Timestamp uint64
	Action    string
	Endpoint  string
	Region    string
	Service   string
	SecretId  string
	SecretKey string
}

type CloudClient struct {
	meta *Meta
}

func NewCloudClient(m *Meta) *CloudClient {
	return &CloudClient{
		meta: m,
	}
}

// private
func (*CloudClient) hmacSha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

func (*CloudClient) sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func (c *CloudClient) generateSign(param string) string {
	var (
		HTTPRequestMethod    = "POST"
		CanonicalURI         = "/"
		CanonicalQueryString = ""
		CanonicalHeaders     = "content-type:application/json; charset=utf-8\nhost:" + c.meta.Endpoint + "\n"
		SignedHeaders        = "content-type;host"
		HashedRequestPayload = c.sha256hex(param) //Lowercase(HexEncode(Hash.SHA256(RequestPayload)))
	)

	CanonicalRequest := HTTPRequestMethod + "\n" +
		CanonicalURI + "\n" +
		CanonicalQueryString + "\n" +
		CanonicalHeaders + "\n" +
		SignedHeaders + "\n" +
		HashedRequestPayload

	date := time.Unix(int64(c.meta.Timestamp), 0).UTC().Format("2006-01-02")
	var (
		Algorithm              = "TC3-HMAC-SHA256" //"HmacSHA256"//"TC3-HMAC-SHA256"
		RequestTimestamp       = strconv.FormatUint(c.meta.Timestamp, 10)
		CredentialScope        = fmt.Sprintf("%s/%s/tc3_request", date, c.meta.Service)
		HashedCanonicalRequest = c.sha256hex(CanonicalRequest) //Lowercase(HexEncode(Hash.SHA256(CanonicalRequest)))。
	)

	var StringToSign = Algorithm + "\n" +
		RequestTimestamp + "\n" +
		CredentialScope + "\n" +
		HashedCanonicalRequest

	secretDate := c.hmacSha256(date, "TC3"+c.meta.SecretKey)
	secretService := c.hmacSha256(c.meta.Service, secretDate)
	secretSigning := c.hmacSha256("tc3_request", secretService)

	signature := hex.EncodeToString([]byte(c.hmacSha256(StringToSign, secretSigning)))

	authorization := Algorithm + " " +
		"Credential=" + c.meta.SecretId + "/" + CredentialScope + ", " +
		"SignedHeaders=" + SignedHeaders + ", " +
		"Signature=" + signature
	return authorization
}

// Request ...
func (c *CloudClient) Request(param string) (string, error) {
	authorization := c.generateSign(param)
	req, err := http.NewRequest("POST", "https://"+c.meta.Endpoint, bytes.NewReader([]byte(param)))
	if err != nil {
		log.Log("Request, error new request, %+v", err)
		return "", err
	}
	req.Header.Set("Host", c.meta.Endpoint)
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-TC-Action", c.meta.Action)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%v", c.meta.Timestamp))
	req.Header.Set("X-TC-Version", c.meta.Version)
	req.Header.Set("X-TC-Region", c.meta.Region)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Log("DetectCelebrity, error response, %+v", err)
		return "", err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", " ")
	if err != nil {
		return string(body), nil
	}
	return prettyJSON.String(), nil
}

// rawCloud 这里是自己封装了云API的v3签名，用于API的调用， 推荐使用官网的SDK
// 图片人脸融合: https://cloud.tencent.com/document/product/670/85618
// 腾讯云V3签名： https://cloud.tencent.com/document/product/867/44978
func rawCloud(modelUrl, imageUrl string) string {
	// 1.创建素材
	meta := &Meta{
		// 创建素材正常情况只能通过控制台使用，QPS默认为1
		// 如果需要通过API的方式调用，有两种方式：
		// 1. 使用本代码来创建素材
		// 2. 联系客服，提供官网SDK
		Action:    "CreateMaterial",
		Version:   "2022-09-27",
		Timestamp: getTimestamp(),
		Endpoint:  "facefusion.tencentcloudapi.com",
		Region:    "ap-guangzhou",
		Service:   "facefusion",
		SecretId:  config.MetaCfg.SecretId,
		SecretKey: config.MetaCfg.SecretKey,
	}
	c := NewCloudClient(meta)

	type Material struct {
		ActivityId   string
		Image        string
		MaterialName string
	}
	type MaterialResponse struct {
		Response struct {
			MaterialStatus int    `json:"MaterialStatus"`
			AuditResult    string `json:"AuditResult"`
			MaterialID     string `json:"MaterialId"`
			RequestID      string `json:"RequestId"`
		} `json:"Response"`
	}
	data, _ := json.Marshal(&Material{
		ActivityId:   "****",
		Image:        download(modelUrl),
		MaterialName: "hello world",
	})
	resp, err := c.Request(string(data))
	if err != nil {
		panic(err)
	}

	materialResp := &MaterialResponse{}
	json.Unmarshal([]byte(resp), &materialResp)

	time.Sleep(1 * time.Second)

	// 2.图片人脸融合
	meta.Action = "FuseFace"
	meta.Timestamp = getTimestamp()

	type MergeInfo struct {
		Url string
	}
	type FaceFusion struct {
		ProjectId  string `json:"ProjectId"`
		ModelId    string
		RspImgType string
		MergeInfos []MergeInfo
	}
	type FaceFusionResponse struct {
		Response struct {
			FusedImage      string        `json:"FusedImage"`
			RequestID       string        `json:"RequestId"`
			ReviewResultSet []interface{} `json:"ReviewResultSet"`
		} `json:"Response"`
	}
	data, _ = json.Marshal(&FaceFusion{
		ProjectId:  "****",
		ModelId:    materialResp.Response.MaterialID,
		RspImgType: "url", // 返回图片形式为URL
		MergeInfos: []MergeInfo{
			{
				Url: imageUrl,
			},
		},
	})
	resp, err = c.Request(string(data))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", resp)
	fusionResp := &FaceFusionResponse{}
	json.Unmarshal([]byte(resp), &fusionResp)
	return fusionResp.Response.FusedImage
}

func download(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(body)
}
