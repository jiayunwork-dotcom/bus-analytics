<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">对标分析</div>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="横向线路对比" name="horizontal">
          <div style="margin-bottom: 16px;">
            <el-select
              v-model="selectedLines"
              multiple
              filterable
              placeholder="选择2-5条线路进行对比"
              style="width: 100%; max-width: 600px;"
            >
              <el-option v-for="l in lines" :key="l.line_no" :label="l.line_name" :value="l.line_no" />
            </el-select>
            <el-button type="primary" style="margin-left: 12px;" :disabled="selectedLines.length < 2 || selectedLines.length > 5" @click="doCompare">
              开始对比
            </el-button>
            <span style="margin-left: 12px; color: #909399; font-size: 13px;">
              已选 {{ selectedLines.length }}/5 条线路
            </span>
          </div>

          <div v-if="compareData.length">
            <el-row :gutter="20">
              <el-col :span="14">
                <div class="sub-section-title">分组柱状图对比</div>
                <div ref="barChart" class="chart-container"></div>
              </el-col>
              <el-col :span="10">
                <div class="sub-section-title">雷达图对比</div>
                <div ref="radarChart" class="chart-container"></div>
              </el-col>
            </el-row>

            <div class="sub-section-title" style="margin-top: 20px;">对比数据表</div>
            <el-table :data="compareData" border stripe>
              <el-table-column prop="line_no" label="线路编号" width="100" />
              <el-table-column prop="line_name" label="线路名称" />
              <el-table-column label="客运强度">
                <template #default="{ row }">{{ row.passenger_intensity?.toFixed(2) }}</template>
              </el-table-column>
              <el-table-column label="高峰满载率">
                <template #default="{ row }">{{ (row.peak_load_factor * 100)?.toFixed(1) }}%</template>
              </el-table-column>
              <el-table-column label="营运速度(km/h)">
                <template #default="{ row }">{{ row.operating_speed?.toFixed(2) }}</template>
              </el-table-column>
              <el-table-column label="准点率(%)">
                <template #default="{ row }">{{ row.on_time_rate?.toFixed(2) }}%</template>
              </el-table-column>
            </el-table>

            <div style="margin-top: 20px; text-align: right;">
              <el-button type="primary" :icon="Download" @click="exportReport">导出月度运营报告(PDF)</el-button>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="线路历史纵向对比" name="vertical">
          <div style="margin-bottom: 16px;">
            <el-row :gutter="16">
              <el-col :span="5">
                <el-select v-model="lineNo" placeholder="选择线路" filterable>
                  <el-option v-for="l in lines" :key="l.line_no" :label="l.line_name" :value="l.line_no" />
                </el-select>
              </el-col>
              <el-col :span="5">
                <el-date-picker v-model="startDate" type="date" placeholder="开始日期" value-format="YYYY-MM-DD" />
              </el-col>
              <el-col :span="5">
                <el-date-picker v-model="endDate" type="date" placeholder="结束日期" value-format="YYYY-MM-DD" />
              </el-col>
              <el-col :span="4">
                <el-select v-model="granularity">
                  <el-option label="周维度" value="week" />
                  <el-option label="月维度" value="month" />
                </el-select>
              </el-col>
              <el-col :span="3">
                <el-button type="primary" @click="loadHistorical">查询</el-button>
              </el-col>
            </el-row>
          </div>

          <div v-if="historicalData.length">
            <el-row :gutter="20">
              <el-col :span="12">
                <div class="sub-section-title">客运强度趋势</div>
                <div ref="piHistChart" class="chart-container"></div>
              </el-col>
              <el-col :span="12">
                <div class="sub-section-title">准点率趋势</div>
                <div ref="otHistChart" class="chart-container"></div>
              </el-col>
            </el-row>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '../api'
import * as echarts from 'echarts'

const activeTab = ref('horizontal')
const lines = ref([])
const selectedLines = ref([])
const compareData = ref([])
const barChart = ref(null)
const radarChart = ref(null)

