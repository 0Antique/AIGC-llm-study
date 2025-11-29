package main

import (
	"context"
	"drawserver/cos"
	"drawserver/log"
	"encoding/json"
	"fmt"
	cossdk "github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"math/rand"
	"strings"
)

var (
	JOB_QUEUE_PUSH   = "aigc"
	JOB_QUEUE_RESULT = "aigcresult"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// randSeq
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// GenJobId 生成jobID， 用于异步任务的查询
func GenJobId() string {
	return randSeq(32)
}

// pushJob
func pushJob(jobId, data string) {
	_, err := cos.PutObject(context.Background(), fmt.Sprintf("%s/%s", JOB_QUEUE_PUSH, jobId), []byte(data))
	if err != nil {
		panic(err)
	}
}

type JobInfo struct {
	JobId string `json:"job_id"`
	Request
}

// getNameByPath: input: a/b/1.jpg | output: 1.jpg
func getNameByPath(path string) string {
	slc := strings.Split(path, "/")
	filename := slc[len(slc)-1]
	return filename
}

// sumJob 获取任务总数
func sumJob() int64 {
	res, _, err := cos.GetInstance().Bucket.Get(context.Background(), &cossdk.BucketGetOptions{
		Prefix:       JOB_QUEUE_PUSH + "/",
		Delimiter:    "",
		EncodingType: "",
		Marker:       "",
		MaxKeys:      10000,
	})
	if err != nil {
		return 0
	}
	return int64(len(res.Contents))
}

// pullJob 拉取任务
// 遍历cos任务目录，筛选新任务
func pullJob() *JobInfo {
	res, _, err := cos.GetInstance().Bucket.Get(context.Background(), &cossdk.BucketGetOptions{
		Prefix:       JOB_QUEUE_PUSH,
		Delimiter:    "",
		EncodingType: "",
		Marker:       "",
		MaxKeys:      10000,
	})
	if err != nil {
		return nil
	}
	var jobId string
	for _, v := range res.Contents {
		if !cos.ObjectExist(fmt.Sprintf("%s/%s", JOB_QUEUE_RESULT, getNameByPath(v.Key))) {
			jobId = v.Key
			break
		}
	}
	if len(jobId) == 0 {
		return nil
	}
	jobId = getNameByPath(jobId)
	log.Log("new job %s", jobId)
	resp, err := cos.GetInstance().Object.Get(context.Background(),
		fmt.Sprintf("%s/%s", JOB_QUEUE_PUSH, jobId),
		&cossdk.ObjectGetOptions{})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	job := &JobInfo{
		JobId: jobId,
	}
	err = json.Unmarshal(body, &job)
	if err != nil {
		return nil
	}
	return job
}

// pullResult 查询任务结果
func pullResult(jobId string) *Response {
	resp, err := cos.GetInstance().Object.Get(context.Background(),
		fmt.Sprintf("%s/%s", JOB_QUEUE_RESULT, jobId),
		&cossdk.ObjectGetOptions{})
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	rsp := &Response{}
	json.Unmarshal(body, &rsp)
	return rsp
}

// pushResult
// step4: 上报结果
func pushResult(jobId, result string) {
	_, err := cos.PutObject(context.Background(), fmt.Sprintf("%s/%s", JOB_QUEUE_RESULT, jobId), []byte(result))
	if err != nil {
		panic(err)
	}
}
