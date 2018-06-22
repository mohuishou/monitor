package main

import (
	"errors"
	"net/http"
	"time"

	"os/exec"

	"fmt"

	"log"

	"bytes"

	"strings"

	"github.com/mohuishou/monitor/config"
	"github.com/robfig/cron"
	gomail "gopkg.in/gomail.v2"
)

var (
	errLog     string
	successLog string
	lastEmail  time.Time
	emailTime  = 3600 // 邮件发送间隔为一个小时
)

func main() {
	c := cron.New()
	conf := config.NewConfig("config.toml")
	log.Println("开始执行！")
	// 监控运行
	c.AddFunc(conf.CronTime, func() {
		run(conf)
	})

	// 每晚10点发送成功邮件
	c.AddFunc("0 0 22 1/1 * ? ", func() {
		err := email(conf.Email, false)
		if err != nil {
			addErrLog("邮件发送失败:" + err.Error())
		}
	})
	c.Start()
	select {}
}

func run(conf config.Config) {
	err := web(conf.URL, time.Duration(conf.Timeout)*time.Second)
	if err == nil {
		addSuccessLog()
		return
	}

	// 记录错误日志
	addErrLog(err.Error())

	// 尝试进行操作
	do(conf.Do)

	// 尝试再次连接
	addErrLog("尝试再次连接")
	err = web(conf.URL, time.Duration(conf.Timeout)*time.Second)
	if err != nil {
		addErrLog(err.Error())
	} else {
		addErrLog("操作执行后系统恢复成功！")
	}

	// 发送错误邮件
	err = email(conf.Email, true)
	if err != nil {
		addErrLog("邮件发送失败:" + err.Error())
	}
}

func addSuccessLog() {
	log.Println("[info]: 系统可以正常访问")
	successLog = fmt.Sprintf(
		"[%s]: 系统可以正常访问 <br /> %s",
		time.Now().Format("2006-01-02 15:04:05"),
		successLog,
	)
}

func addErrLog(str string) {
	log.Println("[err]: ", str)
	errLog = fmt.Sprintf(
		"[%s]: %s <br /> %s",
		time.Now().Format("2006-01-02 15:04:05"),
		str,
		errLog,
	)
}

// 访问网站
func web(url string, timeout time.Duration) error {
	c := http.Client{Timeout: timeout}
	r, err := c.Get(url)
	if err != nil {
		return err
	}
	if r.StatusCode != 200 {
		return errors.New("访问失败:" + r.Status)
	}
	return nil
}

// 操作
func do(ops [][]string) error {
	for _, op := range ops {
		cmd := exec.Command(op[0], op[1:]...)
		var out bytes.Buffer
		cmd.Stdout = &out

		//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
		if err := cmd.Run(); err != nil {
			addErrLog(err.Error())
		} else {
			addErrLog(strings.TrimSpace(out.String()))
		}
	}
	return nil
}

// 发送邮件
func email(conf config.Email, isErr bool) error {
	if time.Now().Unix()-lastEmail.Unix() < int64(emailTime) {
		return errors.New("一小时内已发送错误邮件")
	}
	pre := "【成功】"
	content := successLog
	if isErr {
		content = errLog
		pre = "【失败】"
	}
	m := gomail.NewMessage()
	m.SetHeader("From", conf.From)
	m.SetHeader("To", conf.To)
	m.SetHeader("Subject", pre+conf.Subject)
	m.SetBody("text/html", content)
	d := gomail.NewDialer(conf.SMTP, conf.Port, conf.From, conf.Password)
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	// 邮件发送成功，清空日志信息
	errLog = ""
	successLog = ""
	if isErr {
		lastEmail = time.Now()
	}
	return nil
}
