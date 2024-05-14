package dao

import (
	"errors"
	"time"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DocInter interface {
	Collection
	GetDoc() interface{}
	GetID() interface{}
	SetCreator(u LogUser)
	AddRecord(u LogUser,msg string) []*Record
	GetIndexes() []mongo.IndexModel
}

type Collection interface {
	GetC() string
}

type LogUser interface{
	GetName() string
	GetAccount() string
}

func NewRecord(date time.Time,sum string,acc string,name string)*Record{
	return &Record{
		DateTime: date,
		Summary: sum,
		Account: acc,
		Name: name,
	}
}

type Record struct{
	DateTime time.Time
	Summary string
	Account string
	Name string
}
type CommonDoc struct {
	Records []*Record
}
func (c *CommonDoc) AddRecord(u LogUser, msg string) []*Record {
	c.Records = append(c.Records, &Record{
		DateTime: time.Now(),
		Summary:  msg,
		Account:  u.GetAccount(),
		Name:     u.GetName(),
	})
	return c.Records
}

func (c *CommonDoc) SetCreator(lu LogUser) {
	if c == nil {
		return
	}
	c.Records = append(c.Records, &Record{
		DateTime: time.Now(),
		Summary:  "create",
		Account:  lu.GetAccount(),
		Name:     lu.GetName(),
	})
}

func (u *CommonDoc) GetC() string {
	panic("must override")
}


func Format(inter interface{}, f func(i interface{}) map[string]interface{}) (interface{}, int) {
	ik := reflect.TypeOf(inter).Kind()
	if ik == reflect.Ptr {
		return f(inter), 1
	}
	if ik != reflect.Slice {
		return nil, 0
	}
	v := reflect.ValueOf(inter)
	l := v.Len()
	count := 0
	ret := []map[string]interface{}{}
	for i := 0; i < l; i++ {
		if data := f(v.Index(i).Interface()); data != nil {
			ret = append(ret, data)
			count++
		}
	}
	return ret, count
}

func GetObjectID(id interface{}) (primitive.ObjectID, error) {
	var myID primitive.ObjectID
	switch dtype := reflect.TypeOf(id).String(); dtype {
	case "string":
		str := id.(string)
		return primitive.ObjectIDFromHex(str)
	case "primitive.ObjectID":
		myID = id.(primitive.ObjectID)
	default:
		return primitive.NilObjectID, errors.New("not support type: " + dtype)
	}
	return myID, nil
}