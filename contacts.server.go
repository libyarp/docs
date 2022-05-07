package contacts

import (
	"context"
	"github.com/libyarp/yarp"
	"sync"
)

type Service struct {
	idCounter int64
	mu        sync.Mutex
	contacts  map[int64]Contact
}

func (s *Service) insert(c *Contact) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.idCounter + 1
	s.idCounter = id
	c.ID = &id
	s.contacts[id] = *c
}

func (s *Service) update(id int64, c *Contact) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contacts[id] = *c
}

func (s *Service) UpsertContact(ctx context.Context, headers yarp.Header, req *Contact) (yarp.Header, error) {
	if req.ID != nil {
		s.update(*req.ID, req)
	} else {
		s.insert(req)
	}
	return nil, nil
}

func (s *Service) ListContacts(ctx context.Context, headers yarp.Header, out *ContactStreamer) error {
	for _, c := range s.contacts {
		out.Push(&c)
	}
	return nil
}

func (s *Service) GetContact(ctx context.Context, headers yarp.Header, req *GetContactRequest) (yarp.Header, *GetContactResponse, error) {
	v, ok := s.contacts[req.ID]
	if !ok {
		return nil, &GetContactResponse{}, nil
	}
	return nil, &GetContactResponse{Contact: &v}, nil
}

func RunServer() {
	RegisterMessages()
	s := yarp.NewServer("localhost:9027")
	service := &Service{
		idCounter: 0,
		mu:        sync.Mutex{},
		contacts:  map[int64]Contact{},
	}
	RegisterContactsService(s, service)
	s.Start()
}
