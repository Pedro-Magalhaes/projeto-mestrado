package coord

type Producer interface {
	Write([]byte)
}

func BuildProducer() {
}
