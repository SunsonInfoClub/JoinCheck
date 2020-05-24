package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strconv"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/Tnze/CoolQ-Golang-SDK/cqp"
	"github.com/jinzhu/gorm"
)

//go:generate cqcfg -c .
// cqp: 名称: SunsonCheck
// cqp: 版本: 1.0.0:1
// cqp: 作者: BaiMeow
// cqp: 简介: 台州书生中学信息社团入群审核
func main() { /*此处应当留空*/ }

func init() {
	cqp.AppID = "xyz.baimeow.sunsoncheck" // TODO: 修改为这个插件的ID
	cqp.GroupRequest = onGroupRequest
	cqp.Enable = onEnable
	cqp.Disable = onDisable
	cqp.GroupMsg = onGroupMsg
}

func onEnable() int32 {
	//load conf
	confBytes, err := ioutil.ReadFile(path.Join(cqp.GetAppDir(), "conf.json"))
	if err != nil {
		Error(fmt.Errorf("Fail to Read conf:%v", err))
	}
	if err = json.Unmarshal(confBytes, &conf); err != nil {
		Error(fmt.Errorf("Fail to parse conf:%v", err))
	}
	//init db
	db, err = gorm.Open("sqlite3", path.Join(cqp.GetAppDir(), conf.Database))
	if err != nil {
		Error(fmt.Errorf("Fail to open database:%v", err))
	}
	if !db.HasTable(&Member{}) {
		db.CreateTable(&Member{})
	}
	return 0
}

func onDisable() int32 {
	db.Close()
	return 0
}

var conf struct {
	Group    int64  `json:"Group"`
	Database string `json:"Databese"`
}
var db *gorm.DB

// Member 社团成员
type Member struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	Name      string     `gorm:"primary_key"`
	QQ        int64
	Grade     int
	Class     int
}

func onGroupRequest(subType, sendTime int32, fromGroup, fromQQ int64, msg, responseFlag string) int32 {
	if fromGroup != conf.Group {
		return 0
	}
	mem := Member{}
	db.Where("QQ = ?", strconv.FormatInt(fromQQ, 10)).First(&mem)
	if mem.QQ == 0 {
		Info(fmt.Sprintf("拒绝%d入群", fromQQ))
		cqp.SetGroupAddRequest(responseFlag, subType, deny, "")
		return 1
	}
	Info(fmt.Sprintf("同意%d入群", fromQQ))
	cqp.SetGroupAddRequest(responseFlag, subType, allow, "")
	return 1
}

func onGroupMsg(subType, msgID int32, fromGroup, fromQQ int64, fromAnonymous, msg string, font int32) int32 {
	if fromGroup != conf.Group {
		return 0
	}
	info := cqp.GetGroupMemberInfo(conf.Group, fromQQ, true)
	if info.Level == "1" {
		return 0
	}
	if ok, _ := regexp.MatchString("^/member add [\u4e00-\u9fa5]{2,4} \\d{4} \\d{1,2} \\d{6,12}$", msg); !ok {
		return 0
	}
	mem := Member{}
	_, err := fmt.Sscanf(msg, "/member add %s %d %d %d", &mem.Name, &mem.Grade, &mem.Class, &mem.QQ)
	if err != nil {
		Error(fmt.Errorf("Fail to scan:%v", err))
		return 0
	}
	db.Create(&mem)
	cqp.SendGroupMsg(fromGroup, fmt.Sprintf("成功添加%s的记录", mem.Name))
	return 1
}

// Error 报错
func Error(err error) {
	cqp.AddLog(cqp.Error, "SunSonCheck", err.Error())
	return
}

// Info 一般消息
func Info(s ...interface{}) {
	cqp.AddLog(cqp.Info, "SunSonCheck", fmt.Sprintf("%v", s))
}

const (
	allow = 1 // 允许进群
	deny  = 2 // 拒绝进群
)
