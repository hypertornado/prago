package main

/*type Queue struct {
	maxcount int64
	ar       []chan chan bool
}

func NewQueue(count int64) *Queue {
	return &Queue{count: count}
}

func (q *Queue) Wait() chan chan bool {
	ret := make(chan chan bool)

	go func() {
		c := make(chan bool)
		ret <- c
		<-c
	}()

	q.ar = append(q.ar, ret)
	return ret
}

func (q *Queue) distribute() {

}*/
