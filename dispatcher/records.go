package dispatcher

import (
	"fmt"
	"strings"
	"time"

	"airdispat.ch/message"
)

// Table Names
const (
	TableNameUser        = "melange_user"
	TableNameIdentity    = "melange_identity"
	TableNameMessage     = "melange_message"
	TableNameStorage     = "melange_storage"
	TableNameDataMessage = "melange_data_msg"
)

// Primary Key Names
const (
	PKUser        = "Id"
	PKIdentity    = "Id"
	PKMessage     = "Id"
	PKStorage     = "Id"
	PKDataMessage = "Id"
)

// Create the Table Objects
func (s *Server) CreateTables() {
	// User Management
	s.dbmap.AddTableWithName(User{}, TableNameUser).SetKeys(true, PKUser)
	s.dbmap.AddTableWithName(Identity{}, TableNameIdentity).SetKeys(true, PKIdentity)

	// Message Management
	s.dbmap.AddTableWithName(Message{}, TableNameMessage).SetKeys(true, PKMessage)

	// Data Message Management
	s.dbmap.AddTableWithName(File{}, TableNameDataMessage).SetKeys(true, PKDataMessage)

	// Storage
	s.dbmap.AddTableWithName(Storage{}, TableNameStorage).SetKeys(true, PKStorage)
}

// File represents the AD Data messages.
type File struct {
	Id int
	// To / From Information
	Owner  int
	To     string
	Sender string
	// Message Information
	Name    string
	Path    string
	Length  int64
	Message []byte
	// Metadata
	Received int64
}

// Change Outgoing Message into Encrypted Message
func (o *File) ToDispatch(retriever string) (*message.EncryptedMessage, error) {
	return message.CreateEncryptedMessageFromBytes(o.Message)
}

// Save Data Message
func (m *Server) SaveDataMessage(name string, to []string, from string, message *message.EncryptedMessage, path string, length int64) error {
	ownerId := -1
	if from != "" {
		user, err := m.UserForIdentity(from)
		if err != nil {
			fmt.Println("Got error getting user for identity.")
			return err
		}
		ownerId = user.Id
	}

	data, err := message.ToBytes()
	if err != nil {
		return err
	}

	out := &File{
		To:       strings.Join(to, ","),
		Sender:   from,
		Owner:    ownerId,
		Name:     name,
		Message:  data,
		Path:     path,
		Length:   length,
		Received: time.Now().Unix(),
	}
	return m.dbmap.Insert(out)
}

const QueryDataMessage = "select * from " + TableNameDataMessage + " o where o.Owner = :owner and o.Name = :name and (o.To like :recv or o.To = '')"

// Return Outgoing Message Named
func (m *Server) GetDataMessageNamed(name string, owner string, receiver string) (*File, error) {
	user, err := m.UserForIdentity(owner)
	if err != nil {
		return nil, err
	}

	result := &File{}

	// Create the Query
	err = m.dbmap.SelectOne(result, QueryDataMessage,
		map[string]interface{}{
			"name":  name,
			"recv":  fmt.Sprintf("%%%s%%", receiver),
			"owner": user.Id,
		})

	return result, err
}

// Outgoing Messages
type Message struct {
	Id int
	// Recipient Information
	To     string
	Sender string
	Owner  int
	// Message Information
	Name string
	Data []byte
	Type int
	// Metadata
	Received int64
	// Transient
	allowed []string `db:"-"`
}

const QueryOutgoingNamed = "select * from " + TableNameMessage + " o where o.Owner = :owner and o.Name = :name and ((o.To like :recv and o.Type = 1) or (o.To = '' and o.Type = 0))"
const QueryAnyNamed = "select * from " + TableNameMessage + " o where o.Owner = :owner and o.Name = :name"
const QueryOutgoingPublic = "select * from " + TableNameMessage + " o where o.Owner = :owner and (o.To like :recv or o.To = '') and o.Received > :time and o.Type = 0"
const QueryOutgoing = "select * from " + TableNameMessage + " o where o.Owner = :owner and (o.Type = 0 or o.Type = 1) and o.Received > :time"

func (m *Server) GetOutgoingMessagesFor(since uint64, owner string) ([]*Message, error) {
	user, err := m.UserForIdentity(owner)
	if err != nil {
		return nil, err
	}

	var results []*Message

	// Create the Query
	_, err = m.dbmap.Select(&results, QueryOutgoing,
		map[string]interface{}{
			"owner": user.Id,
			"time":  since,
		})

	return results, err
}

func (m *Server) GetAnyMessageWithName(name string, owner string) (*Message, error) {
	user, err := m.UserForIdentity(owner)
	if err != nil {
		return nil, err
	}

	result := &Message{}

	// Create the Query
	err = m.dbmap.SelectOne(result, QueryAnyNamed,
		map[string]interface{}{
			"name":  name,
			"owner": user.Id,
		})

	return result, err
}

