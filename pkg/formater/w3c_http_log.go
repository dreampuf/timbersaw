package formater

import (
	"regexp"
	"strconv"
	"sync"
)

type Entity struct {
	RemoteHost, Uid, Timestamp, Method, Request, Status string
	Bytes uint64
}

var (
	reCLF = regexp.MustCompile(`^(?P<client>\S+) \S+ (?P<userid>\S+) \[(?P<datetime>[^\]]+)\] "(?P<method>[A-Z]+) (?P<request>[^ "]+)? HTTP/[0-9.]+" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-)`)
	entityPool = sync.Pool{New: func() interface{} {
		return &Entity{}
	}}
)

type w3cformatter struct {}

func (w *w3cformatter) Format(content string) *Entity {
	v := reCLF.FindStringSubmatch(content)
	if len(v) != 8 {
		return nil
	}
	byteSize, _ := strconv.ParseUint(v[6], 10, 64)

	entity := entityPool.Get().(*Entity)
	entity.RemoteHost = v[1]
	entity.Uid = v[2]
	entity.Timestamp = v[3]
	entity.Method = v[4]
	entity.Request = v[5]
	entity.Status = v[6]
	entity.Bytes = byteSize

	return entity
}

func (w *w3cformatter) Put(e *Entity) {
	entityPool.Put(e)

}

func NewHTTPCommanLogFormatter() Formater {
	return &w3cformatter{}
}
