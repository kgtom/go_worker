package main

import (
	"fmt"
	"sync"
	"time"
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
	//result chan ResultType //可能需要返回处理结果给其他channel
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
	index    int
	done     chan struct{}
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
	// worker := new(worker)

	//job := &job{Id: 1}
	// worker.jobadd = make(chan *jobPair)
	// worker.jobdel = make(chan string)
	worker := &worker{
		jobqueue: make(map[string]*job),
		pending:  new(safepending),
		jobadd:   make(chan *jobPair),
		jobdel:   make(chan string),
		index:    5,
		done:     make(chan struct{}),
	}
	var wg sync.WaitGroup
	worker.Run(&wg)
	job := &job{Id: 1}
	worker.PushJob("tom", job)
	wg.Wait()
}
func (w *worker) Run(wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second * 3)
		defer ticker.Stop()
		for {
			select {
			case jobpair := <-w.jobadd:
				w.insertJob(jobpair.key, jobpair.value)
				fmt.Println("case insert....", jobpair.key)
			case delkey := <-w.jobdel:
				w.deleteJob(delkey)
			case <-ticker.C:
				fmt.Println("timeout")
				return
			default:
				//fmt.Println("end")

			}
			//注意：
			//当case <- jobpair/delkey 和 case <- ticker.C 同时成立时，Select会//随机公平地选出一个执行，有可能选择到前者，导致超时随机行失败
			//解决方案：将【超时】和【收包】放在各自单独的select里面，【超时】一定可以执行到
			// 	 select {
			// case jobpair := <-w.jobadd:
			// 	w.insertJob(jobpair.key, jobpair.value)
			// 	fmt.Println("case insert....", jobpair.key) // if ch not empty, time.After will nerver exec
			//         fmt.Println("sleep one seconds ...")
			//         time.Sleep(1 * time.Second)
			//         fmt.Println("sleep one seconds end...")
			//     default: // forbid block
			// }
			// select {
			//     case <- ticker.C:
			//         fmt.Println("timeout")
			//         return
			//     default: // forbid block
			// }
			//参考：http://studygolang.com/articles/5224
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
