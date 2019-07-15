package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TH = 6
)

// сюда писать код
func main2() {

	var ok = true
	var recieved uint32
	freeFlowJobs := []job{
		job(func(in, out chan interface{}) {
			out <- 1
			time.Sleep(10 * time.Millisecond)
			currRecieved := atomic.LoadUint32(&recieved)
			// в чем тут суть
			// если вы накапливаете значения, то пока вся функция не отрабоатет - дальше они не пойдут
			// тут я проверяю, что счетчик увеличился в следующей функции
			// это значит что туда дошло значение прежде чем текущая функция отработала
			if currRecieved == 0 {
				ok = false
			}
		}),
		job(func(in, out chan interface{}) {
			for _ = range in {
				atomic.AddUint32(&recieved, 1)
			}
		}),
	}
	ExecutePipeline(freeFlowJobs...)
	if !ok || recieved == 0 {
		fmt.Println("no value free flow - dont collect them")
	}
}
func main() {
	inputData := []int{0, 1, 1, 2, 3, 5}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
			}
			fmt.Println(data)
		}),
	}

	start := time.Now()
	ExecutePipeline(hashSignJobs...)
	end := time.Since(start)
	fmt.Println("Time passed from: ", end)
}

func ExecutePipeline(jobs ...job) {

	wg := &sync.WaitGroup{}

	input := make(chan interface{})

	for _, jobFunc := range jobs {

		wg.Add(1)
		output := make(chan interface{}, TH)
		go startWorker(jobFunc, input, output, wg)
		input = output
	}

	wg.Wait()

}

func startWorker(jobWorker job, in, out chan interface{}, wg *sync.WaitGroup) {

	defer wg.Done()
	defer close(out)
	jobWorker(in, out)

}

// Calculates crc32(data)+"~"+crc32(md5(data)) two strins concatenations with ~
// where data - input information (numbers) from the previous function.
func SingleHash(input, output chan interface{}) {

	start := time.Now()

	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for in := range input {
		wg.Add(1)
		go singleHashWorker(in, output, wg, mu)

	}
	wg.Wait()

	end := time.Since(start)
	fmt.Println("Time SingleHash: ", end)
}

func singleHashWorker(in interface{}, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {

	defer wg.Done()
	data := strconv.Itoa(in.(int)) // reads int data from an input

	Md532Chan := make(chan string)
	go asyncMd5(data, Md532Chan, mu)
	Md5Data := <-Md532Chan

	crc32Chan := make(chan string)
	go asyncCrc32(data, crc32Chan)

	crc32Md5Chan := make(chan string)
	go asyncCrc32(Md5Data, crc32Md5Chan)

	out <- <-crc32Chan + "~" + <-crc32Md5Chan

}

func asyncCrc32(data string, out chan string) {
	out <- DataSignerCrc32(data)
}

func asyncMd5(data string, out chan string, mu *sync.Mutex) {
	// DataSignerMd5 can be called only once at the same time. Executes about 10 msec.
	// If it is called several times simultaneously,  it will be overheat in 1 sec.
	mu.Lock()
	Md5Data := DataSignerMd5(data)
	mu.Unlock()
	out <- Md5Data
}

// MultiHash calculates crc32(th+data))
// (number cast to string and string concatenation),
// where th=0..5 (those 6 hashes for an each input value ),
// then get a result concatenation in a calcuation order (0..5),
// where data - comes from input (and passed to output from SingleHash)
func MultiHash(in, out chan interface{}) {

	start := time.Now()

	wg := &sync.WaitGroup{}
	dataArray := make([][]string, TH, TH)
	i := 0
	for input := range in {
		wg.Add(1)
		if i >= TH {
			dataArray = append(dataArray, make([]string, TH, TH))
		} else {
			dataArray[i] = make([]string, TH, TH)
		}
		go multiHashWorker(input.(string), out, dataArray[i], wg)
		i++
	}

	wg.Wait()

	for i = 0; i < len(dataArray); i++ {
		wg.Add(1)
		go func(dataArray []string, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			out <- strings.Join(dataArray, "")
		}(dataArray[i], out, wg)
	}
	wg.Wait()
	end := time.Since(start)
	fmt.Println("Time MultiHash: ", end)
}

func multiHashWorker(input string, out chan interface{}, dataArray []string, wg *sync.WaitGroup) {

	defer wg.Done()

	for i := 0; i < TH; i++ {
		wg.Add(1)
		go func(input string, dataArray []string, i int, wg *sync.WaitGroup) {
			defer wg.Done()
			data := strconv.Itoa(i) + input
			dataArray[i] = DataSignerCrc32(data)
		}(input, dataArray, i, wg)
	}

}

func CombineResults(in, out chan interface{}) {
	var result []string
	for input := range in {
		result = append(result, input.(string))
	}

	sort.Strings(result)
	out <- strings.Join(result, "_")
}
