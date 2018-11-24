package srun

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
)

const (
	challengeUrl = "http://10.0.0.55/cgi-bin/get_challenge"
	portalUrl    = "http://10.0.0.55/cgi-bin/srun_portal"
	succeedUrl = "http://10.0.0.55/srun_portal_pc_succeed.php"
	url          = "http://10.0.0.55"
)

func genInfo(q QLogin, token string) string  {
	x_encode_json := map[string]interface{} {
		"username": q.Username,
		"password": q.Password,
		"ip": q.Ip,
		"acid": q.Acid,
		"enc_ver": "srun_bx1",
	}
	x_encode_raw, err := json.Marshal(x_encode_json);
	if err != nil {
		logs.Error(err)
		return ""
	}
	xen := string(x_encode_raw)
	x_encode_res := X_encode(xen, token)
	dict_key := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	dict_val := "LVoJPiCN2R8G90yg+hmFHuacZ1OWMnrsSTXkYpUq/3dlbfKwv6xztjI7DeBE45QA="
	dict := map[string]string{}
	for idx, v := range dict_key {
		dict[string(v)] = dict_val[idx:idx+1]
	}
	b64_arr := []byte{}
	for _,c := range x_encode_res {
		b64_arr = append(b64_arr, byte(c))
	}
	b64_res := base64.StdEncoding.EncodeToString(b64_arr)
	target := ""
	for _, s := range b64_res {
		target += dict[string(s)]
	}
	return "{SRBX1}" + target
}


func Login(username, password string) (token, ip string) {
	qLogin := NewQLogin(username, password)

	//	get token
	qc := NewQChallenge(username)
	rc := RChallenge{}
	if err := getJson(challengeUrl,qc,  &rc); err != nil {
		logs.Error(err)
		return
	}
	token = rc.Challenge
	ip = rc.ClientIp

	qLogin.Ip = ip
	info := genInfo(qLogin, token)
	qLogin.Info = info
	qLogin.Password = pwdHmd5("", token)
	qLogin.Chksum = Checksum(qLogin, token)

	ra := RAction{}
	err := getJson(portalUrl, qLogin, &ra)
	if err != nil {
		logs.Error(err)
		return
	}
	if ra.Res != "ok" {
		fmt.Println("登录失败:", ra.Res)
		fmt.Println("msg:", ra.Error)
		logs.Error(ra)
		return
	}

	fmt.Println("登录成功!")
	fmt.Println("ip:", ra.ClientIp)

	qs := QInfo{
		Acid:qLogin.Acid,
		Username:qLogin.Username,
		ClientIp:ra.ClientIp,
		AccessToken:token,
	}

	parseHtml(succeedUrl, qs)
	return
}

func Info(username, token, ip string)  {
	qs := QInfo{
		Acid: 1,
		Username: username,
		ClientIp: ip,
		AccessToken:token,
	}
	parseHtml(succeedUrl, qs)
	return
}

func Logout(username string)  {
	q := QLogout{
		Action:"logout",
		Username:username,
		Acid: 1,
		Ip:"",
	}
	ra := RAction{}
	err := getJson(portalUrl, q, &ra)
	if err != nil {
		logs.Error(err)
		return
	}
	if ra.Error == "ok" {
		fmt.Println("下线成功！")
	}else {
		logs.Error(ra)
	}
}