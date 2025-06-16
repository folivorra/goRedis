package cli

import (
	"bufio"
	"context"
	"fmt"
	"github.com/folivorra/goRedis/internal/model"
	"github.com/folivorra/goRedis/internal/storage"
	"os"
	"strconv"
	"strings"
)

type Manager struct {
	store *storage.InMemoryStorage
}

func NewManager(store *storage.InMemoryStorage) *Manager {
	return &Manager{
		store: store,
	}
}

func (m *Manager) Start(ctx context.Context) {
	go func() {
		reader := bufio.NewReader(os.Stdin)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				cmdLine, err := reader.ReadString('\n')
				if err != nil {
					continue
				}

				m.handle(strings.TrimSpace(cmdLine))
			}
		}
	}()
}

func (m *Manager) handle(cmd string) {
	if cmd == "" {
		return
	}

	switch {
	case cmd == "get all":
		all, err := m.store.GetAllItems()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		for _, item := range all {
			fmt.Printf("ID: %d | Name: %s | Price: %.2f\n", item.ID, item.Name, item.Price)
		}

	case strings.HasPrefix(cmd, "get "):
		parts := strings.Split(cmd, " ")
		if len(parts) != 2 {
			fmt.Println("Usage: get <id>")
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Println("Invalid ID")
			return
		}
		item, err := m.store.GetItem(id)
		if err != nil {
			fmt.Println("Not found")
		} else {
			fmt.Printf("ID: %d | Name: %s | Price: %.2f\n", item.ID, item.Name, item.Price)
		}

	case strings.HasPrefix(cmd, "set "):
		parts := strings.Split(cmd, " ")
		if len(parts) != 4 {
			fmt.Println("Usage: set <id> <name> <price>")
			return
		}
		id, err := strconv.Atoi(parts[1])
		if id <= 0 || err != nil {
			fmt.Println("Invalid id")
			return
		}
		name := parts[2]
		price, err := strconv.ParseFloat(parts[3], 64)
		if price < 0 || err != nil {
			fmt.Println("Invalid price")
			return
		}
		err = m.store.AddItem(model.Item{ID: id, Name: name, Price: price})
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Added: ID: %d | Name: %s | Price: %.2f\n", id, name, price)
		}

	case strings.HasPrefix(cmd, "del "):
		parts := strings.Split(cmd, " ")
		if len(parts) != 2 {
			fmt.Println("Usage: del <id>")
			return
		}
		id, err := strconv.Atoi(parts[1])
		if id <= 0 || err != nil {
			fmt.Println("Invalid id")
			return
		}
		err = m.store.DeleteItem(id)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Item with ID %d deleted\n", id)
		}

	case strings.HasPrefix(cmd, "update "):
		parts := strings.Split(cmd, " ")
		if len(parts) != 4 {
			fmt.Println("Usage: update <id> <name> <price>")
			return
		}
		id, err := strconv.Atoi(parts[1])
		if id <= 0 || err != nil {
			fmt.Println("Invalid id")
			return
		}
		name := parts[2]
		price, err := strconv.ParseFloat(parts[3], 64)
		if price < 0 || err != nil {
			fmt.Println("Invalid price")
			return
		}
		err = m.store.UpdateItem(model.Item{ID: id, Name: name, Price: price})
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("Updated: ID: %d | Name: %s | Price: %.2f\n", id, name, price)
		}

	default:
		fmt.Printf("Unknown command. Try:\n" +
			"    get all\n" +
			"    get <id>\n" +
			"    set <id> <name> <price>\n" +
			"    del <id>\n" +
			"    update <id> <name> <price>\n")
	}
}
