package channels

import (
	"log"
	"net"
	"runtime"
	"sync"
)

type Channel struct {
	Subscribers map[string]net.Conn
	Secret      uint64
	Mutex       sync.Mutex
}

func NewChannel(secret uint64) *Channel {
	return &Channel{
		Secret:      secret,
		Subscribers: make(map[string]net.Conn),
		Mutex:       sync.Mutex{},
	}
}

func (c *Channel) Contains(conn net.Conn) bool {
	_, ok := c.Subscribers[conn.RemoteAddr().String()]
	return ok
}

func (c *Channel) AddSubscriber(conn net.Conn) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.Subscribers[conn.RemoteAddr().String()] = conn
	go log.Println("Added new subscriber:", conn.RemoteAddr().String())
}

func (c *Channel) RemoveSubscriber(conn net.Conn) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	delete(c.Subscribers, conn.RemoteAddr().String())
	go log.Println("Removed subscriber:", conn.RemoteAddr().String())
}

func (c *Channel) IsEmpty() bool {
	return len(c.Subscribers) == 0
}

func (c *Channel) Broadcast(msg string) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	var wg sync.WaitGroup
	for _, conn := range c.Subscribers {
		wg.Add(1)
		runtime.Gosched()
		go func() {
			defer wg.Done()
			conn.Write([]byte(msg))
		}()
	}
	wg.Wait()
}
