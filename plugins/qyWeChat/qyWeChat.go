package qyWeChat

import (
	"encoding/json"
	"github.com/Cgo/kernel/logger"
	"strconv"
	"time"

	//"go/types"
	"io/ioutil"
	"net/http"
	"strings"
)

const(
	//CorpID = "wwe63c0447ee01bb4c"	// 企业的corpid
	//Secret = "ABO-m-dtsHzWSRahCM8x-0ru2DyJOEdAcxD0TQvM9LA"	// 投票调查APP的secret
	//REDIRECT_URI = "http://qywx.cbim.org.cn/votes"	// 投票调查回调地址
	//STATE = "VoteCbim"	// 随机定义子串
	//TongXunLu_Secret = "Ar66ycsOg2rYB79N4rPVGkhpOU6GmbLZq14-JyS2SA0"	// 通讯录同步助手的 secret
)

type typeWxApi struct {
	CorpID string
	Secret string
	REDIRECT_URI string
	STATE string
	TongXunLu_Secret string

	accessToken string	// 当前有效的token
	token_expires int64 // token 失效时间
}

// 创建微信API 实例
func New(CropID, Secret, REDIRECT_URI, STATE, TongXunLu_Secret string)*typeWxApi{
	tmp := &typeWxApi{
		CorpID: CropID,
		Secret: Secret,
		REDIRECT_URI: REDIRECT_URI,
		STATE: STATE,
		TongXunLu_Secret: TongXunLu_Secret}
	token, err := getToken(CropID, Secret)
	if err != nil {
		return nil
	}
	tmp.accessToken = token.Access_token
	tmp.token_expires = token.Expires_in + time.Now().Unix()

	return tmp
}

// 发送一个get请求
func get(url string)([]byte, error){
	var tmp []byte
	req, err := http.Get(url)
	if err != nil {
		return tmp,err
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return tmp,err
	}

	log.Println(string(body))

	return body,nil

}

// 发送一个post请求
func post(url, jsonStr string)([]byte, error){
	req, err := http.Post(url, "application/json;charset=utf-8", strings.NewReader(jsonStr))
	var tmp []byte
	if err != nil {
		return tmp, err
	}

	defer req.Body.Close()
	str, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return tmp, err
	}
	return str,nil
}


// 刷新/获得access token
type accessTokenType struct {
	Errcode int64
	Errmsg string
	Access_token string
	Expires_in int64
}
func getToken(CropID, Secret string)(accessTokenType, error){
	var data accessTokenType
	requestUrl := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + CropID + "&corpsecret=" + Secret
	str, err := get(requestUrl)
	if err != nil {
		return data, err
	}
	json.Unmarshal([]byte(str),&data)
	return data, nil
}
func (this *typeWxApi) ResetToken()(error){

	token, err := getToken(this.CorpID, this.Secret)
	if err != nil {
		return err
	}
	this.accessToken = token.Access_token
	this.token_expires = token.Expires_in + time.Now().Unix()
	return nil

}

// 获得企业全部成员
type entAllUsersType struct {
	Errcode int64
	Errmsg string
	Userlist []struct{
		Userid string
		Name string
		Department []int
	}
}
func (this *typeWxApi) GetAllUsers()(entAllUsersType, error){
	var data entAllUsersType

	// token 请求
	token, err := getToken(this.CorpID, this.TongXunLu_Secret)
	if err != nil {
		return data, err
	}

	// 用户请求
	requestUrl := "https://qyapi.weixin.qq.com/cgi-bin/user/simplelist?access_token="+token.Access_token+"&department_id=1&fetch_child=1"
	str, err := get(requestUrl)
	if err != nil {
		return data, err
	}
	json.Unmarshal(str,&data)

	return data, nil
}


