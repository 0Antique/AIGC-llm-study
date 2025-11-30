package main

import (
	"context"
	"drawserver/cos"
	"drawserver/log"
	"encoding/json"
	"fmt"
	"os"
	// "os/exec"
	"time"
	"encoding/base64"
    "io/ioutil"
	// "bytes"
    // "crypto/hmac"
    // "crypto/sha1"
    // "encoding/base64"
    // "sort"
    // "strconv"
    "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
    "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
    "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
    aiart "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/aiart/v20221229"
)

/* start.sh
#!/bin/bash
python test.py "$1" output/$2 >> script_draw.log
*/

/* test.py
# https://github.com/CompVis/stable-diffusion
from torch import autocast
from diffusers import StableDiffusionPipeline
import sys

pipe = StableDiffusionPipeline.from_pretrained(
	"stable-diffusion-v1-4",
	use_auth_token=True
).to("cuda")

prompt = "a photo of an astronaut riding a horse on mars"
prompt = sys.argv[1]
print("values: {} {}".format(sys.argv[1], sys.argv[2]))

with autocast("cuda"):
	image = pipe(prompt, num_inference_steps=100).images[0]
	image.save(sys.argv[2] + ".png")
*/


// AI画画： 文生图
// 生成图片 -> 上传结果
func run(req *JobInfo) {
	begin := time.Now()

	log.Log("got a job, %+v", req)
	jobId := req.JobId
	// 生成图片
	// cmd := exec.Command("sh", "start.sh", req.Prompt, jobId)

	// err := cmd.Run()
	// if err != nil {
	// 	fmt.Println("Execute Command failed:" + err.Error())
	// 	return
	// }

//设置腾讯云的密钥和ID
	credential := common.NewCredential(     
            "",
            "",
        )
    // 实例化一个client选项，可选的，没有特殊需求可以跳过
    cpf := profile.NewClientProfile()
    cpf.HttpProfile.Endpoint = "aiart.tencentcloudapi.com"
    // 实例化要请求产品的client对象,clientProfile是可选的
    client, _ := aiart.NewClient(credential, "ap-上海", cpf)

    // 实例化一个请求对象,每个接口都会对应一个request对象
    request := aiart.NewTextToImageRequest()
    // 图片的分辨率
    request.ResultConfig = &aiart.ResultConfig {
                Resolution: common.StringPtr("1024:768"),
        }
    // 返回的图片格式 
    request.RspImgType = common.StringPtr("base64")
    request.Prompt = common.StringPtr(req.Prompt)

    // 返回的resp是一个TextToImageResponse的实例，与请求对象对应
    response, err := client.TextToImage(request)
    if _, ok := err.(*errors.TencentCloudSDKError); ok {
            fmt.Printf("An API error has returned: %s", err)
            return
    }
    if err != nil {
            panic(err)
    }
    // 输出json格式的字符串回包
    fmt.Printf("%s", response.ToJsonString())



    // resultImage := response.ResultImage
    // requestId := resp.RequestId

	dt := json.Unmarshal([]byte(response.ToJsonString()), &data)
	if dt == nil {
		fmt.Println("JSON unmarshal error:", err)
	}

	resultImage, exists := data["Response"]["ResultImage"]
	if !exists {
		fmt.Println("ResultImage not found in JSON.")
	}

	fmt.Println("ResultImage:", resultImage)


    decoded, err := base64.StdEncoding.DecodeString(resultImage)
    err = saveImage(decoded, fmt.Sprintf("output/%s.png", jobId))
    if err != nil {
		panic(err)
	}

	result, err := os.ReadFile(fmt.Sprintf("output/%s.png", jobId))
	if err != nil {
		panic(err)
	}
	

	// 上传图片内容到cos
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

	// step4: 上传图片结果
	pushResult(jobId, string(data))
}

// eventLoop 定时拉取任务
func eventLoop() {
	for {
		time.Sleep(1 * time.Second)
		// step3： 拉取任务
		req := pullJob()
		if req == nil {
			continue
		}
		run(req)
	}
}

// 保存图片数据到文件
func saveImage(data []byte, filename string) error {
    return ioutil.WriteFile(filename, data, 0644)
}


var data map[string]map[string]string

