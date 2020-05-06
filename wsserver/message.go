package wsserver

// Message - slice bytes
//[]byte
type Message []byte

// Messages struc for queue messages for/from a client
//  type []Template
type Messages []Message

// AddMessage - func AddMessage(mtype int, message string)
// add new messages in array
func (m *Messages) AddMessage(message Message) {
	*m = append(m.GetMessages(), message)
}

// GetMessages - func (ver *Messages) GetMessages() Messages
// get array of messages
func (m *Messages) GetMessages() Messages {
	return *m
}

// DelFirstM - func (ver *Messages) DelFirstM()
// delete first messages in array
func (m *Messages) DelFirstM() {
	if len(*m) >= 2 {
		new := m.GetMessages()
		*m = append(new[1:2], new[2:]...)
	} else {
		*m = make(Messages, 0, 100)
	}
}
