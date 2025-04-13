package alarm

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gwaypg/log-server/module/db"
	"github.com/gwaypg/log-server/module/mail"

	"github.com/gwaylib/database"
	"github.com/gwaylib/errors"
	"github.com/gwaylib/log"
)

type MailServer struct {
	SmtpHost string `json:"stmp_host"`
	SmtpPort int    `json:"stmp_port"`
	AuthName string `json:"auth_name"`
	AuthPwd  string `json:"auth_pwd"`
}

// TODO:implement
type SmsServer struct {
	AuthName string `json:"auth_name"`
	AuthPwd  string `json:"auth_pwd"`
}

type Receiver struct {
	NickName string `json:"nickname"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
}

func SearchReceiver(arr []*Receiver, nickName string) (int, bool) {
	for i, a := range arr {
		if a.NickName == nickName {
			return i, true
		}
	}
	return -1, false
}

func RemoveReceiver(arr []*Receiver, i int) []*Receiver {
	switch i {
	case 0:
		return arr[1:]
	case len(arr) - 1:
		return arr[:i]
	default:
		result := []*Receiver{}
		result = append(result, arr[:i]...)
		result = append(result, arr[i+1:]...)
		return result
	}
}

type AlarmCfg struct {
	Readme     string      `json:"readme"`
	MailServer MailServer  `json:"mail_server"`
	SmsServer  SmsServer   `json:"sms_server"`
	Receivers  []*Receiver `json:"receviers"`
}

type Alarm struct {
	ticker     *time.Ticker
	mutex      sync.Mutex
	mailClient *mail.MailClient
	cfg        *AlarmCfg
}

func NewAlarm() *Alarm {
	s := &Alarm{}
	return s
}
func (s *Alarm) Deamon() {
	s.ticker = time.NewTicker(5 * time.Minute)
	for {
		<-s.ticker.C
		if _, err := s.LoadCfg(); err != nil {
			log.Warn(errors.As(err))
			continue
		}
		if err := s.Apply(s.Cfg()); err != nil {
			log.Warn(errors.As(err))
			continue
		}
	}
}
func (s *Alarm) LoadCfg() (*AlarmCfg, error) {
	data := []byte{}
	if err := database.QueryElem(db.GetCache("master"), &data, "SELECT cfgdata FROM lserver_cfg WHERE cfgname='alarm'"); err != nil {
		if !errors.ErrNoData.Equal(err) {
			return nil, errors.As(err)
		}
		// no data
	}
	cfg := &AlarmCfg{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, errors.As(err)
		}
	}
	s.cfg = cfg
	return cfg, nil
}

func (s *Alarm) SaveCfg() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(s.cfg.Readme) == 0 {
		s.cfg.Readme = "a warn configuration, it will be read every 5 minutes"
	}
	data, err := json.MarshalIndent(s.cfg, "", "	")
	if err != nil {
		return errors.As(err)
	}

	exist := false
	mdb := db.GetCache("master")
	if err := database.QueryElem(mdb, &exist, "SELECT COUNT(1) FROM lserver_cfg WHERE cfgname='alarm'"); err != nil {
		return errors.As(err)
	}
	if !exist {
		if _, err := mdb.Exec("INSERT INTO lserver_cfg(cfgname,cfgdata)VALUES(?,?)", "alarm", data); err != nil {
			return errors.As(err)
		}
	} else {
		if _, err := mdb.Exec("UPDATE lserver_cfg SET cfgdata=? WHERE cfgname=?", data, "alarm"); err != nil {
			return errors.As(err)
		}
	}
	return nil
}

func (s *Alarm) Cfg() *AlarmCfg {
	if s.cfg.Receivers == nil {
		s.cfg.Receivers = []*Receiver{}
	}
	return s.cfg
}

func (s *Alarm) MailClient() (*mail.MailClient, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.mailClient, s.mailClient != nil
}

func (s *Alarm) Apply(cfg *AlarmCfg) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.mailClient != nil {
		s.mailClient.Close()
		s.mailClient = nil
	}
	if len(cfg.MailServer.SmtpHost) > 0 {
		s.mailClient = mail.NewMailClient(
			cfg.MailServer.SmtpHost,
			cfg.MailServer.SmtpPort,
			cfg.MailServer.AuthName,
			cfg.MailServer.AuthPwd,
		)
		if err := s.mailClient.Test(); err != nil {
			return errors.As(err)
		}
	}

	// TODO: make sms client
	s.cfg = cfg
	return nil
}

// When Apply or Deamon is called, need to call Close
func (s *Alarm) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.ticker != nil {
		s.ticker.Stop()
	}

	if s.mailClient != nil {
		s.mailClient.Close()
		s.mailClient = nil
	}
	return nil
}
