package dto

type CapturePicReq struct {
	//截图信息 map的key是文件名，value是url
	PicInfo map[string]string `json:"pic_info"`
	//oss上文件的存放路径
	PicDir string `json:"pic_dir"`
	//oss的accessKey
	AccessKeyId string `json:"access_key_id"`
	//oss的accessSecret
	AccessKeySecret string `json:"access_key_secret"`
	//oss的endpoint
	OssUrl string `json:"oss_url"`
	//oss的bucket名
	BucketName string `json:"bucket_name"`
}
