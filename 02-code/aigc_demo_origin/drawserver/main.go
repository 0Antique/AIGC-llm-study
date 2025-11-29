package main

import (
	"context"
	"drawserver/config"
	"drawserver/cos"
	"drawserver/log"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Request struct {
	SessionId string `json:"session_id"`
	JobId     string `json:"job_id"`
	Prompt    string `json:"prompt"`
	ModelUrl  string `json:"model_url"`
	ImageUrl  string `json:"image_url"`
}

type Response struct {
	SessionId  string `json:"session_id"`
	JobId      string `json:"job_id"`
	JobStatus  string `json:"job_status"`
	CostTime   int64  `json:"cost_time"`
	ResultUrl  string `json:"result_url"`
	TotalCnt   int64  `json:"total_cnt"`
	RemainTime int64  `json:"remain_time"`
	QueueLen   int64  `json:"queue_len"`
}

func Init() {
	cos.SetConfig(config.CosCfg)
}

// GetPresignedURL 获取cos预签名地址
func GetPresignedURL(ctx context.Context, path string) (string, error) {
	uri, err := cos.GetInstance().Object.GetPresignedURL(ctx, http.MethodGet, path,
		config.CosCfg.Secret.SecretId, config.CosCfg.Secret.SecretKey, time.Hour, nil)
	if err != nil {
		log.Log("GetPresignedURL, error: %+v", err)
		return "", err
	}
	return uri.String(), nil
}

// createJobHandler 直接提交任务
// Deprecated: 测试用，当前使用cos管理任务
func createJobHandler(writer http.ResponseWriter, request *http.Request) {
	begin := time.Now()
	body, err := io.ReadAll(request.Body)
	req := &Request{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}
	log.Log("got a job, %+v", req)
	jobId := GenJobId()
	cmd := exec.Command("sh", "start.sh", req.Prompt, jobId)

	err = cmd.Run()
	if err != nil {
		fmt.Println("Execute Command failed:" + err.Error())
		return
	}

	result, err := os.ReadFile(fmt.Sprintf("output/%s.png", jobId))
	if err != nil {
		panic(err)
	}
	_, err = cos.PutObject(context.Background(), fmt.Sprintf("aidraw/%s.png", jobId), result)
	if err != nil {
		panic(err)
	}

	// cos预签名地址
	presignUrl, _ := GetPresignedURL(context.Background(), fmt.Sprintf("aidraw/%s.png", jobId))
	resp := &Response{
		SessionId: req.SessionId,
		JobId:     jobId,
		JobStatus: "FINISNED",
		CostTime:  time.Since(begin).Milliseconds(),
		ResultUrl: presignUrl,
	}
	log.Log("job finished, %+v", resp)
	data, _ := json.Marshal(resp)
	writer.Write(data)
}

// submitJobHandler 提交任务
// step1: 提交关键词
func submitJobHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	req := &Request{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}
	log.Log("got a submit request, %+v", req)
	jobId := GenJobId()
	// step2： 新建任务
	pushJob(jobId, string(body))
	resp := &Response{
		SessionId:  req.SessionId,
		JobId:      jobId,
		TotalCnt:   sumJob(),
		RemainTime: 38 + rand.Int63n(10),
		QueueLen:   1,
	}
	data, _ := json.Marshal(resp)
	writer.Write(data)
}

// describeJobHandler 查询任务
// step6: 查询结果
func describeJobHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	req := &Request{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}
	log.Log("got a query request, %+v", req.JobId)
	var ret *Response
	ret = pullResult(req.JobId)
	if ret == nil {
		ret = &Response{
			SessionId: req.SessionId,
			JobId:     req.JobId,
			JobStatus: "RUNNING",
		}
	}
	data, _ := json.Marshal(ret)
	writer.Write(data)
}

// faceFusionHandler ...
// step5: 融合结果
func faceFusionHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	req := &Request{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		panic(err)
	}

	ret := &Response{
		SessionId: req.SessionId,
		ResultUrl: rawCloud(req.ModelUrl, req.ImageUrl),
	}
	data, _ := json.Marshal(ret)
	writer.Write(data)
}

var (
	flagAddr   = flag.String("a", "0.0.0.0:8000", "")
	flagWorker = flag.Bool("w", false, "")
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	Init()

	// true， 则开启AI画画服务
	// false，则只提供小程序服务端功能
	if *flagWorker {
		go eventLoop()
	}

	// 小程序服务端
	// 提交任务
	http.HandleFunc("/frontend/create", submitJobHandler)
	// 查询任务
	http.HandleFunc("/frontend/query", describeJobHandler)
	// 人脸融合
	http.HandleFunc("/frontend/fusion", faceFusionHandler)

	// AI画画服务端
	// 文字生成图片
	http.HandleFunc("/aigc/create", createJobHandler)
	_ = http.ListenAndServe(*flagAddr, nil)
}
