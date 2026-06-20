<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">数据概览</div>
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="stat-card blue">
            <div class="stat-label">运营线路</div>
            <div class="stat-value">{{ summary.line_count || 0 }}</div>
            <div class="stat-icon"><el-icon><Van /></el-icon></div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card green">
            <div class="stat-label">运营班次</div>
            <div class="stat-value">{{ summary.trip_count || 0 }}</div>
            <div class="stat-icon"><el-icon><Calendar /></el-icon></div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card orange">
            <div class="stat-label">运营车辆</div>
            <div class="stat-value">{{ summary.vehicle_count || 0 }}</div>
            <div class="stat-icon"><el-icon><Van /></el-icon></div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="stat-card purple">
            <div class="stat-label">总乘客数</div>
            <div class="stat-value">{{ formatNumber(summary.total_passengers || 0) }}</div>
            <div class="stat-icon"><el-icon><User /></el-icon></div>
          </div>
        </el-col>
      </el-row>
      <div v-if="summary.date_range && summary.date_range.length" class="date-range">
        数据日期范围：<strong>{{ summary.date_range[0] }}</strong> 至 <strong>{{ summary.date_range[1] }}</strong>
      </div>
    </div>

    <el-row :gutter="20">
      <el-col :span="12">
        <div class="section-card">
          <div class="section-title">线路效率排行</div>
          <el-table :data="topLines" height="400" stripe>
            <el-table-column prop="line_no" label="线路编号" width="100" />
            <el-table-column prop="line_name" label="线路名称" />
            <el-table-column label="客运强度" width="120">
              <template #default="{ row }">
                <span :class="getMetricClass(row.passenger_intensity, avgPI)">
                  {{ row.passenger_intensity?.toFixed(2) }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="准点率(%)" width="110">
              <template #default="{ row }">
                <span :class="getMetricClass(row.on_time_rate, avgOT)">
                  {{ row.on_time_rate?.toFixed(2) }}
                </span>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="section-card">
          <div class="section-title">客运强度分布</div>
          <div ref="chartRef" class="chart-container"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="12">
        <div class="section-card">
          <div class="section-title">高峰满载率对比</div>
          <div ref="loadChartRef" class="chart-container"></div>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="section-card">
          <div class="section-title">车辆利用率概览</div>
          <el-row :gutter="12">
            <el-col :span="8">
              <div class="mini-stat">
                <div class="mini-label">低利用率</div>
                <div class="mini-value red">{{ vehicleStats.lowCount || 0 }}</div>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="mini-stat">
                <div class="mini-label">正常</div>
                <div class="mini-value green">{{ vehicleStats.normalCount || 0 }}</div>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="mini-stat">
                <div class="mini-label">超负荷</div>
                <div class="mini-value orange">{{ vehicleStats.highCount || 0 }}</div>
              </div>
            </el-col>
          </el-row>
          <div ref="utilChartRef" style="height: 300px; margin-top: 20px;"></div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, nextTick } from 'vue'
import api from '../api'
import * as echarts from 'echarts'

const summary = ref({})
const allMetrics = ref([])
const chartRef = ref(null)
const loadChartRef = ref(null)
const utilChartRef = ref(null)
const vehicleStats = ref({ lowCount: 0, normalCount: 0, highCount: 0 })
let dateRange = ref({ max_date: '' })

const topLines = computed(() => {
  return [...allMetrics.value].sort((a, b) => b.passenger_intensity - a.passenger_intensity).slice(0, 10)
})

const avgPI = computed(() => {
  if (!allMetrics.value.length) return 0
  return allMetrics.value.reduce((s, x) => s + x.passenger_intensity, 0) / allMetrics.value.length
})

const avgOT = computed(() => {
  if (!allMetrics.value.length) return 0
  return allMetrics.value.reduce((s, x) => s + x.on_time_rate, 0) / allMetrics.value.length
})

const getMetricClass = (v, avg) => {
  if (v > avg * 1.2) return 'metric-green'
  if (v < avg * 0.7) return 'metric-red'
  return 'metric-yellow'
}

const formatNumber = (n) => {
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return n
}

const loadData = async () => {
  summary.value = await api.getSummary()
  allMetrics.value = await api.getLineEfficiencies()
  dateRange.value = await api.getDateRange()
  
  if (dateRange.value.max_date) {
    const vdata = await api.getVehicleUtilizations(dateRange.value.max_date)
    vehicleStats.value.lowCount = vdata.low_count
    vehicleStats.value.highCount = vdata.high_count
    vehicleStats.value.normalCount = vdata.vehicles.length - vdata.low_count - vdata.high_count
    renderUtilChart(vdata.vehicles)
  }

  await nextTick()
  renderPIChart()
  renderLoadChart()
}

const renderPIChart = () => {
  if (!chartRef.value) return
  const chart = echarts.init(chartRef.value)
  chart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: allMetrics.value.map(m => m.line_no) },
    yAxis: { type: 'value', name: '人次/km' },
    series: [{
      type: 'bar',
      data: allMetrics.value.map(m => +m.passenger_intensity.toFixed(2)),
      itemStyle: { color: '#409eff' },
      label: { show: true, position: 'top', fontSize: 10 }
    }]
  })
}

const renderLoadChart = () => {
  if (!loadChartRef.value) return
  const chart = echarts.init(loadChartRef.value)
  chart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['高峰满载率', '平峰满载率'] },
    xAxis: { type: 'category', data: allMetrics.value.slice(0, 8).map(m => m.line_no) },
    yAxis: { type: 'value', max: 1, axisLabel: { formatter: v => (v * 100) + '%' } },
    series: [
      { name: '高峰满载率', type: 'bar', data: allMetrics.value.slice(0, 8).map(m => +m.peak_load_factor.toFixed(3)), itemStyle: { color: '#f56c6c' } },
      { name: '平峰满载率', type: 'bar', data: allMetrics.value.slice(0, 8).map(m => +m.off_peak_load_factor.toFixed(3)), itemStyle: { color: '#67c23a' } }
    ]
  })
}

const renderUtilChart = (vehicles) => {
  if (!utilChartRef.value) return
  const chart = echarts.init(utilChartRef.value)
  const counts = {}
  vehicles.forEach(v => {
    const t = Math.min(v.daily_trips, 15)
    counts[t] = (counts[t] || 0) + 1
  })
  chart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: Object.keys(counts).map(k => k + '趟') },
    yAxis: { type: 'value', name: '车辆数' },
    series: [{
      type: 'bar',
      data: Object.values(counts),
      itemStyle: { color: '#e6a23c' }
    }]
  })
}

onMounted(loadData)
</script>

<style scoped>
.stat-card {
  padding: 24px;
  border-radius: 8px;
  position: relative;
  overflow: hidden;
  color: white;
  margin-bottom: 20px;
}
.stat-card.blue { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.stat-card.green { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); }
.stat-card.orange { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.stat-card.purple { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
.stat-label { font-size: 14px; opacity: 0.9; }
.stat-value { font-size: 32px; font-weight: 700; margin-top: 8px; }
.stat-icon { position: absolute; right: 16px; top: 50%; transform: translateY(-50%); font-size: 56px; opacity: 0.3; }
.date-range { margin-top: 16px; color: #6b7280; font-size: 14px; }
.mini-stat { text-align: center; padding: 16px; background: #f9fafb; border-radius: 8px; }
.mini-label { color: #6b7280; font-size: 13px; margin-bottom: 8px; }
.mini-value { font-size: 24px; font-weight: 700; }
.mini-value.red { color: #f56c6c; }
.mini-value.green { color: #67c23a; }
.mini-value.orange { color: #e6a23c; }
</style>
