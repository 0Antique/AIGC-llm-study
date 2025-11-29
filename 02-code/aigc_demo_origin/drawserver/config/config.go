package config

import (
	"drawserver/cos"
)

// CosCfg COS配置， 用于任务管理， 具体配置参考cos接入指引
// cos文档：https://cloud.tencent.com/document/product/436/31215
var CosCfg = &cos.Config{
	Secret: cos.Secret{
		SecretId:  "****",
		SecretKey: "****",
	},
	Bucket: cos.Bucket{
		Bucket: "****",
		Region: "****",
	},
}

// MetaCfg 融合API
var MetaCfg = &struct {
	SecretId  string
	SecretKey string
}{
	SecretId:  "****",
	SecretKey: "****",
}
