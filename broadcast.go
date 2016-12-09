package goop

type naiiveBroadcast struct {
	b          chan priority
	recipients int
}

func newBroadcast(recipients int) *naiiveBroadcast {
	return &naiiveBroadcast{
		recipients: recipients,
		b:          make(chan priority, recipients*recipients),
	}
}

func (b *naiiveBroadcast) B() <-chan priority {
	return b.b
}

func (b *naiiveBroadcast) Close() {
	close(b.b)
}

func (b *naiiveBroadcast) Broadcast(v priority) {
	for i := 0; i < b.recipients; i++ {
		b.b <- v
	}
}
