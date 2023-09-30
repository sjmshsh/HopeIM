package report

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"
)

// Result 存储单词请求的结果，包括状态码，错误，响应时长和内容长度
type Result struct {
	StatusCode    int           // 状态码
	Err           error         // 错误
	Duration      time.Duration // 响应时长
	ContentLength int64         // 内容长度
}

const mapcap = 100000

type Report struct {
	fastest  float64 // 最快响应时间
	slowest  float64 // 最慢响应时间
	average  float64 // 平均响应时间
	rps      float64 // 每秒钟处理的请求数量
	avgTotal float64 // 记录所有请求的响应时间总和

	lats        []float64 // 延迟的缩写
	errorDist   map[string]int
	statusCodes []int
	results     chan *Result
	SizeTotal   int64
	numRes      int
	total       time.Duration // 计算测试的总时长，用来计算RPS
	sizeTotal   int64
	w           io.Writer
}

func New(w io.Writer, n int) *Report {
	if n < 500 {
		n = 500
	}
	cap := min(n, mapcap)

	r := &Report{
		lats:        make([]float64, 0, cap),
		w:           w,
		results:     make(chan *Result, 10),
		errorDist:   make(map[string]int),
		statusCodes: make([]int, 0, cap),
	}
	go r.start()

	return r
}

func (r *Report) Add(res *Result) {
	r.results <- res
}

func (r *Report) Finalize(total time.Duration) {
	close(r.results)

	r.total = total
	r.rps = float64(r.numRes) / r.total.Seconds()
	r.average = r.avgTotal / float64(len(r.lats))
	r.print()
}

func (r *Report) start() {
	// Loop will continue until channel is closed
	for res := range r.results {
		r.numRes++
		if res.Err != nil {
			r.errorDist[res.Err.Error()]++
		} else {
			r.avgTotal += res.Duration.Seconds()
			if len(r.lats) < mapcap {
				r.lats = append(r.lats, res.Duration.Seconds())
				r.statusCodes = append(r.statusCodes, res.StatusCode)
			}
			if res.ContentLength > 0 {
				r.sizeTotal += res.ContentLength
			}
		}
	}
}

func (r *Report) histogram() []Bucket {
	// 表示分成的区间数, 这里假设是4
	bc := 4
	// 用于存储各个区间的边界值
	buckets := make([]float64, bc+1)
	// 存储每个区间内的请求数量
	counts := make([]int, bc+1)
	// 计算每一个区间的宽度
	bs := (r.slowest - r.fastest) / float64(bc)
	for i := 0; i < bc; i++ {
		// 循环计算每个区间的边界值并存储到buckets数组中
		buckets[i] = r.fastest + bs*float64(i)
	}
	buckets[bc] = r.slowest
	var bi int
	var max int
	// 统计请求数量
	for i := 0; i < len(r.lats); {
		if r.lats[i] <= buckets[bi] {
			i++
			counts[bi]++
			if max < counts[bi] {
				max = counts[bi]
			}
		} else if bi < len(buckets)-1 {
			bi++
		}
	}

	// 构建Bucket数组
	res := make([]Bucket, len(buckets))
	for i := 0; i < len(buckets); i++ {
		res[i] = Bucket{
			Mark:      buckets[i],
			Count:     counts[i],
			Frequency: float64(counts[i]) / float64(len(r.lats)),
		}
	}
	return res
}

func (r *Report) snapshot() Snapshot {
	snapshot := Snapshot{
		Average:     r.average,
		Rps:         r.rps,
		Total:       r.total,
		SizeTotal:   r.sizeTotal,
		ErrorDist:   r.errorDist,
		Lats:        make([]float64, len(r.lats)),
		StatusCodes: make([]int, len(r.lats)),
	}

	if len(r.lats) == 0 {
		return snapshot
	}

	copy(snapshot.Lats, r.lats)
	copy(snapshot.StatusCodes, r.statusCodes)

	sort.Float64s(r.lats)
	r.fastest = r.lats[0]
	r.slowest = r.lats[len(r.lats)-1]

	snapshot.Histogram = r.histogram()
	snapshot.LatencyDistribution = r.latencies()

	snapshot.Fastest = r.fastest
	snapshot.Slowest = r.slowest

	statusCodeDist := make(map[int]int, len(snapshot.StatusCodes))
	for _, statusCode := range snapshot.StatusCodes {
		statusCodeDist[statusCode]++
	}
	snapshot.StatusCodeDist = statusCodeDist

	return snapshot
}

