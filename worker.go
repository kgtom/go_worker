package main

import (
	"fmt"
	"sync"
)

type safepending struct {
	pending int
	sync.RWMutex
}

func (s *safepending) Inc() {
	s.Lock()
	s.pending++
	s.Unlock()
}

func (s *safepending) Dec() {
	s.Lock()
	s.pending--
	s.Unlock()
}

type job struct {
	// something
	Id int
}

type jobPair struct {
	key   string
	value *job
}

type worker struct {
	jobqueue map[string]*job // key:UserName
	pending  *safepending
	jobadd   chan *jobPair
	jobdel   chan string
}

// 并不是真正的map insert操作，仅发消息给另外一个线程
func (w *worker) PushJob(user string, job *job) {
	fmt.Println("insert.....", user)
	pair := &jobPair{
		key:   user,
		value: job,
	}
	w.jobadd <- pair
}

// 并不是真正的map delete操作，仅发消息给另外一个线程
func (w *worker) RemoveJob(user string) {
	w.jobdel <- user
}
func main() {
	fmt.Println("main start.....")
	worker := new(worker)
	job := &job{Id: 1}
	var wg sync.WaitGroup
	worker.Run(&wg)
	worker.PushJob("tom", job)
	wg.Wait()
}
func (w *worker) Run(wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case jobpair := <-w.jobadd:
				w.insertJob(jobpair.key, jobpair.value)
				fmt.Println("case insert....", jobpair.key)
			case delkey := <-w.jobdel:
				w.deleteJob(delkey)

			}
		}
	}()

}
func (w *worker) insertJob(key string, value *job) error {
	fmt.Println("worker... run", key)
	w.jobqueue[key] = value
	w.pending.Inc()
	fmt.Println("pending...", w.pending.pending)
	return nil
}

func (w *worker) deleteJob(key string) {
	delete(w.jobqueue, key)
	w.pending.Dec()
}
