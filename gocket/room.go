package gocket

import "sync"

const (
	idsRoom = "__rid"
)

type Rooms struct {
	mu    sync.RWMutex
	rooms map[string]map[string]*Conn
}

func newRooms() *Rooms {
	return &Rooms{rooms: make(map[string]map[string]*Conn)}
}

func (s *Rooms) Get(connId string) (*Conn, bool) {
	s.mu.RLock()
	var (
		conn *Conn
		ok   bool
	)
	idRoom, roomExist := s.rooms[idsRoom]
	if roomExist {
		conn, ok = idRoom[connId]
	}
	s.mu.RUnlock()
	return conn, ok
}

func (r *Rooms) Join(conn *Conn, rooms ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, room := range rooms {
		conns, exists := r.rooms[room]
		if !exists {
			conns = make(map[string]*Conn)
			r.rooms[room] = conns
		}
		conns[conn.Id()] = conn
	}
}

func (r *Rooms) Leave(connId string, rooms ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, room := range rooms {
		conns, exists := r.rooms[room]
		if !exists {
			continue
		}
		delete(conns, connId)
		if len(conns) == 0 {
			delete(r.rooms, room)
		}
	}
}

func (s *Rooms) ToBroadcast(rooms ...string) []*Conn {
	s.mu.RLock()
	result := make([]*Conn, 0)
	for _, room := range rooms {
		if conns, ok := s.rooms[room]; ok {
			for _, conn := range conns {
				result = append(result, conn)
			}
		}
	}
	s.mu.RUnlock()
	return result
}

func (s *Rooms) Len(room string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.rooms[room])
}

func (s *Rooms) Clear(rooms ...string) {
	s.mu.Lock()
	for _, room := range rooms {
		delete(s.rooms, room)
	}
	s.mu.Unlock()
}

func (s *Rooms) Rooms() []string {
	s.mu.RLock()
	result := make([]string, 0, len(s.rooms))
	for room := range s.rooms {
		if room != idsRoom {
			result = append(result, room)
		}
	}
	s.mu.RUnlock()
	return result
}

func (s *Rooms) AllConns() []*Conn {
	return s.ToBroadcast(idsRoom)
}

func (s *Rooms) add(conn *Conn) {
	s.Join(conn, idsRoom)
}

func (s *Rooms) remove(conn *Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for room, conns := range s.rooms {
		delete(conns, conn.Id())
		if len(conns) == 0 {
			delete(s.rooms, room)
		}
	}
}