// 用于计算特定百分位下的响应时间分布
func (r *Report) latencies() []LatencyDistribution {
	// 定义百分位数列表
	pctls := []int{10, 50, 75, 90, 99}
	// 这个数组用来存储特定百分位对应的响应时间值
	data := make([]float64, len(pctls))
	j := 0
	// i是当前迭代的索引, 表示当前响应时间在r.lats中的位置
	for i := 0; i < len(r.lats) && j < len(pctls); i++ {
		// i * 100将当前位置转换成以百分比为单位的数值，范围是0 - 100
		// len(r.lats)返回响应时间列表r.lats的长度
		// 这样就可以得到该位置在总长度中所占的百分比
		current := i * 100 / len(r.lats)
		if current >= pctls[j] {
			data[j] = r.lats[i]
			j++
		}
	}

	res := make([]LatencyDistribution, len(pctls))
	for i := 0; i < len(pctls); i++ {
		if data[i] > 0 {
			res[i] = LatencyDistribution{Percentage: pctls[i], Latency: data[i]}
		}
	}
	return res
}

func (r *Report) print() {
	buf := &bytes.Buffer{}
	if err := newTemplate().Execute(buf, r.snapshot()); err != nil {
		log.Println("error:", err.Error())
		return
	}
	r.printf(buf.String())

	r.printf("\n")
}

func (r *Report) printf(s string, v ...interface{}) {
	fmt.Fprintf(r.w, s, v...)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Bucket 表示直方图中的一个区间
type Bucket struct {
	Mark      float64 // 标记该区间的值(响应时间的分界点)
	Count     int     // 该区间内的请求数量
	Frequency float64 // 该区间的频率(请求在总请求数中所占比例)
}

// 用于计算响应时间的直方图
func histogram(buckets []Bucket) string {
	max := 0
	// 计算最大请求数量
	for _, b := range buckets {
		if v := b.Count; v > max {
			max = v
		}
	}
	res := new(bytes.Buffer)
	// 生成直方图, 使用找到的最大请求数量来归一化每一个bucket的bar长度, 以确保直方图不失真
	// 这样可以保持比例一致，更好的展现数据的分布情况
	for i := 0; i < len(buckets); i++ {
		// Normalize bar lengths.
		var barLen int
		if max > 0 {
			barLen = (buckets[i].Count*40 + max/2) / max
		}
		res.WriteString(fmt.Sprintf("  %4.3f [%v]\t|%v\n", buckets[i].Mark, buckets[i].Count, strings.Repeat(barChar, barLen)))
	}
	return res.String()
}

// Snapshot 表示基准测试的快照或总结
type Snapshot struct {
	AvgTotal float64 // 所有请求的平均响应时间
	Fastest  float64 // 最快的响应时间
	Slowest  float64 // 最慢的响应时间
	Average  float64 // 平均响应时间
	Rps      float64 // 每秒请求数

	AvgDelay float64
	DelayMax float64
	DelayMin float64

	Lats        []float64
	StatusCodes []int

	Total time.Duration // 总测试时间

	ErrorDist      map[string]int
	StatusCodeDist map[int]int
	SizeTotal      int64
	SizeReq        int64
	NumRes         int64

	// 表示响应时间的中位数分布
	LatencyDistribution []LatencyDistribution
	Histogram           []Bucket // 响应时间直方图
}

// LatencyDistribution 表示响应时间的分位数分布
type LatencyDistribution struct {
	Percentage int     // 百分位数，表示响应时间的位置
	Latency    float64 // 响应时间对应的值
}
