package main
//
//import (
//	"fmt"
//	"encoding/json"
//	"reflect"
//
//
//)
//const (
//	JSON_PREFIX = ""
//	JSON_INDENT = "    "
//)
//
//func main() {
//	user:=new(User)
//	user.Email="aaa"
//	user.UserId="bbb"
//	usrBytes:=toJSON(user)
//	fmt.Println(string(usrBytes))
//	testa := fromJSON(usrBytes, User{})
//	usera, err := testa.(User)
//	fmt.Println(err)
//
//	fmt.Println(usera.UserId)
//
//}
//type User struct {
//	UserId  string   `json:"userId,omitempty"`
//	Email   string   `json:"email,omitempty"`
//	Phone   string   `json:"phone,omitempty"`
//	DocType string   `json:"docType,omitempty"`
//}
//
//func toJSON(anyStruct interface{}) ([]byte) {
//	bytes, _ := json.MarshalIndent(anyStruct, JSON_PREFIX, JSON_INDENT)
//	return bytes
//}
//func fromJSON(jsonBytes []byte, anyInterface interface{}) (interface{})  {
//	instance := reflect.New(reflect.TypeOf(anyInterface));
//	_ = json.Unmarshal(jsonBytes, &instance)
//	fmt.Println(reflect.TypeOf(anyInterface))
//	return instance;
//}
//
//
