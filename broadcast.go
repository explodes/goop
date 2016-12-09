package operations

type Broadcast interface {
	B() <-chan priority
	Close()
	Broadcast(v priority)
}

func newBroadcast(recipients int) Broadcast {
	return &simpleBroadcast{
		recipients: recipients,
		b:          make(chan priority, recipients * recipients),
	}
}

// todo: this sucks because it relies on a predetermined number of sends to a channel to simulate a broadcast
type simpleBroadcast struct {
	b          chan priority
	recipients int
}

func (b *simpleBroadcast) B() <-chan priority {
	return b.b
}

func (b *simpleBroadcast) Close() {
	close(b.b)
}

func (b *simpleBroadcast) Broadcast(v priority) {
	for i := 0; i < b.recipients; i++ {
		b.b <- v
	}
}
