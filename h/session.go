package h

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
	"tollsys/tollmon/g"
	"tollsys/tollmon/redis"

	"github.com/gin-gonic/gin"
)

const SESSION = "Session"

var (
	lock    = &sync.Mutex{}
	manager *SessionManager
)

//Session管理器，实际操作Session存储和cookie
type SessionManager struct {
	cookieName string
	storage    *SessionStorage
	maxAge     int64
	lock       sync.Mutex
}

//声明SessionStorage操作接口，抽象Session存储操作，存储模式不同实现方式也不同
//本项目中采用redis存储管理session
type ISessionStorage interface {
	InitSession(sid string, maxAge int64) (*Session, error)
	GetSession(sid string) (*Session, bool)
	SetSession(s *Session) error
	DestroySession(sid string) error
	GC()
}

//声明Session操作接口,不同存储方式的Session操作不同，实现也不同
type ISession interface {
	Set(key, value interface{})
	Get(key interface{}) interface{}
	Remove(Key interface{}) error
	GetID() string
}
type SessionStorage struct {
	lock sync.Mutex
}

type Session struct {
	Sid            string                 `json:"sid"`
	lock           sync.Mutex
	LastAccessTime time.Time              `json:"lastAccessTime"`
	MaxAge         int64                  `json:"maxAge"`
	Data           map[string]interface{} `json:"data"`
}

func newSession() *Session {
	return &Session{
		Data:   make(map[string]interface{}),
		MaxAge: 60 * 60 * 24 * int64(g.Config().Session.MaxAge),
	}
}
func (s *Session) Set(key string, value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Data[key] = value
}
func (s *Session) Get(key string) interface{} {
	if value := s.Data[key]; value != nil {
		return value
	}
	return nil
}
func (s *Session) Remove(key string) (e error) {
	e = errors.New("session is nil")
	if value := s.Data[key]; value != nil {
		delete(s.Data, key)
		return nil
	}
	return e
}
func (s *Session) GetID() string {
	return s.Sid
}

func newSessionStorage() *SessionStorage {
	return &SessionStorage{}
}

//新建session，并将session序列化为json写入redis，同时返回该session
//新建存储采用HSETNX，当不存在时才写入redis
func (ss *SessionStorage) InitSession(sid string, maxAge int64) (*Session, error) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	newSession := newSession()
	newSession.Sid = sid
	newSession.LastAccessTime = time.Now()
	value, _ := g.Json.Marshal(newSession)
	redis.HSetNX(SESSION, sid, value)
	return newSession, nil
}

//在redis中通过sessionid获取该session并反序列化为session对象
func (ss *SessionStorage) GetSession(sid string) (*Session, bool) {
	if redis.HExists(SESSION, sid) {
		s := newSession()
		b := redis.HGet(SESSION, sid)
		err := g.Json.Unmarshal(b, s)
		if err != nil {
			g.LogDebug(err.Error())
			return nil, false
		}
		return s, true
	}
	return nil, false
}

//在redis中更新session，要求session参数
func (ss *SessionStorage) SetSession(s *Session) error {
	e := errors.New(s.GetID() + " is blank")
	if ok := redis.HExists(SESSION, s.GetID()); ok {
		b, _ := g.Json.Marshal(s)
		redis.HSet(SESSION, s.GetID(), b)
		s = nil
		return nil
	}
	return e
}

//手动删除session并gc
func (ss *SessionStorage) DestroySession(sid string) error {
	e := errors.New(sid + " is blank")
	if redis.HExists(SESSION, sid) {
		redis.HDel(SESSION, sid)
		return nil
	}
	return e
}

//自动gc过期session
//由go程启动,30分钟执行一次gc 清除过期的session
func (ss *SessionStorage) GC() {
	g.LogInfo("goroutine:session gc")
	for {
		lists := map[string]*Session{}
		b := redis.HGetALL(SESSION)
		for index := 0; index < len(b); index += 2 {
			k := string(b[index].([]byte))
			v := Session{}
			g.Json.Unmarshal(b[index+1].([]byte), &v)
			lists[k] = &v
		}
		if len(lists) < 1 {
			time.Sleep(time.Duration(30 * time.Minute))
			continue
		}

		for k, v := range lists {
			t := v.LastAccessTime.Unix() + v.MaxAge

			if t < time.Now().Unix() {
				g.LogInfo(k, " is over time -> GC")
				ss.DestroySession(k)
			}
		}
		time.Sleep(time.Duration(30 * time.Minute))
	}
}

//创建全局session管理实例,返回单例
func NewSessionManager() *SessionManager {
	g.LogInfo("Init Session Manager...")
	lock.Lock()
	defer lock.Unlock()
	if manager == nil {
		manager = &SessionManager{
			cookieName: g.Config().Session.CookieName,
			maxAge:     60 * 60 * 24 * int64(g.Config().Session.MaxAge),
			storage:    newSessionStorage(),
		}
		go manager.storage.GC()
	}
	g.LogInfo("Init Session Manager OK...")
	return manager
}
func (m *SessionManager) GetCookName() string {
	return m.cookieName
}

