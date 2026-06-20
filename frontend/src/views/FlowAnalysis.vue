<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">客流分析</div>
      <el-row :gutter="16" style="margin-bottom: 16px;">
        <el-col :span="6">
          <el-select v-model="selectedLine" placeholder="选择线路" filterable @change="loadData">
            <el-option v-for="l in lines" :key="l.line_no" :label="l.line_name" :value="l.line_no" />
          </el-select>
        </el-col>
        <el-col :span="6">
          <el-date-picker v-model="selectedDate" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" @change="loadData" />
        </el-col>
      </el-row>
    </div>

    <el-row :gutter="20">
      <el-col :span="12">
        <div class="section-card">
          <div class="section-title">断面客流图
            <el-tag v-if="maxSection" type="danger" style="margin-left: 12px;">
              最大断面: {{ maxSection.from_station }}→{{ maxSection.to_station }} ({{ maxSection.passengers }}人)
            </el-tag>
          </div>
          <div ref="sectionChart" class="chart-container"></div>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="section-card">
          <div class="section-title">客流时段分布</div>
          <div ref="hourlyChart" class="chart-container"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :span="8">
        <div class="section-card">
          <div class="section-title">潮汐现象识别</div>
          <div v-if="tidalInfo" class="tidal-info">
            <div class="tidal-badge" :class="{ 'is-tidal': tidalInfo.is_tidal }">
              <el-icon v-if="tidalInfo.is_tidal"><WarningFilled /></el-icon>
              <el-icon v-else><CircleCheck /></el-icon>
              {{ tidalInfo.is_tidal ? '潮汐线路' : '非潮汐线路' }}
            </div>
            <div v-if="tidalInfo.is_tidal" class="tidal-detail">
              <p>方向特征：<strong>{{ tidalInfo.tidal_direction }}</strong></p>
              <p>对向空驶率：<strong style="color: #f56c6c;">{{ tidalInfo.empty_rate?.toFixed(1) }}%</strong></p>
            </div>
            <div v-else class="tidal-detail">
              <p>上下行客流较为均衡</p>
            </div>
          </div>
        </div>
      </el-col>
      <el-col :span="16">
        <div class="section-card">
          <div class="section-title">
            OD推断 (基于刷卡数据)
            <el-tag type="success" style="margin-left: 12px;">
              推断成功率: {{ odSuccessRate?.toFixed(1) }}%
            </el-tag>
          </div>
          <el-table :data="odPairs.slice(0, 10)" stripe size="small">
            <el-table-column type="index" label="排名" width="60" align="center" />
            <el-table-column prop="origin" label="上车站点" />
            <el-table-column label="下车站点">
              <template #default="{ row }">{{ row.destination || '未识别' }}</template>
            </el-table-column>
            <el-table-column prop="count" label="人次" width="100" align="center" />
          </el-table>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import api from '../api'
import * as echarts from 'echarts'

const lines = ref([])
const selectedLine = ref('')
const selectedDate = ref('')
const sectionChart = ref(null)
const hourlyChart = ref(null)
const tidalInfo = ref(null)
const odSuccessRate = ref(0)
const odPairs = ref([])
const maxSection = ref(null)

const loadData = async () => {
  if (!selectedLine.value || !selectedDate.value) return

  const sectionData = await api.getSectionFlow(selectedLine.value, selectedDate.value)
  const hourly = await api.getHourlyDistribution(selectedLine.value, selectedDate.value)
  tidalInfo.value = await api.checkTidalPattern(selectedLine.value, selectedDate.value)
  const odData = await api.inferOD(selectedLine.value, selectedDate.value)
  odSuccessRate.value = odData.success_rate
  odPairs.value = odData.od_pairs

  findMaxSection(sectionData.sections)

  await nextTick()
  renderSectionChart(sectionData.sections)
  renderHourlyChart(hourly)
}

const findMaxSection = (sections) => {
  let max = null
  Object.keys(sections).forEach(dir => {
    sections[dir].forEach(s => {
      if (!max || s.passengers > max.passengers) {
        max = { ...s, direction: +dir }
      }
    })
  })
  maxSection.value = max
}

const renderSectionChart = (sections) => {
  if (!sectionChart.value) return
  const up = sections['1'] || []
  const down = sections['2'] || []
  const seqs = [...new Set([...up.map(s => s.station_seq), ...down.map(s => s.station_seq)])].sort((a, b) => a - b)
  const names = {}
  up.forEach(s => { names[s.station_seq] = s.from_station })
  down.forEach(s => { if (!names[s.station_seq]) names[s.station_seq] = s.from_station })

  const chart = echarts.init(sectionChart.value)
  chart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['上行', '下行'] },
    xAxis: { type: 'category', data: seqs.map(s => names[s] || '站' + s) },
    yAxis: { type: 'value', name: '在车人数' },
    series: [
      { name: '上行', type: 'line', smooth: true, data: seqs.map(s => up.find(x => x.station_seq === s)?.passengers || 0), itemStyle: { color: '#409eff' }, areaStyle: { opacity: 0.2 } },
      { name: '下行', type: 'line', smooth: true, data: seqs.map(s => down.find(x => x.station_seq === s)?.passengers || 0), itemStyle: { color: '#f56c6c' }, areaStyle: { opacity: 0.2 } }
    ]
  })
}

const renderHourlyChart = (hourly) => {
  if (!hourlyChart.value) return
  const chart = echarts.init(hourlyChart.value)
  chart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: hourly.map(h => h.hour + ':00') },
    yAxis: { type: 'value', name: '上车人数' },
    series: [{
      type: 'bar',
      data: hourly.map(h => h.boarders),
      itemStyle: {
        color: (p) => {
          const hr = p.dataIndex
          if ((hr >= 7 && hr < 9) || (hr >= 17 && hr < 19)) return '#f56c6c'
          return '#67c23a'
        }
      },
      markArea: {
        itemStyle: { color: 'rgba(245, 108, 108, 0.08)' },
        data: [[{ xAxis: '7:00' }, { xAxis: '9:00' }], [{ xAxis: '17:00' }, { xAxis: '19:00' }]]
      }
    }]
  })
}

onMounted(async () => {
  lines.value = await api.getLines()
  const dr = await api.getDateRange()
  if (lines.value.length) selectedLine.value = lines.value[0].line_no
  if (dr.max_date) selectedDate.value = dr.max_date
  if (selectedLine.value && selectedDate.value) loadData()
})
</script>

<style scoped>
.tidal-info { text-align: center; padding: 20px 0; }
.tidal-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border-radius: 24px;
  background: #f0f9eb;
  color: #67c23a;
  font-size: 18px;
  font-weight: 600;
}
.tidal-badge.is-tidal {
  background: #fef0f0;
  color: #f56c6c;
}
.tidal-detail {
  margin-top: 16px;
  color: #606266;
}
.tidal-detail p { margin: 8px 0; }
</style>
