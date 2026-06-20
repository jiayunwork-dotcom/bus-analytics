<template>
  <div class="page-container">
    <div class="page-header">
      <el-button :icon="ArrowLeft" @click="$router.back()">返回</el-button>
      <h2>线路详情 - {{ lineNo }}</h2>
    </div>

    <div class="section-card">
      <div class="section-title">效率指标日趋势</div>
      <el-row :gutter="16" style="margin-bottom: 16px;">
        <el-col :span="5">
          <el-date-picker
            v-model="startDate"
            type="date"
            placeholder="开始日期"
            value-format="YYYY-MM-DD"
          />
        </el-col>
        <el-col :span="5">
          <el-date-picker
            v-model="endDate"
            type="date"
            placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="loadData">查询</el-button>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="12">
          <div class="sub-title">客运强度趋势</div>
          <div ref="piChart" class="chart-container"></div>
        </el-col>
        <el-col :span="12">
          <div class="sub-title">满载率趋势</div>
          <div ref="loadChart" class="chart-container"></div>
        </el-col>
      </el-row>
      <el-row :gutter="16" style="margin-top: 20px;">
        <el-col :span="12">
          <div class="sub-title">营运速度趋势 (km/h)</div>
          <div ref="speedChart" class="chart-container"></div>
        </el-col>
        <el-col :span="12">
          <div class="sub-title">准点率趋势 (%)</div>
          <div ref="otChart" class="chart-container"></div>
        </el-col>
      </el-row>
    </div>

    <div class="section-card">
      <div class="section-title">日度数据明细</div>
      <el-table :data="trendData" stripe border>
        <el-table-column label="日期">
          <template #default="{ row }">{{ extractDate(row.line_name) }}</template>
        </el-table-column>
        <el-table-column label="客运强度">
          <template #default="{ row }">{{ row.passenger_intensity?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="高峰满载率">
          <template #default="{ row }">{{ (row.peak_load_factor * 100)?.toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column label="平峰满载率">
          <template #default="{ row }">{{ (row.off_peak_load_factor * 100)?.toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column label="营运速度(km/h)">
          <template #default="{ row }">{{ row.operating_speed?.toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="准点率(%)">
          <template #default="{ row }">{{ row.on_time_rate?.toFixed(2) }}%</template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import api from '../api'
import * as echarts from 'echarts'

const route = useRoute()
const lineNo = route.params.lineNo

const startDate = ref('')
const endDate = ref('')
const trendData = ref([])
const piChart = ref(null)
const loadChart = ref(null)
const speedChart = ref(null)
const otChart = ref(null)

const extractDate = (name) => {
  const m = name.match(/\(([^)]+)\)/)
  return m ? m[1] : name
}

const loadData = async () => {
  trendData.value = await api.getLineDailyTrend(lineNo, {
    start_date: startDate.value,
    end_date: endDate.value
  })
  await nextTick()
  renderCharts()
}

const renderCharts = () => {
  const dates = trendData.value.map(d => extractDate(d.line_name))

  if (piChart.value) {
    echarts.init(piChart.value).setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', name: '人次/km' },
      series: [{ type: 'line', smooth: true, data: trendData.value.map(d => +d.passenger_intensity.toFixed(2)), areaStyle: {}, itemStyle: { color: '#409eff' } }]
    })
  }
  if (loadChart.value) {
    echarts.init(loadChart.value).setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['高峰', '平峰'] },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 1, axisLabel: { formatter: v => (v*100)+'%' } },
      series: [
        { name: '高峰', type: 'line', smooth: true, data: trendData.value.map(d => +d.peak_load_factor.toFixed(3)), itemStyle: { color: '#f56c6c' } },
        { name: '平峰', type: 'line', smooth: true, data: trendData.value.map(d => +d.off_peak_load_factor.toFixed(3)), itemStyle: { color: '#67c23a' } }
      ]
    })
  }
  if (speedChart.value) {
    echarts.init(speedChart.value).setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', name: 'km/h' },
      series: [{ type: 'line', smooth: true, data: trendData.value.map(d => +d.operating_speed.toFixed(2)), areaStyle: {}, itemStyle: { color: '#e6a23c' } }]
    })
  }
  if (otChart.value) {
    echarts.init(otChart.value).setOption({
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, name: '%' },
      series: [{ type: 'line', smooth: true, data: trendData.value.map(d => +d.on_time_rate.toFixed(2)), areaStyle: {}, itemStyle: { color: '#909399' }, markLine: { data: [{ type: 'average', name: '均值' }] } }]
    })
  }
}

onMounted(loadData)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}
.page-header h2 {
  margin: 0;
}
.sub-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}
</style>
