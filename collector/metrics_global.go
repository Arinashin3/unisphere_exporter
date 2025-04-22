package collector

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strconv"
	"strings"
	"sync"
	"unisphere_exporter/types"
	"unisphere_exporter/utils"
)

//type HTTPClient interface {
//	Do(req *http.Request) (*http.Response, error)
//}

func MetricQuery(c *utils.UnisphereClient, p []string, i int) (int, float64) {
	var b types.MetricPath
	var req types.MetricQueryResult

	b.Paths = p
	b.Interval = i
	body, err := json.Marshal(b)
	if err != nil {
		log.Printf("Unmarshal Body %v", err)
		return 0, 0.0
	}
	if err := c.Post("/api/types/metricRealTimeQuery", "compact=true", body, &req); err != nil {
		log.Printf("Post Body %v", err)
		return 0, 0.0
	}
	for _, entry := range req.Entries {
		return entry.Content.Id, 1.0
	}
	return 0, 0.0
}

func MetricQueryResult(c *utils.UnisphereClient, qid int) (types.MetricQueryResult, float64) {
	var qresult types.MetricQueryResult
	err := c.Get("/api/types/metricQueryResult", "compact=true&filter=queryId eq "+strconv.Itoa(qid), &qresult)
	if err != nil {
		log.Printf("Get Body %v", err)
		return qresult, 0.0
	}
	return qresult, 1.0

}

func ProbeMetricGlobal(c utils.UnisphereClient, registry *prometheus.Registry, wg *sync.WaitGroup) float64 {
	// Variable qr is return value
	defer wg.Done()
	qr := 0.0
	labels := []string{"sp_name"}
	interval := 60

	paths := []string{
		"sp.*.blockCache.global.summary.cleanPages",
		"sp.*.blockCache.global.summary.dirtyBytes",
		"sp.*.blockCache.global.summary.dirtyPages",
		"sp.*.blockCache.global.summary.flushedBlocks",
		"sp.*.blockCache.global.summary.flushes",
		"sp.*.blockCache.global.summary.freePages",
		"sp.*.blockCache.global.summary.maxPages",
		"sp.*.blockCache.global.summary.readHits",
		"sp.*.blockCache.global.summary.readHitsRate",
		"sp.*.blockCache.global.summary.readMisses",
		"sp.*.blockCache.global.summary.readMissesRate",
		"sp.*.blockCache.global.summary.writeHits",
		"sp.*.blockCache.global.summary.writeHitsRate",
		"sp.*.blockCache.global.summary.writeMisses",
		"sp.*.blockCache.global.summary.writeMissesRate",
		"sp.*.cpu.summary.busyTicks",
		"sp.*.cpu.summary.utilization",
		"sp.*.cpu.summary.waitTicks",
		"sp.*.cpu.uptime",
		"sp.*.fibreChannel.blockSize",
		"sp.*.net.basic.inBytes",
		"sp.*.net.basic.inBytesRate",
		"sp.*.net.basic.outBytes",
		"sp.*.net.basic.outBytesRate",
		"sp.*.memory.bufferCache.freeBufferBytes",
		"sp.*.memory.bufferCache.highWatermarkHits",
		"sp.*.memory.bufferCache.hits",
		"sp.*.memory.bufferCache.lookups",
		"sp.*.memory.bufferCache.lowWatermarkHits",
		"sp.*.memory.bufferCache.watermarkHits",
		"sp.*.memory.pageSize",
		"sp.*.memory.summary.freeBytes",
		"sp.*.memory.summary.swapFreeBytes",
		"sp.*.memory.summary.swapTotalUsedBytes",
		"sp.*.memory.summary.totalBytes",
		"sp.*.memory.summary.totalUsedBytes",
		"sp.*.platform.storageProcessorTemperature",
		"sp.*.cifs.global.basic.readAvgSize",
		"sp.*.cifs.global.basic.readBytes",
		"sp.*.cifs.global.basic.readBytesRate",
		"sp.*.cifs.global.basic.reads",
		"sp.*.cifs.global.basic.readsRate",
		"sp.*.cifs.global.basic.totalCalls",
		"sp.*.cifs.global.basic.totalCallsRate",
		"sp.*.cifs.global.basic.writeAvgSize",
		"sp.*.cifs.global.basic.writeBytes",
		"sp.*.cifs.global.basic.writeBytesRate",
		"sp.*.cifs.global.basic.writes",
		"sp.*.cifs.global.basic.writesRate",
		"sp.*.cifs.global.usage.currentConnections",
		"sp.*.cifs.global.usage.currentOpenFiles",
		"sp.*.nfs.basic.readAvgSize",
		"sp.*.nfs.basic.readBytes",
		"sp.*.nfs.basic.readBytesRate",
		"sp.*.nfs.basic.reads",
		"sp.*.nfs.basic.readsRate",
		"sp.*.nfs.basic.writeAvgSize",
		"sp.*.nfs.basic.writeBytes",
		"sp.*.nfs.basic.writeBytesRate",
		"sp.*.nfs.basic.writes",
		"sp.*.nfs.basic.writesRate",
		"sp.*.nfs.currentThreads",
		"sp.*.nfs.maxUsedThreads",
		"sp.*.nfs.totalCalls",
		"sp.*.nfs.totalCallsRate",
		"sp.*.storage.filesystemSummary.readBytes",
		"sp.*.storage.filesystemSummary.reads",
		"sp.*.storage.filesystemSummary.writeBytes",
		"sp.*.storage.filesystemSummary.writes",
	}

	qid, qr := MetricQuery(&c, paths, interval)
	if qr == 0.0 {
		return qr
	}
	qresult, qr := MetricQueryResult(&c, qid)
	if qr == 0.0 {
		return qr
	}

	for _, entry := range qresult.Entries {
		p := strings.ReplaceAll(entry.Content.Path, "sp.*.", "")
		p = strings.ReplaceAll(p, ".", "_")
		p = strings.ToLower("unisphere_" + p)

		mSpMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: p, Help: "Full Path is " + entry.Content.Path}, labels)
		registry.Register(mSpMetric)

		for k, v := range entry.Content.Values {
			var f float64

			switch i := v.(type) {
			case int:
				f = float64(i)
			case float32:
				f = float64(i)
			case float64:
				f = i
			case bool:
				if i == true {
					f = 1.0
				} else {
					f = 0.0
				}
			}
			mSpMetric.WithLabelValues(k).Set(f)
		}
	}

	return 1.0
}