//forDebug
func (m *SessionManager) TestSession() {
	m.lock.Lock()
	defer m.lock.Unlock()

	session, err := m.storage.InitSession(randomId(), int64(g.Config().Session.MaxAge))
	if err != nil {
		panic(err)
	}
	session.Set("station", "1F01000000000401")
	m.storage.SetSession(session)

	s, ok := m.storage.GetSession(session.Sid)
	if ok {
		g.LogDebug(s)
	}
}

//开始Session流程，判断request中cookie请求，若无cookie或cookie对应session不存在则创建
//若存在则更新cookie
func (m *SessionManager) BeginSession(c *gin.Context) (*Session, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	cookie, err := c.Request.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		g.LogDebug("Current Session Not Exist")
		session, _ := m.storage.InitSession(randomId(), m.maxAge)
		cookie = getCookie(m.cookieName, session.Sid, m.maxAge)
		http.SetCookie(c.Writer, cookie)
		return session, nil //创建新的session、cookie
	}
	//cookie存在
	sid, _ := url.QueryUnescape(cookie.Value)
	if session, ok := m.storage.GetSession(sid); ok {
		return session, nil //对应的session存在
	} else {
		if session == nil {
			session, _ = m.storage.InitSession(randomId(), m.maxAge)
			cookie = getCookie(m.cookieName, session.Sid, m.maxAge)
			http.SetCookie(c.Writer, cookie)
		}
		return session, nil //对应的session不存在，创建新的session并更新cookie
	}
}

//更新cookie和session，cookie写入httprequest，session写入redis
func (m *SessionManager) Update(c *gin.Context, key string, val interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()

	cookie, err := c.Request.Cookie(m.cookieName)
	if err != nil {
		g.LogError("get cookie err:", err.Error())
		return
	}
	sid, _ := url.QueryUnescape(cookie.Value)

	if session, ok := m.storage.GetSession(sid); ok {
		session.LastAccessTime = time.Now()
		session.Set(key, val)
		cookie.Expires = time.Now().Add(30 * 24 * time.Hour)
		cookie.MaxAge = int(session.MaxAge)

		http.SetCookie(c.Writer, cookie)
		m.storage.SetSession(session)
	}
}

//根据request请求获取session
func (m *SessionManager) GetSession(c *gin.Context) *Session {

	cookie, err := c.Request.Cookie(m.cookieName)
	if err != nil {
		g.LogError("get cookie err:", err.Error())
		return nil
	}
	sid, _ := url.QueryUnescape(cookie.Value)
	if v, ok := m.storage.GetSession(sid); ok {
		return v
	}
	return nil
}

//根据request请求获取cookie，并更新对应session的data，data要求map
func (m *SessionManager) SetSession(c *gin.Context, data map[string]interface{}) {
	cookie, err := c.Request.Cookie(m.cookieName)
	if err != nil {
		g.LogError("get cookie err:", err.Error())
		return
	}
	sid, _ := url.QueryUnescape(cookie.Value)
	if s, ok := m.storage.GetSession(sid); ok {
		for k, v := range data {
			s.Set(k, v)
		}
		m.storage.SetSession(s)
		g.LogInfo("update session to redis ok:", s.GetID())
	}
}
func (m *SessionManager) GetSessionByID(sid string) *Session {
	if v, ok := m.storage.GetSession(sid); ok {
		return v
	}
	return nil
}
func (m *SessionManager) CheckSession(c *gin.Context) bool {
	sid := getSessionIDByRequest(c)
	if sid != "" {
		if redis.HExists(SESSION, sid) {
			return true
		}
	}
	return false
}

//检查Session是否存在
func (m *SessionManager) CheckSessionByID(sid string) bool {
	if redis.HExists(SESSION, sid) {
		return true
	}
	return false
}

//手动删除Session,并清空cookie
func (m *SessionManager) Destroy(c *gin.Context) {
	sid := getSessionIDByRequest(c)
	if sid != "" {
		m.lock.Lock()
		defer m.lock.Unlock()

		m.storage.DestroySession(sid)
		cookie := getCookie(m.cookieName, "", 0)
		http.SetCookie(c.Writer, cookie)
	}
}
func (m *SessionManager) SetMaxAge(i int64) {
	m.maxAge = i
}
func randomId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	//加密
	return base64.URLEncoding.EncodeToString(b)
}
func getCookie(cookieName string, sid string, maxAge int64) *http.Cookie {
	c := http.Cookie{
		Name:  cookieName,
		Value: url.QueryEscape(sid),
		Path:  "/",
		//HttpOnly: false,
		MaxAge:  int(maxAge),
		Expires: time.Now().Add(time.Duration(maxAge)),
	}
	return &c
}
func getSessionIDByRequest(c *gin.Context) string {
	cookie, err := c.Request.Cookie(Manager.cookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	sid, _ := url.QueryUnescape(cookie.Value)
	return sid
}
