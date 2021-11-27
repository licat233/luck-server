package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	// 	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var LineMaxcount = 1
var IpMaxcount = 5

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type RespData struct {
	PrezisID int8   `json:"prezis_id"`
	WinCode  string `json:"win_code"`
}

type ReqData struct {
	LineID string `json:"line_id"`
}

type VerifyResp struct {
	Token string `json:"token"`
	Count int    `json:"count"` //剩余次数
}

type SearchReq struct {
	WinCode string `json:"win_code"`
}

type SearchResp struct {
	Prizename  string `json:"prize_name"`  // 礼品名称
	Prizeimage string `json:"prize_image"` // 礼品图片
	LineId     string `json:"line_id"`
}

//自定义秘钥
var jwtkey = []byte("xianggoumaoyi")

type Claims struct {
	LineId   string
	RemoteIP string
	jwt.StandardClaims
}

type Prize struct {
	Id     int    // 对应的前端Index
	Name   string // 礼品名称
	Image  string // 礼品图片
	Chance int32  // 对应的几率 值越大 获取到的几率越小
	Win    bool   //中獎了
}

var Prizes = []Prize{
	{
		Id:     0,
		Name:   "泰國身體乳",
		Chance: 100,
		Image:  "https://img.alicdn.com/imgextra/i4/917298378/O1CN01dLJu6z2BlAxjK7iEB_!!917298378.png",
		Win:    true,
	}, {
		Id:     1,
		Name:   "1000元折扣券",
		Chance: 100,
		Image:  "https://img.alicdn.com/imgextra/i4/917298378/O1CN01Lf5sMT2BlAxjK76oR_!!917298378.png",
		Win:    true,
	}, {
		Id:     2,
		Name:   "植題沐浴油",
		Chance: 100,
		Image:  "https://img.alicdn.com/imgextra/i4/917298378/O1CN01fBhkcG2BlAxYPbeOa_!!917298378.png",
		Win:    true,
	}, {
		Id:     3,
		Name:   "抗皺除斑套裝",
		Chance: 100,
		Image:  "https://img.alicdn.com/imgextra/i4/917298378/O1CN01D3CXix2BlAxgrV2jH_!!917298378.png",
		Win:    true,
	}, {
		Id:     4,
		Name:   "免單",
		Chance: 100,
		Image:  "https://img.alicdn.com/imgextra/i3/917298378/O1CN01nqMqdo2BlAxkNwLRW_!!917298378.png",
		Win:    true,
	}, {
		Id:     5,
		Name:   "神秘大禮包",
		Chance: 100,
		Image:  "https://img.alicdn.com/imgextra/i3/917298378/O1CN01a1e5Kc2BlAxabAHJB_!!917298378.png",
		Win:    true,
	}, {
		Id:     6,
		Name:   "奧迪口紅",
		Chance: 100,
		Image:  "https://img.alicdn.com/imgextra/i2/917298378/O1CN01gdpx1Z2BlAxl2xnkY_!!917298378.png",
		Win:    true,
	}, {
		Id:     7,
		Name:   "謝謝參與",
		Chance: 100,
		Image:  "https://img.alicdn.com/imgextra/i1/917298378/O1CN01thznsy2BlAxchoaCM_!!917298378.png",
		Win:    false,
	},
}

// 声明一个全局的rdb变量
var rdb *redis.Client

