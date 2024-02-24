package store

import (
	"context"
	"fmt"
	"slices"
	// "strings"
	"time"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Transacao struct {
	Valor       int32  `bson:"v" json:"valor"`
	Tipo        string `bson:"t" json:"tipo"`
	Descricao   string `bson:"d" json:"descricao"`
	RealizadaEm string `bson:"r" json:"realizada_em"`
}

type ResultTransfer struct {
	Saldo int64 `bson:"t"`
}

var limites map[int32]int32 = map[int32]int32{
	1: 100000,
	2: 80000,
	3: 1000000,
	4: 10000000,
	5: 500000,
}

func (t Transacao) EhValida() error {
	tipos := []string{"d", "c"}
	if !slices.Contains(tipos, t.Tipo) {
		return fmt.Errorf("Tipo de transacao invalido: %s", t.Tipo)
	}

	if len(t.Descricao) > 10 || len(t.Descricao) == 0 {
		return fmt.Errorf("Descricao de transacao invalida: %s", t.Descricao)
	}

	return nil
}

// func (t *Transacao) serializeTrans() string {
// 	return fmt.Sprintf("{\"valor\":%d,\"tipo\":\"%s\",\"descricao\":\"%s\",\"realizada_em\":\"%s\"}", t.Valor, t.Tipo, t.Descricao,time.Now().Format(time.RFC3339))
// }
var syncker = make(map[int32](chan int32))
var buffSize = 1000

var transBuffers = make(map[int32][]ResultTransfer)

func initializeBuffers(size int) {
	for i := range 5 {
		transBuffers[int32(i+1)] = make([]ResultTransfer, buffSize)
	}
}

func createSyncKerForId(id int32) {
	workers := os.Getenv("N_WORKERS")

	w, err := strconv.Atoi(workers)
	if err != nil {
		panic(err)
	}
	syncker[id] = make(chan int32, w)

	for i := range w {
		syncker[id] <- int32(i)
	}
}

func releaseNext(id int32) {
	syncker[id] <- 0
}


func AddTransfer(id int32, transacao *Transacao) (string, error) {

	// serialTrans := transacao.serializeTrans()

	var opVal int32
	var filter bson.D

	if transacao.Tipo == "d" {
		opVal = -transacao.Valor
		filter = bson.D{{"_id", id}, {"g", bson.D{{"$gte", transacao.Valor}}}}
	} else {
		opVal = transacao.Valor
		filter = bson.D{{"_id", id}}
	}

	// project := bson.D{
	// 	{"$setField", bson.D{
	// 		{"u", bson.D{
	// 			{"$slice", []interface{}{"$u", 9}},
	// 		}},
	// 	}},
	// }

	set := bson.D{
		{"$set", bson.D{
			{"g", bson.D{
				{"$add", []interface{}{"$g", opVal}},
			}},
			{"t", bson.D{
				{"$add", []interface{}{"$t", opVal}},
			}},
			{"u", bson.D{
				{"$concatArrays", []interface{}{[]interface{}{/*serialTrans*/transacao}, bson.D{
					{"$slice", []interface{}{"$u", 9}},
				}}},
			}},
		}},
	}

	after := options.After
	opts := options.FindOneAndUpdateOptions{
		Projection:     bson.D{{"t", 1}},
		ReturnDocument: &after,
	}

	<-syncker[id]
	defer releaseNext(id)
	acc := &ResultTransfer{}
	err := db.coll.FindOneAndUpdate(context.TODO(), filter, mongo.Pipeline{/*project,*/ set}, &opts).Decode(&acc)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("{\"saldo\":%d,\"limite\":%d}", acc.Saldo, limites[id]), nil
}

type Transacoes struct {
	Valor       int64  `bson:"v" json:"valor"`
	Tipo        string `bson:"t" json:"tipo"`
	Descricao   string `bson:"d" json:"descricao"`
	RealizadaEm string `bson:"r" json:"realizada_em"`
}

type Saldo struct {
	Total       int64  `bson:"t" json:"total"`
	DataExtrato string `json:"data_extrato"`
	Limite      int64  `bson:"l" json:"limite"`
}

type AccountInfo struct {
	Total             int64        `bson:"t" json:"total"`
	// Limite            int64        `bson:"l" json:"limite"`
	UltimasTransacoes []Transacoes `bson:"u" json:"ultimas_transacoes"`
}

func GetAccInfo(id int32) (string, error) {

	opts := options.FindOneOptions{
		Projection: bson.D{
			{"t", 1},
			{"u", 1},
		},
	}
	acc := &AccountInfo{}

	filter := bson.D{{"_id", id}}
	err := db.coll.FindOne(context.TODO(), filter, &opts).Decode(acc)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("{\"saldo\":{\"total\":%d,\"limite\":%d,\"data_extrato\":\"%s\"},\"ultimas_transacoes\":%s}", acc.Total, limites[id], time.Now().Format(time.RFC3339), marshalUltimasTransacoes(acc)), err
}

func marshalUltimasTransacoes(acc *AccountInfo) string {

	// return fmt.Sprintf("[%s]", strings.Join(acc.UltimasTransacoes, ","))

	s := ""
	maxLen := len(acc.UltimasTransacoes) - 1

	for i, trans := range acc.UltimasTransacoes {
		sep := ","
		if i == maxLen {
			sep = ""
		}
		s += fmt.Sprintf("{\"valor\":%d,\"tipo\":\"%s\",\"descricao\":\"%s\",\"realizada_em\":\"%s\"}%s", trans.Valor, trans.Tipo, trans.Descricao, trans.RealizadaEm, sep)
	}

	return fmt.Sprintf("[%s]", s)
}