// userid 转 openid
type typeOpenid struct{
	Errcode int64 `json:"errcode"`
	Errmsg string	`json:"errmsg"`
	Openid string	`json:"openid"`
}
func (this *typeWxApi) Userid2Openid(userid string)(typeOpenid, error){

	//  判断token失效
	if time.Now().Unix()>= this.token_expires {
		this.ResetToken()
	}
	var data typeOpenid

	requestUrl := "https://qyapi.weixin.qq.com/cgi-bin/user/convert_to_openid?access_token=" + this.accessToken
	res, err := post(requestUrl, `{"userid":"`+userid+`"}`)
	if err != nil {
		return data, err
	}
	json.Unmarshal(res, &data)

	return data,nil
}


// openid 转 userid
type typeUserid struct{
	Errcode int64 	`json:"errcode"`
	Errmsg string	`json:"errmsg"`
	Userid string	`json:"userid"`
}
func (this *typeWxApi) Openid2Userid(openid string)(typeUserid,error){

	var data typeUserid
	// 兑换通讯录的token
	token, err := getToken(this.CorpID, this.TongXunLu_Secret)

	if err != nil {
		return data,err
	}

	// 兑换userid
	requestUrl := "https://qyapi.weixin.qq.com/cgi-bin/user/convert_to_userid?access_token=" + token.Access_token
	res, err :=post(requestUrl,  `{"openid":"`+openid+`"}`)
	if err != nil {
		return data, err
	}

	json.Unmarshal(res, &data)

	return data,nil
}


// 换取当前登录用户的code
func (this *typeWxApi) GetUserCode(w http.ResponseWriter, r *http.Request){
	reDirectUrl := "https://open.weixin.qq.com/connect/oauth2/authorize?appid="+this.CorpID+"&redirect_uri="+this.REDIRECT_URI+"&response_type=code&scope=snsapi_base&state="+this.STATE+"#wechat_redirect"
	http.Redirect(w, r, reDirectUrl, http.StatusFound)
}


// 获得当前用户的UserId
type userIDtype struct {
	Errcode int64
	Errmsg string
	UserId string
	DeviceId string
}
func (this *typeWxApi) GetUserID(code string)(userIDtype, error){
	var data userIDtype
	// token 超时
	if time.Now().Unix()>= this.token_expires {
		this.ResetToken()
	}
	requestUrl := "https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo?access_token="+this.accessToken+"&code=" + code

	str, err := get(requestUrl)

	if err != nil {
		return data, err
	}

	json.Unmarshal(str, &data)

	return data, nil
}


// 根据Userid 获得用户详细信息
type UserInfoType struct {
	Errcode int64
	Errmsg string
	Userid string
	Name string
	Department []int64
	Order []int64
	Position string
	Mobile string
	Gender int64
	Email string
	Avatar string
	Enable int64
}
func (this *typeWxApi) GetUserInfo(userid string)(UserInfoType, error){
	var data UserInfoType
	// token 超时
	if time.Now().Unix()>= this.token_expires {
		this.ResetToken()
	}
	requestUrl := "https://qyapi.weixin.qq.com/cgi-bin/user/get?access_token="+this.accessToken+"&userid=" + userid
	str, err := get(requestUrl)

	if err != nil {
		return data, err
	}
	json.Unmarshal(str, &data)
	return data, nil
}

// 获得公司部门列表
type departmentType struct {
	Errcode int64	`json:"errcode"`
	Errmsg string	`json:"errmsg"`
	Department []departmentChipType`json:"department"`
}
type departmentChipType struct{
	Id int64	`json:"id"`
	Name string	`json:"name"`
	Parentid int64	`json:"parentid"`
	Order int64	`json:"order"`
}
func (this *typeWxApi) GetDepartment(dpId... int64)(departmentType, error){
	var data departmentType
	// token 超时
	if time.Now().Unix()>= this.token_expires {
		this.ResetToken()
	}
	requestUrl := "https://qyapi.weixin.qq.com/cgi-bin/department/list?access_token="+this.accessToken

	if len(dpId) > 0 {
		requestUrl = requestUrl + "&id=" + strconv.FormatInt(dpId[0], 10)
	}
	str, err := get(requestUrl)

	if err != nil {
		return data, err
	}

	json.Unmarshal(str, &data)

	return data,nil
}