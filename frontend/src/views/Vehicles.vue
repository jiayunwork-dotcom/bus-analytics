<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">车辆周转分析</div>
      <el-row :gutter="16" style="margin-bottom: 16px;">
        <el-col :span="6">
          <el-date-picker v-model="selectedDate" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" @change="loadData" />
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="8">
          <div class="stat-box low">
            <div class="stat-num">{{ data.low_count || 0 }}</div>
            <div class="stat-text">低利用率车辆 (趟次<均值50%)</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-box normal">
            <div class="stat-num">{{ data.vehicles?.length - (data.low_count || 0) - (data.high_count || 0) }}</div>
            <div class="stat-text">正常车辆</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-box high">
            <div class="stat-num">{{ data.high_count || 0 }}</div>
            <div class="stat-text">超负荷车辆 (趟次>均值150%)</div>
          </div>
        </el-col>
      </el-row>

      <div style="margin-top: 16px; text-align: center; color: #606266;">
        全网日均趟次：<strong>{{ avgTrips?.toFixed(1) }}</strong> 趟
      </div>
    </div>

    <el-row :gutter="20">
      <el-col :span="12">
        <div class="section-card">
          <div class="section-title">车辆趟次分布直方图</div>
          <div ref="histChart" class="chart-container"></div>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="section-card">
          <div class="section-title">单车甘特图</div>
          <el-select v-model="selectedVehicle" placeholder="选择车辆" filterable style="width: 200px; margin-bottom: 12px;" @change="loadGantt">
            <el-option v-for="v in data.vehicles || []" :key="v.vehicle_no" :label="v.vehicle_no" :value="v.vehicle_no" />
          </el-select>
          <div ref="ganttChart" style="height: 300px;"></div>
        </div>
      </el-col>
    </el-row>

    <div class="section-card">
      <div class="section-title">车辆利用率明细</div>
      <el-table :data="data.vehicles || []" stripe>
        <el-table-column prop="vehicle_no" label="车辆编号" width="140" />
        <el-table-column prop="daily_trips" label="日趟次" width="100" align="center" sortable>
          <template #default="{ row }">
            <span :class="getUtilClass(row)">{{ row.daily_trips }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="first_trip_time" label="首班时间" width="120" />
        <el-table-column prop="last_trip_time" label="末班时间" width="120" />
        <el-table-column prop="operating_km" label="日营运里程(km)" width="140" align="right" sortable />
        <el-table-column label="利用率评级" width="120" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.level === 'low'" type="danger">低利用率</el-tag>
            <el-tag v-else-if="row.level === 'high'" type="warning">超负荷</el-tag>
            <el-tag v-else type="success">正常</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import api from '../api'
import * as echarts from 'echarts'

const selectedDate = ref('')
const selectedVehicle = ref('')
const data = ref({})
const avgTrips = ref(0)
const histChart = ref(null)
const ganttChart = ref(null)

const getUtilClass = (row) => {
  if (row.level === 'low') return 'metric-red'
  if (row.level === 'high') return 'metric-yellow'
  return 'metric-green'
}

const loadData = async () => {
  if (!selectedDate.value) return
  data.value = await api.getVehicleUtilizations(selectedDate.value)
  avgTrips.value = data.value.avg_trips || 0
  selectedVehicle.value = data.value.vehicles?.[0]?.vehicle_no || ''
  await nextTick()
  renderHistChart()
  if (selectedVehicle.value) loadGantt()
}

const loadGantt = async () => {
  if (!selectedVehicle.value) return
  const gantt = await api.getVehicleGantt(selectedVehicle.value, selectedDate.value)
  renderGantt(gantt)
}

const renderHistChart = () => {
  if (!histChart.value) return
  const chart = echarts.init(histChart.value)
  const counts = {}
  ;(data.value.vehicles || []).forEach(v => {
    const t = Math.min(v.daily_trips, 15)
    counts[t] = (counts[t] || 0) + 1
  })
  chart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: Object.keys(counts).map(k => k + '趟') },
    yAxis: { type: 'value', name: '车辆数' },
    series: [{ type: 'bar', data: Object.values(counts), itemStyle: { color: '#e6a23c' }, label: { show: true, position: 'top' } }]
  })
}

const renderGantt = (gantt) => {
  if (!ganttChart.value) return
  const chart = echarts.init(ganttChart.value)
  const categories = [...new Set(gantt.map(g => g.line_no))]
  chart.setOption({
    tooltip: { formatter: p => `${p.value[3]}<br/>${p.value[1]} - ${p.value[2]}` },
    xAxis: {
      type: 'time',
      min: '2024-01-01 05:00',
      max: '2024-01-01 23:00',
      axisLabel: { formatter: v => { const d = new Date(v); return d.getHours() + ':' + String(d.getMinutes()).padStart(2, '0') } }
    },
    yAxis: { type: 'category', data: categories },
    series: [{
      type: 'custom',
      renderItem: (params, api) => {
        const catIdx = categories.indexOf(api.value(3))
        const start = api.coord([api.value(0), catIdx])
        const end = api.coord([api.value(1), catIdx])
        return {
          type: 'rect',
          shape: { x: start[0], y: start[1] - 10, width: end[0] - start[0], height: 20 },
          style: { fill: api.value(2) === 1 ? '#409eff' : '#67c23a' }
        }
      },
      encode: { x: [0, 1], y: 3 },
      data: gantt.map(g => ['2024-01-01 ' + g.start_time, '2024-01-01 ' + g.end_time, g.direction, g.line_no])
    }]
  })
}

onMounted(async () => {
  const dr = await api.getDateRange()
  if (dr.max_date) {
    selectedDate.value = dr.max_date
    loadData()
  }
})
</script>

<style scoped>
.stat-box {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  color: white;
}
.stat-box.low { background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%); }
.stat-box.normal { background: linear-gradient(135deg, #a1c4fd 0%, #c2e9fb 100%); }
.stat-box.high { background: linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%); }
.stat-num { font-size: 32px; font-weight: 700; }
.stat-text { font-size: 13px; margin-top: 4px; opacity: 0.9; }
</style>
