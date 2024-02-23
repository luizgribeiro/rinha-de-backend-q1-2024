package main

import (
	"fmt"
	"luizgribeiro/rinha-backend-q1-24/pkg/store"
	"os"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type preAllocIdx struct {
	mu  sync.Mutex
	idx int
	max int
}

// func (p *preAllocIdx) getCurrentIdx() int {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	curr := p.idx
// 	p.idx = p.idx + 1/p.max
// 	return curr
// }
var prealloc string = os.Getenv("PRE_ALLOC")

var pre, _ = strconv.Atoi(prealloc)

var ch chan int = make(chan int, pre)

func setupChan() {
	for i := range pre {
		ch <- i
	}
}

func main() {
	store.Init()
	app := fiber.New()
	setupChan()
	// prealloc := os.Getenv("PRE_ALLOC")

	// pre, _ := strconv.Atoi(prealloc)
	// idx := preAllocIdx{idx: 0, max: pre}

	transacoes := make([]store.Transacao, pre)

	app.Post("/clientes/:id/transacoes", func(c *fiber.Ctx) error {
		id := c.Params("id")

		// i := idx.getCurrentIdx()
		// transacao := &store.Transacao{}
		i := <- ch
		go func () {
			ch <- i
		}()

		if err := c.BodyParser(&transacoes[i]); err != nil {
			// fmt.Println("body parser error")
			return c.Status(422).SendString(err.Error())
		}

		if err := transacoes[i].EhValida(); err != nil {
			// fmt.Println("valida error")
			return c.SendStatus(422)
		}

		_id, err := strconv.Atoi(id)
		if err != nil {
			return c.SendStatus(400)
		}

		if _id < 1 || _id > 5 {
			return c.SendStatus(404)
		}

		accStatus, err := store.AddTransfer(int32(_id), &transacoes[i])
		if err != nil {
			// fmt.Println("trans error", err, transacao)
			return c.SendStatus(422)
		}

		return c.Status(200).SendString(accStatus)
	})

	app.Get("/clientes/:id/extrato", func(c *fiber.Ctx) error {
		id := c.Params("id")
		i, err := strconv.Atoi(id)
		if err != nil {
			return c.SendStatus(400)
		}

		if i < 1 || i > 5 {
			return c.SendStatus(404)
		}

		info, err := store.GetAccInfo(int32(i))
		if err != nil {
			fmt.Println(err, i)
			c.SendStatus(404)
		}
		return c.SendString(info)
	})

	fmt.Println("Starting http server")

	port := os.Getenv("HTTP_PORT")

	app.Listen(fmt.Sprintf("0.0.0.0:%s", port))
}
