// AGPL License
// Copyright 2022 ysicing(i@ysicing.me).

package util

import (
	"net"
	"strconv"
	"time"

	appsv1beta1 "github.com/ysicing/cloudflow/apis/apps/v1beta1"
)

type DB interface {
	CheckNetWork() error
	CheckAuth() error
}

type DBMeta struct {
	appsv1beta1.GlobalDBSpec
}

func (m DBMeta) CheckNetWork() error {
	address := net.JoinHostPort(m.Source.Host, strconv.Itoa(m.Source.Port))
	_, err := net.DialTimeout("tcp", address, 5*time.Second)
	return err
}

func NewDB(gdb appsv1beta1.GlobalDBSpec) DB {
	if gdb.Type == "mysql" {
		return &MysqlMeta{DBMeta: DBMeta{GlobalDBSpec: gdb}}
	}
	return nil
}