// 初始化连接
func initClient() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err = rdb.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if initClient() != nil {
		log.Fatalln("redis conn failed")
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 	r.Use(cors.New(cors.Config{
	// 		AllowOrigins: []string{"*"},
	// 		AllowMethods: []string{"*"},
	// 		AllowHeaders: []string{"*"},
	// 	}))

	//静态文件服务
	r.StaticFS("/luck/static", http.Dir("./client/static"))
	// 	r.Static("/luck/static", "./client/static")
	r.StaticFile("/luck/", "./client/index.html")

	r.POST("/luck/verify", verify)
	r.POST("/luck/goodluck", goodluck)
	r.POST("/luck/prizes", getPrizes)
	r.POST("/luck/search", searchWincode)

	r.Run(":65456")
}

func searchWincode(c *gin.Context) {
	resp := Response{
		Success: false,
		Message: "request error",
		Data:    nil,
	}
	json := SearchReq{}
	c.BindJSON(&json)
	wincode := strings.TrimSpace(json.WinCode)
	if len(wincode) == 0 {
		c.JSON(400, resp)
		return
	}
	wincodeRes, err := rdb.Get(wincode).Result()
	if err == redis.Nil {
		resp.Message = "wincode not found"
		c.JSON(200, resp)
		return
	} else if err != nil {
		resp.Message = "redis server get id error"
		//redis出错了
		c.JSON(500, resp)
		return
	}
	pdslice := strings.Split(wincodeRes, "$$")
	lineid := pdslice[0]
	prizeid, err := strconv.Atoi(pdslice[1])
	if err != nil {
		resp.Message = "prizeid not 'int'"
		c.JSON(500, resp)
		return
	}

	result := SearchResp{
		Prizename:  "",
		Prizeimage: "",
		LineId:     lineid,
	}
	for _, v := range Prizes {
		if v.Id == prizeid {
			result.Prizename = v.Name
			result.Prizeimage = v.Image
			resp.Success = true
			resp.Data = result
			c.JSON(200, resp)
			return
		}
	}
	resp.Message = "not found prize"
	c.JSON(500, resp)
	return
}

func getPrizes(c *gin.Context) {
	c.JSON(200, Response{
		Success: true,
		Message: "请求成功",
		Data:    Prizes,
	})
}

func luckPrize() Prize {
	sort.Slice(Prizes[:], func(i, j int) bool {
		return Prizes[i].Chance < Prizes[j].Chance
	})
	var allprob int32
	for _, v := range Prizes {
		allprob += v.Chance
	}
	result := Prize{
		Id:     1000,
		Name:   "",
		Chance: 0,
		Win:    false,
	}
	rand.Seed(time.Now().UnixNano())
	random := rand.Int31n(allprob)
	for _, v := range Prizes {
		if random <= v.Chance {
			return v
		}
		random -= v.Chance
	}
	return result
}

//redis存儲已經抽取的次數

func verify(c *gin.Context) {
	resp := Response{
		Success: false,
		Message: "request error",
		Data:    nil,
	}
	json := ReqData{}
	c.BindJSON(&json)
	lineid := strings.TrimSpace(json.LineID)
	if len(lineid) == 0 {
		c.JSON(400, resp)
		return
	}
	ip := GetRequestIP(c)

	//初始化
	var linenum, ipnum int
	//檢查lineid
	lineidcount, err := rdb.Get(lineid).Result()
	//lineid 不存在
	if err == redis.Nil {
		//设置已抽次数
		err := rdb.Set(lineid, 0, 0).Err()
		if err != nil {
			resp.Message = "redis server set id error"
			//redis出错了
			c.JSON(500, resp)
			return
		}
	} else if err != nil {
		resp.Message = "redis server get id error"
		//redis出错了
		c.JSON(500, resp)
		return
	} else {
		count, err := strconv.Atoi(lineidcount)
		if err != nil {
			resp.Message = "your lineid error"
			c.JSON(400, resp)
			return
		}
		linenum = count

	}

	if ip != "127.0.0.1" {
		//檢查 ip
		ipcount, err := rdb.Get(ip).Result()
		//lineid 不存在
		if err == redis.Nil {
			//设置已抽次数
			err = rdb.Set(ip, 0, 0).Err()
			if err != nil {
				resp.Message = "redis server set ip error"
				//redis出错了
				c.JSON(500, resp)
				return
			}
		} else if err != nil {
			resp.Message = "redis server get ip error"
			//redis出错了
			c.JSON(500, resp)
			return
		} else {
			count, err := strconv.Atoi(ipcount)
			if err != nil {
				resp.Message = "your lineid error"
				c.JSON(400, resp)
				return
			}
			ipnum = count
		}
	}

	//設置token
	jwttoken, err := settingToken(lineid, ip)
	if err != nil {
		resp.Message = "set token error"
		c.JSON(500, resp)
		return
	}

	data := VerifyResp{
		Token: jwttoken,
		Count: linenum,
	}
	resp.Success = true
	//lineid已抽次数>=2,ip已抽次数>=5
	if linenum >= LineMaxcount || ipnum >= IpMaxcount {
		data.Count = 0
		resp.Message = "你的抽獎次數已用完"
		resp.Data = data
		c.JSON(200, resp)
		return
	}

	//成功
	data.Count = LineMaxcount - linenum
	resp.Message = "祝你好运"
	resp.Data = data
	c.JSON(200, resp)
	return
}

func goodluck(c *gin.Context) {
	resp := Response{
		Success: false,
		Message: "request error",
		Data:    nil,
	}
	cla, err := gettingToken(c)
	if err != nil {
		resp.Message = "違規請求"
		c.JSON(401, resp)
		return
	}
	ip := GetRequestIP(c)
	if ip != cla.RemoteIP {
		resp.Message = "違規請求"
		c.JSON(401, resp)
		return
	}
	luckp := luckPrize()
	if luckp.Id == 1000 {
		resp.Message = "server error"
		c.JSON(500, resp)
		return
	}
	//lineid
	lineid := cla.LineId
	lineidcount, err := rdb.Get(lineid).Result()
	if err == redis.Nil {
		resp.Message = "lineid not found"
		c.JSON(400, resp)
		return
	} else if err != nil {
		resp.Message = "redis server get id error"
		//redis出错了
		c.JSON(500, resp)
		return
	}
	count, err := strconv.Atoi(lineidcount)
	if err != nil {
		resp.Message = "your lineid error"
		c.JSON(400, resp)
		return
	}
	if count >= LineMaxcount {
		resp.Success = true
		resp.Message = "你的次数已经用完"
		c.JSON(200, resp)
		return
	}
	var countval int
	if ip != "127.0.0.1" {
		//ip 处理
		ipcount, err := rdb.Get(ip).Result()
		if err == redis.Nil {
			resp.Message = "ip not found"
			c.JSON(400, resp)
			return
		} else if err != nil {
			resp.Message = "redis server get ip error"
			//redis出错了
			c.JSON(500, resp)
			return
		}
		countval, err = strconv.Atoi(ipcount)
		if err != nil {
			resp.Message = "your lineid error"
			c.JSON(400, resp)
			return
		}
		if countval > IpMaxcount { //考慮有的人可能有多個LINE號
			resp.Success = true
			resp.Message = "你的次数已经用完"
			c.JSON(200, resp)
			return
		}
	}

	//设置已抽次数
	err = rdb.Set(lineid, count+1, 0).Err()
	if err != nil {
		resp.Message = "redis server set id error"
		//redis出错了
		c.JSON(500, resp)
		return
	}
	err = rdb.Set(ip, countval+1, 0).Err()
	if err != nil {
		resp.Message = "redis server set ip error"
		//redis出错了
		c.JSON(500, resp)
		return
	}

	wincode, err := generateWinCode(lineid, luckp.Id, luckp.Win)
	if err != nil {
		resp.Message = "redis server set wincode error"
		//redis出错了
		c.JSON(500, resp)
		return
	}
	data := RespData{
		PrezisID: int8(luckp.Id), //这里可做概率抽取
		WinCode:  wincode,
	}
	resp.Success = true
	resp.Message = "successfull"
	resp.Data = data
	c.JSON(200, resp)
}

func generateWinCode(lineid string, prizeid int, w bool) (string, error) {
	if w == false {
		return "", nil
	}
	data := fmt.Sprintf("%s$$%d", lineid, prizeid)
	md5str := Md5Encrypt(data)
	err := rdb.Set(md5str, data, 0).Err()
	if err != nil {
		return "", errors.New("redis server set ip error")
	}
	return md5str, nil
}

func Md5Encrypt(data string) string {
	md5Ctx := md5.New()                            //md5 init
	md5Ctx.Write([]byte(data))                     //md5 updata
	cipherStr := md5Ctx.Sum(nil)                   //md5 final
	encryptedData := hex.EncodeToString(cipherStr) //hex_digest
	return encryptedData
}

//颁发token
func settingToken(line, ip string) (string, error) {
	expireTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		LineId:   line,
		RemoteIP: ip,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(), //过期时间
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// fmt.Println(token)
	return token.SignedString(jwtkey)
}

//解析token
func gettingToken(ctx *gin.Context) (*Claims, error) {
	tokenString := ctx.GetHeader("Authorization")
	//vcalidate token formate
	if tokenString == "" {
		return nil, errors.New("缺少token")
	}

	token, claims, err := ParseToken(tokenString)
	if err != nil || !token.Valid {
		return nil, errors.New("权限不足")
	}
	return claims, nil
}

func ParseToken(tokenString string) (*jwt.Token, *Claims, error) {
	Claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, Claims, func(token *jwt.Token) (i interface{}, err error) {
		return jwtkey, nil
	})
	return token, Claims, err
}

//获取ip
func GetRequestIP(c *gin.Context) string {
	reqIP := c.ClientIP()
	if reqIP == "::1" {
		reqIP = "127.0.0.1"
	}
	return reqIP
}