const lineNo = ref('')
const startDate = ref('')
const endDate = ref('')
const granularity = ref('week')
const historicalData = ref([])
const piHistChart = ref(null)
const otHistChart = ref(null)

const doCompare = async () => {
  compareData.value = await api.compareLines(selectedLines.value)
  await nextTick()
  renderBarChart()
  renderRadarChart()
}

const renderBarChart = () => {
  if (!barChart.value) return
  const chart = echarts.init(barChart.value)
  const dims = ['客运强度', '高峰满载率', '营运速度', '准点率']
  const series = compareData.value.map((d, idx) => ({
    name: d.line_no,
    type: 'bar',
    data: [
      +d.passenger_intensity.toFixed(2),
      +(d.peak_load_factor * 100).toFixed(1),
      +d.operating_speed.toFixed(2),
      +d.on_time_rate.toFixed(2)
    ]
  }))
  chart.setOption({
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: { data: compareData.value.map(d => d.line_no) },
    xAxis: { type: 'category', data: dims },
    yAxis: { type: 'value' },
    series
  })
}

const renderRadarChart = () => {
  if (!radarChart.value) return
  const chart = echarts.init(radarChart.value)
  const maxPI = Math.max(...compareData.value.map(d => d.passenger_intensity)) * 1.2
  const maxSpeed = Math.max(...compareData.value.map(d => d.operating_speed)) * 1.2
  chart.setOption({
    tooltip: {},
    legend: { data: compareData.value.map(d => d.line_no) },
    radar: {
      indicator: [
        { name: '客运强度', max: maxPI },
        { name: '高峰满载率', max: 100 },
        { name: '营运速度', max: maxSpeed },
        { name: '准点率', max: 100 }
      ]
    },
    series: [{
      type: 'radar',
      data: compareData.value.map(d => ({
        name: d.line_no,
        value: [
          +d.passenger_intensity.toFixed(2),
          +(d.peak_load_factor * 100).toFixed(1),
          +d.operating_speed.toFixed(2),
          +d.on_time_rate.toFixed(2)
        ]
      }))
    }]
  })
}

const loadHistorical = async () => {
  if (!lineNo.value) return
  historicalData.value = await api.getHistoricalTrend(lineNo.value, {
    start_date: startDate.value,
    end_date: endDate.value,
    granularity: granularity.value
  })
  await nextTick()
  renderHistoricalCharts()
}

const renderHistoricalCharts = () => {
  const periods = historicalData.value.map(d => d.period)
  if (piHistChart.value) {
    echarts.init(piHistChart.value).setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: periods },
      yAxis: { type: 'value', name: '人次/km' },
      series: [{ type: 'line', smooth: true, data: historicalData.value.map(d => d.passenger_intensity), areaStyle: {}, itemStyle: { color: '#409eff' } }]
    })
  }
  if (otHistChart.value) {
    echarts.init(otHistChart.value).setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: periods },
      yAxis: { type: 'value', name: '%', max: 100 },
      series: [{ type: 'line', smooth: true, data: historicalData.value.map(d => d.on_time_rate), areaStyle: {}, itemStyle: { color: '#67c23a' }, markLine: { data: [{ type: 'average' }] } }]
    })
  }
}

const exportReport = async () => {
  const blob = await api.exportReport(selectedLines.value)
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'monthly_report.pdf'
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('报告导出成功')
}

onMounted(async () => {
  lines.value = await api.getLines()
  const dr = await api.getDateRange()
  if (lines.value.length) {
    lineNo.value = lines.value[0].line_no
    selectedLines.value = lines.value.slice(0, Math.min(3, lines.value.length)).map(l => l.line_no)
  }
  if (dr.max_date) endDate.value = dr.max_date
  if (dr.min_date) startDate.value = dr.min_date
  if (selectedLines.value.length >= 2) doCompare()
})
</script>

<style scoped>
.sub-section-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}
</style>