// Return Outgoing Message Named
func (m *Server) GetOutgoingMessageWithName(name string, owner string, receiver string) (*Message, error) {
	user, err := m.UserForIdentity(owner)
	if err != nil {
		return nil, err
	}

	result := &Message{}

	// Create the Query
	err = m.dbmap.SelectOne(result, QueryOutgoingNamed,
		map[string]interface{}{
			"name":  name,
			"recv":  fmt.Sprintf("%%%s%%", receiver),
			"owner": user.Id,
		})

	return result, err
}

// Return Outgoing Public Messages for a Receiver
func (m *Server) GetOutgoingPublicMessagesFor(since uint64, owner string, receiver string) ([]*Message, error) {
	user, err := m.UserForIdentity(owner)
	if err != nil {
		return nil, err
	}

	var results []*Message

	// Create the Query
	_, err = m.dbmap.Select(&results, QueryOutgoingPublic,
		map[string]interface{}{
			"recv":  fmt.Sprintf("%%%s%%", receiver),
			"owner": user.Id,
			"time":  since,
		})

	return results, err
}

const (
	TypeOutgoingPublic = iota
	TypeOutgoingPrivate
	TypeIncoming
)

// Save Outgoing Message
func (m *Server) SaveMessage(name string, to []string, from string, message *message.EncryptedMessage, messageType int) error {
	ownerId := -1
	if from != "" {
		user, err := m.UserForIdentity(from)
		if err != nil {
			fmt.Println("Got error getting user for identity.")
			return err
		}
		ownerId = user.Id
	}

	data, err := message.ToBytes()
	if err != nil {
		return err
	}

	out := &Message{
		To:       strings.Join(to, ","),
		Sender:   from,
		Owner:    ownerId,
		Name:     name,
		Data:     data,
		Type:     messageType,
		Received: time.Now().Unix(),
	}
	return m.dbmap.Insert(out)
}

// Change Outgoing Message into Encrypted Message
func (o *Message) ToDispatch(retriever string) (*message.EncryptedMessage, error) {
	return message.CreateEncryptedMessageFromBytes(o.Data)
}

// Save Outgoing Message
func (m *Server) SaveIncomingMessage(message *message.EncryptedMessage) error {
	keys := make([]string, len(message.Header))
	i := 0
	for key, _ := range message.Header {
		keys[i] = key
		i++
	}

	return m.SaveMessage("", keys, "", message, TypeIncoming)
}

const QueryIncoming = "select * from " + TableNameMessage + " o where o.To like :owner and o.Received > :time and o.Type = 2"

// Return Incoming Messages Since
func (m *Server) GetIncomingMessagesSince(since uint64, owner string) ([]*Message, error) {
	var results []*Message

	// Create the Query
	_, err := m.dbmap.Select(&results, QueryIncoming,
		map[string]interface{}{
			"owner": fmt.Sprintf("%%%s%%", owner),
			"time":  since,
		})

	return results, err
}

type Storage struct {
	Id    int
	Key   string
	Value []byte
	Owner int
}

const QueryStorage = "select s.Key, s.Value from " + TableNameStorage + " s, " + TableNameUser + " u, " + TableNameIdentity + " i where " +
	"s.Key = :key and u.Id = i.Owner and i.Signing = :signing and u.Id = s.Owner"

func (m *Server) GetData(author string, key string) ([]byte, error) {
	var result *Storage

	// Create the Query
	err := m.dbmap.SelectOne(&result, QueryStorage,
		map[string]interface{}{
			"key":     key,
			"signing": author,
		})
	if err != nil {
		return nil, err
	}

	return result.Value, err
}

func (m *Server) SetData(author string, key string, data []byte) error {
	u, err := m.UserForIdentity(author)
	if err != nil {
		return err
	}

	insertion := &Storage{
		Key:   key,
		Value: data,
		Owner: u.Id,
	}
	return m.dbmap.Insert(insertion)
}

type User struct {
	Id           int
	Name         string
	Receiving    string
	RegisteredOn int64
}

const QueryIdentity = "select u.Id, u.Name, u.Receiving, u.RegisteredOn from " +
	TableNameUser + " u, " + TableNameIdentity + " i where " +
	"u.Id = i.Owner and i.Signing = :key"

func (m *Server) UserForIdentity(id string) (*User, error) {
	result := &User{}
	// Create the Query
	err := m.dbmap.SelectOne(result, QueryIdentity,
		map[string]interface{}{
			"key": id,
		})
	return result, err
}

type Identity struct {
	// Signing Key and Encryption Key
	Id         int
	Owner      int
	Signing    string
	Encrypting []byte
}
