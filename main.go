package main

import (
	"fmt"
	"luizgribeiro/rinha-backend-q1-24/pkg/store"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func main() {
	store.Init()
	app := fiber.New()
	app.Post("/clientes/:id/transacoes", func(c *fiber.Ctx) error {
		id := c.Params("id")

		transacao := &store.Transacao{}
		if err := c.BodyParser(transacao); err != nil {
			return c.Status(422).SendString(err.Error())
		}

		if err := transacao.EhValida(); err != nil {
			return c.SendStatus(422)
		}

		_id, err := strconv.Atoi(id)
		if err != nil {
			return c.SendStatus(400)
		}

		if _id < 1 || _id > 5 {
			return c.SendStatus(404)
		}

		accStatus, err := store.AddTransfer(int32(_id), transacao)
		if err != nil {
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
