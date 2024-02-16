package store

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Transacao struct {
	Valor     int32  `bson:"valor" json:"valor"`
	Tipo      string `bson:"tipo" json:"tipo"`
	Descricao string `bson:"descricao" json:"descricao"`
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

var syncker = make(map[int32](chan int32))

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

type ResultTransfer struct {
	Saldo  int32 `bson:"saldo" json:"saldo"`
	Limite int32 `bson:"limite" json:"limite"`
}

func AddTransfer(id int32, transacao *Transacao) (*ResultTransfer, error) {

	_, ok := syncker[id]

	if !ok {
		createSyncKerForId(id)
	}

	var opVal int32
	var filter bson.D

	if transacao.Tipo == "d" {
		opVal = -transacao.Valor
		filter = bson.D{{"_id", id}, {"gordurinha", bson.D{{"$gte", transacao.Valor}}}}
	} else {
		opVal = transacao.Valor
		filter = bson.D{{"_id", id}}
	}

	project := bson.D{
		{"$project", bson.D{
			{"id", 1},
			{"saldo", 1},
			{"gordurinha", 1},
			{"ultimas_transacoes", bson.D{
				{"$slice", []interface{}{"$ultimas_transacoes", 9}},
			}},
		}},
	}

	set := bson.D{
		{"$set", bson.D{
			{"gordurinha", bson.D{
				{"$add", []interface{}{"$gordurinha", opVal}},
			}},
			{"saldo.total", bson.D{
				{"$add", []interface{}{"$saldo.total", opVal}},
			}},
			{"ultimas_transacoes", bson.D{
				{"$concatArrays", []interface{}{[]interface{}{transacao}, "$ultimas_transacoes"}},
			}},
		}},
	}

	after := options.After
	opts := options.FindOneAndUpdateOptions{
		Projection:     bson.D{{"limite", "$saldo.limite"}, {"saldo", "$saldo.total"}},
		ReturnDocument: &after,
	}

	acc := &ResultTransfer{}
	<-syncker[id]
	defer releaseNext(id)
	err := db.coll.FindOneAndUpdate(context.TODO(), filter, mongo.Pipeline{project, set}, &opts).Decode(&acc)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

type Transacoes struct {
	Valor       int64  `bson:"valor" json:"valor"`
	Tipo        string `bson:"tipo" json:"tipo"`
	Descricao   string `bson:"descricao" json:"descricao"`
	RealizadaEm string `bson:"realizada_em" json:"realizada_em"`
}

type Saldo struct {
	Total       int64  `bson:"total" json:"total"`
	DataExtrato string `json:"data_extrato"`
	Limite      int64  `bson:"limite" json:"limite"`
}

type AccountInfo struct {
	ID                int64        `bson:"_id" json:"total"`
	Saldo             Saldo        `bson:"saldo" json:"saldo"`
	UltimasTransacoes []Transacoes `bson:"ultimas_transacoes" json:"ultimas_transacoes"`
}

func GetAccInfo(id int32) (*AccountInfo, error) {

	opts := options.FindOneOptions{
		Projection: bson.D{{"saldo.total", 1}, {"saldo.limite", 1}, {"ultimas_transacoes", 1}},
	}
	acc := &AccountInfo{}

	filter := bson.D{{"_id", id}}
	err := db.coll.FindOne(context.TODO(), filter, &opts).Decode(acc)

	if err != nil {
		return nil, err
	}

	acc.Saldo.DataExtrato = time.Now().Format(time.RFC3339)

	return acc, err
}
