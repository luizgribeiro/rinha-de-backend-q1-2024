package main

import (
	"encoding/json"
	"fmt"
	"luizgribeiro/rinha-backend-q1-24/pkg/store"
	"net/http"
	"os"
	"strconv"
)

func main() {
	store.Init()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /clientes/{id}/transacoes", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		transacao := store.Transacao{}
		d := json.NewDecoder(r.Body)

		err := d.Decode(&transacao)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		err = transacao.EhValida()
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		_id, err := strconv.Atoi(id)
		if err != nil {
			fmt.Fprintf(w, "id conversoin error\n", err)
			return
		}

		if _id < 1 || _id > 5 {
			w.WriteHeader(404)
			return
		}

		accountStatus, err := store.AddTransfer(int32(_id), &transacao)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		payload, err := json.Marshal(accountStatus)
		if err != nil {
			fmt.Fprintf(w, "some error!\n", err)
			return
		}
		_, err = w.Write(payload)
	})

	mux.HandleFunc("GET /clientes/{id}/extrato", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		i, err := strconv.Atoi(id)
		if err != nil {
			fmt.Fprintf(w, "parsing id error!\n", err)
		}

		if i < 1 || i > 5 {
			w.WriteHeader(404)
		}

		info, err := store.GetAccInfo(int32(i))
		if err != nil {
			fmt.Fprintf(w, "some error!\n", err)
		}
		mi, err := json.Marshal(info)

		if err != nil {
			fmt.Fprintf(w, "marshal error!\n", err)

		}
		_, err = w.Write(mi)
	})

	fmt.Println("Starting http server")

	port := os.Getenv("HTTP_PORT")

	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), mux)
}
