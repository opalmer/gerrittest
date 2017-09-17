package main

//
//import "fmt"
//
////import (
////	"github.com/opalmer/logrusutil"
////	"github.com/opalmer/gerrittest"
////)
//
////import (
////	"github.com/opalmer/gerrittest"
////	"github.com/opalmer/logrusutil"
////	log "github.com/sirupsen/logrus"
////)
////
////const codePath = "scripts/example.sh"
////const code = `
////#!/bin/bash
////
////echo "Hello, world"
////`
////
////var destroy = true
////
////func chkerr(err error) {
////	if err != nil {
////		destroy = false
////		panic(err.Error())
////	}
////}
////
////func kill(g *gerrittest.Gerrit) {
////	if destroy {
////		g.Destroy()
////	}
////}
//
//type Thing struct {
//	A []byte
//	B []byte
//}
//
//func (t *Thing) Copy() *Thing {
//	out := new(Thing)
//	*out = *t
//	return out
//}
//
//func main() {
//	thing1 := &Thing{A: []byte("hello")}
//	thing2 := thing1.Copy()
//	//thing2.A = []byte("yes")
//	thing2.B = []byte("world")
//	fmt.Println(string(thing1.A))
//	fmt.Println(string(thing1.B))
//
//	fmt.Println(string(thing2.A))
//	fmt.Println(string(thing2.B))
//
//	//logcfg := logrusutil.NewConfig()
//	//logcfg.Level = "debug"
//	//chkerr(logrusutil.ConfigureLogger(log.StandardLogger(), logcfg))
//	//gerrit, err := gerrittest.New(gerrittest.NewConfig())
//	//chkerr(err)
//	//defer kill(gerrit)
//	//
//	//change, err := gerrit.CreateChange("foobar")
//	//chkerr(err)
//	//
//	//chkerr(change.Write(codePath, 0600, []byte("hello")))
//	//
//	//chkerr(change.AmendAndPush())
//	//_, err = change.AddFileComment("1", codePath, 1, "Test comment.")
//	//chkerr(err)
//	//
//	////chkerr(change.Remove("foo"))
//	////chkerr(change.AmendAndPush())
//	////
//	////_, err = change.AddFileComment("2", codePath, 2, "Test comment.")
//	////chkerr(err)
//	////
//	////_, err = change.ApplyLabel("", gerrittest.CodeReviewLabel, 2)
//	////chkerr(err)
//	////_, err = change.AddTopLevelComment("", "Looks good!")
//	////chkerr(err)
//	////_, err = change.ApplyLabel("", gerrittest.VerifiedLabel, 1)
//	////chkerr(err)
//
//}
