package room

import (
	"fmt"
	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/scheduler"
	"github.com/lonng/nano/session"
	"time"
)

type (
	// Manager represents a component that contains a bundle of room
	Manager struct {
		component.Base
		timer *scheduler.Timer
		rooms map[int]*Room
	}

	// UserMessage represents a message that user sent
	UserMessage struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	// NewUser message will be received when new user join room
	NewUser struct {
		Content string `json:"content"`
	}

	// AllMembers contains all members uid
	AllMembers struct {
		Members []int64 `json:"members"`
	}

	// JoinResponse represents the result of joining room
	JoinResponse struct {
		Code   int    `json:"code"`
		Result string `json:"result"`
	}
)

const (
	testRoomID = 1
	roomIDKey  = "ROOM_ID"
)

func NewManager() *Manager {
	return &Manager{
		rooms: map[int]*Room{},
	}
}

// AfterInit component lifetime callback
func (mgr *Manager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		if !s.HasKey(roomIDKey) {
			return
		}
		room := s.Value(roomIDKey).(*Room)
		room.group.Leave(s)
	})
	mgr.timer = scheduler.NewTimer(time.Minute, func() {
		for roomId, room := range mgr.rooms {
			println(fmt.Sprintf("UserCount: RoomID=%d, Time=%s, Count=%d",
				roomId, time.Now().String(), room.group.Count()))
		}
	})
}

// Join room
func (mgr *Manager) Join(s *session.Session, msg []byte) error {
	// NOTE: join test room only in demo
	room, found := mgr.rooms[testRoomID]
	if !found {
		room = &Room{
			group: nano.NewGroup(fmt.Sprintf("room-%d", testRoomID)),
		}
		mgr.rooms[testRoomID] = room
	}

	fakeUID := s.ID() //just use s.ID as uid !!!
	s.Bind(fakeUID)   // binding session uids.Set(roomIDKey, room)
	s.Set(roomIDKey, room)
	s.Push("onMembers", &AllMembers{Members: room.group.Members()})
	// notify others
	room.group.Broadcast("onNewUser", &NewUser{Content: fmt.Sprintf("New user: %d", s.ID())})
	// new user join group
	room.group.Add(s) // add session to group
	return s.Response(&JoinResponse{Result: "success"})
}

// Message sync last message to all members
func (mgr *Manager) Message(s *session.Session, msg *UserMessage) error {
	if !s.HasKey(roomIDKey) {
		return fmt.Errorf("not join room yet")
	}
	room := s.Value(roomIDKey).(*Room)
	return room.group.Broadcast("onMessage", msg)
}
