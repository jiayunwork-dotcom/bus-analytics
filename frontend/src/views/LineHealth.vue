<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">线路健康度评分</div>
      <div class="score-legend">
        <span class="legend-item"><span class="legend-color legend-green"></span>优秀 (≥80分)</span>
        <span class="legend-item"><span class="legend-color legend-yellow"></span>良好 (60-80分)</span>
        <span class="legend-item"><span class="legend-color legend-red"></span>待提升 (&lt;60分)</span>
      </div>

      <div class="sub-section-title" style="margin-top: 16px;">所有线路得分排名</div>
      <div ref="barChart" class="ranking-chart-container"></div>

      <div v-if="!compareMode" style="margin-top: 24px;">
        <div class="sub-section-title" v-if="selectedLine">
          {{ selectedLine.line_name }} ({{ selectedLine.line_no }}) - 五维健康度详情
          <el-tag :type="tagType(selectedLine.score_level)" size="small" style="margin-left: 12px;">
            {{ selectedLine.total_score.toFixed(2) }} 分
          </el-tag>
        </div>
        <el-alert v-else type="info" show-icon :closable="false" style="margin-bottom: 16px;">
          点击上方任意线路横条，查看该线路的五维健康度详情
        </el-alert>

        <el-row :gutter="20" v-if="selectedLine">
          <el-col :span="12">
            <div class="sub-section-title">五维雷达图</div>
            <div ref="radarChart" class="chart-container"></div>
          </el-col>
          <el-col :span="12">
            <div class="sub-section-title">子指标明细</div>
            <el-table :data="getSubScoreList(selectedLine)" border stripe>
              <el-table-column prop="name" label="指标" width="110" />
              <el-table-column prop="raw_value_str" label="原始值" />
              <el-table-column label="得分" width="100">
                <template #default="{ row }">
                  <span style="font-weight: 600;">{{ row.score.toFixed(2) }}</span>
                  <span style="color: #909399;">/{{ row.max_score }}</span>
                </template>
              </el-table-column>
              <el-table-column label="得分占比" width="130">
                <template #default="{ row }">
                  <el-progress :percentage="row.score_ratio" :stroke-width="12" />
                </template>
              </el-table-column>
            </el-table>
          </el-col>
        </el-row>
      </div>

      <div v-else style="margin-top: 24px;">
        <div class="sub-section-title">对比线路选择 <span style="color:#909399; font-weight:normal; font-size:13px;">（勾选2-4条线路）</span></div>
        <div class="checkbox-group">
          <el-checkbox
            v-for="item in healthScores"
            :key="item.line_no"
            v-model="compareCheckMap[item.line_no]"
            :label="item.line_no"
            @change="onCompareCheckChange"
          >
            {{ item.line_name }} ({{ item.total_score.toFixed(1) }}分)
          </el-checkbox>
        </div>
        <div v-if="Object.values(compareCheckMap).filter(Boolean).length >= 2" style="margin-top: 24px;">
          <el-row :gutter="20">
            <el-col :span="12">
              <div class="sub-section-title">多线路五维雷达图对比</div>
              <div ref="compareRadarChart" class="chart-container"></div>
            </el-col>
            <el-col :span="12">
              <div class="sub-section-title">子指标得分对比</div>
              <el-table :data="compareTableData" border stripe>
                <el-table-column prop="name" label="指标" width="110" fixed />
                <el-table-column
                  v-for="ln in compareLineNos"
                  :key="ln"
                  :label="getLineName(ln)"
                  align="center"
                >
                  <template #default="{ row }">
                    <div>
                      <div style="font-weight: 600;">
                        {{ row[ln + '_score'].toFixed(2) }}
                        <span style="color: #909399; font-weight: normal;">/{{ row.max_score }}</span>
                      </div>
                      <div style="font-size: 12px; color: #606266;">
                        {{ row[ln + '_raw'] }}
                      </div>
                      <el-progress
                        :percentage="row[ln + '_ratio']"
                        :stroke-width="6"
                        :color="getLineColor(ln)"
                        style="margin-top: 4px;"
                      />
                    </div>
                  </template>
                </el-table-column>
              </el-table>

              <div class="sub-section-title" style="margin-top: 20px;">总分对比</div>
              <el-table :data="compareTotalTableData" border stripe>
                <el-table-column prop="line_name" label="线路" width="150" />
                <el-table-column label="总分">
                  <template #default="{ row }">
                    <span style="font-weight: 600; font-size: 16px;">{{ row.total_score.toFixed(2) }}</span>
                  </template>
                </el-table-column>
                <el-table-column prop="score_level" label="等级" width="100">
                  <template #default="{ row }">
                    <el-tag :type="tagType(row.score_level)" size="small">
                      {{ levelText(row.score_level) }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
            </el-col>
          </el-row>
        </div>
        <el-empty v-else description="请至少勾选2条线路进行对比" />
      </div>

      <div style="margin-top: 32px; text-align: right;">
        <el-button
          v-if="!compareMode"
          type="primary"
          :icon="Histogram"
          @click="enterCompareMode"
        >
          进入对比模式
        </el-button>
        <el-button
          v-else
          @click="exitCompareMode"
        >
          退出对比模式
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Histogram } from '@element-plus/icons-vue'
import api from '../api'
import * as echarts from 'echarts'

const healthScores = ref([])
const selectedLine = ref(null)
const barChart = ref(null)
const radarChart = ref(null)
const compareRadarChart = ref(null)

const compareMode = ref(false)
const compareCheckMap = reactive({})
const compareLineNos = ref([])

const lineColors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399']
const lineColorMap = {}

const enterCompareMode = () => {
  compareMode.value = true
  selectedLine.value = null
}

const exitCompareMode = () => {
  compareMode.value = false
  compareLineNos.value = []
  Object.keys(compareCheckMap).forEach(k => compareCheckMap[k] = false)
}

const onCompareCheckChange = () => {
  const selected = Object.keys(compareCheckMap).filter(k => compareCheckMap[k])
  if (selected.length > 4) {
    ElMessage.warning('最多只能选择4条线路进行对比')
    const lastKey = selected[selected.length - 1]
    compareCheckMap[lastKey] = false
    return
  }
  compareLineNos.value = selected
  if (compareLineNos.value.length >= 2) {
    nextTick(() => {
      renderCompareRadarChart()
    })
  }
}

const getLineName = (lineNo) => {
  const item = healthScores.value.find(s => s.line_no === lineNo)
  return item ? item.line_name + '(' + lineNo + ')' : lineNo
}

const getLineColor = (lineNo) => {
  if (!lineColorMap[lineNo]) {
    const idx = compareLineNos.value.indexOf(lineNo)
    lineColorMap[lineNo] = lineColors[idx % lineColors.length]
  }
  return lineColorMap[lineNo]
}

const getSubScoreList = (item) => [
  item.passenger_intensity,
  item.peak_load_factor,
  item.on_time_rate,
  item.operating_speed,
  item.vehicle_utilization
]

const tagType = (level) => {
  if (level === 'green') return 'success'
  if (level === 'yellow') return 'warning'
  return 'danger'
}

const levelText = (level) => {
  if (level === 'green') return '优秀'
  if (level === 'yellow') return '良好'
  return '待提升'
}

const scoreColor = (score) => {
  if (score >= 80) return '#67c23a'
  if (score >= 60) return '#e6a23c'
  return '#f56c6c'
}

const compareTableData = computed(() => {
  const names = ['客运强度', '高峰满载率', '准点率', '营运速度', '车辆利用率']
  const keys = ['passenger_intensity', 'peak_load_factor', 'on_time_rate', 'operating_speed', 'vehicle_utilization']
  const maxScores = [25, 20, 25, 15, 15]
  return names.map((name, idx) => {
    const row = { name, max_score: maxScores[idx] }
    compareLineNos.value.forEach(ln => {
      const item = healthScores.value.find(s => s.line_no === ln)
      if (item) {
        const sub = item[keys[idx]]
        row[ln + '_score'] = sub.score
        row[ln + '_raw'] = sub.raw_value_str
        row[ln + '_ratio'] = sub.score_ratio
      }
    })
    return row
  })
})

const compareTotalTableData = computed(() => {
  return compareLineNos.value.map(ln => {
    const item = healthScores.value.find(s => s.line_no === ln)
    return item ? {
      line_no: item.line_no,
      line_name: item.line_name + '(' + item.line_no + ')',
      total_score: item.total_score,
      score_level: item.score_level
    } : {}
  }).sort((a, b) => b.total_score - a.total_score)
})

const renderBarChart = () => {
  if (!barChart.value) return
  const chart = echarts.init(barChart.value)
  const sorted = [...healthScores.value].sort((a, b) => b.total_score - a.total_score)
  const lineNames = sorted.map(s => s.line_name + '(' + s.line_no + ')')
  const scores = sorted.map(s => s.total_score)
  const colors = sorted.map(s => scoreColor(s.total_score))
  const lineNos = sorted.map(s => s.line_no)

  chart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params) => {
        const p = params[0]
        const idx = p.dataIndex
        const item = sorted[idx]
        return `<div style="font-weight:600;">${item.line_name} (${item.line_no})</div>
          <div>综合得分: <b>${item.total_score.toFixed(2)}</b> / 100</div>
          <div>等级: ${levelText(item.score_level)}</div>`
      }
    },
    grid: { left: 140, right: 80, top: 20, bottom: 30, containLabel: false },
    xAxis: {
      type: 'value',
      max: 100,
      splitLine: { lineStyle: { type: 'dashed' } }
    },
    yAxis: {
      type: 'category',
      data: lineNames,
      inverse: true,
      axisLabel: { width: 130, overflow: 'truncate' }
    },
    series: [{
      type: 'bar',
      data: scores.map((v, i) => ({
        value: v,
        itemStyle: { color: colors[i] },
        lineNo: lineNos[i]
      })),
      barWidth: 18,
      label: {
        show: true,
        position: 'right',
        formatter: (p) => p.value.toFixed(2) + ' 分',
        fontWeight: 600
      }
    }]
  })

  chart.off('click')
  chart.on('click', (params) => {
    if (params.componentType === 'series') {
      const lineNo = params.data.lineNo
      if (compareMode.value) {
        compareCheckMap[lineNo] = !compareCheckMap[lineNo]
        onCompareCheckChange()
      } else {
        selectedLine.value = healthScores.value.find(s => s.line_no === lineNo)
        nextTick(() => renderRadarChart())
      }
    }
  })

  const resize = () => chart.resize()
  window.addEventListener('resize', resize)
  window._lineHealthBarResize = resize
}

const renderRadarChart = () => {
  if (!radarChart.value || !selectedLine.value) return
  const chart = echarts.init(radarChart.value)
  const sub = getSubScoreList(selectedLine.value)
  chart.setOption({
    tooltip: {},
    radar: {
      indicator: sub.map(s => ({
        name: s.name,
        max: s.max_score
      })),
      radius: '65%',
      axisName: { fontSize: 13 }
    },
    series: [{
      type: 'radar',
      data: [{
        value: sub.map(s => s.score),
        name: selectedLine.value.line_name,
        areaStyle: {
          color: scoreColor(selectedLine.value.total_score),
          opacity: 0.25
        },
        lineStyle: {
          color: scoreColor(selectedLine.value.total_score),
          width: 2
        },
        itemStyle: {
          color: scoreColor(selectedLine.value.total_score)
        }
      }]
    }]
  })
}

const renderCompareRadarChart = () => {
  if (!compareRadarChart.value || compareLineNos.value.length < 2) return
  const chart = echarts.init(compareRadarChart.value)
  const firstItem = healthScores.value.find(s => s.line_no === compareLineNos.value[0])
  const firstSubs = getSubScoreList(firstItem)
  chart.setOption({
    tooltip: {},
    legend: {
      data: compareLineNos.value.map(ln => getLineName(ln))
    },
    radar: {
      indicator: firstSubs.map(s => ({
        name: s.name,
        max: s.max_score
      })),
      radius: '60%',
      axisName: { fontSize: 13 }
    },
    series: [{
      type: 'radar',
      data: compareLineNos.value.map((ln, idx) => {
        const item = healthScores.value.find(s => s.line_no === ln)
        const sub = getSubScoreList(item)
        return {
          value: sub.map(s => s.score),
          name: getLineName(ln),
          areaStyle: { opacity: 0.15 },
          lineStyle: { width: 2 },
          itemStyle: {},
          _color: lineColors[idx % lineColors.length]
        }
      }).map(d => ({
        ...d,
        areaStyle: { ...d.areaStyle, color: d._color },
        lineStyle: { ...d.lineStyle, color: d._color },
        itemStyle: { color: d._color }
      }))
    }]
  })
}

onMounted(async () => {
  healthScores.value = await api.getLineHealthScores()
  healthScores.value.forEach(s => { compareCheckMap[s.line_no] = false })
  await nextTick()
  renderBarChart()
})
</script>

<style scoped>
.score-legend {
  display: flex;
  gap: 24px;
  margin-top: 12px;
  font-size: 13px;
  color: #606266;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}
.legend-color {
  display: inline-block;
  width: 20px;
  height: 12px;
  border-radius: 3px;
}
.legend-green { background: #67c23a; }
.legend-yellow { background: #e6a23c; }
.legend-red { background: #f56c6c; }

.ranking-chart-container {
  width: 100%;
  height: 480px;
  margin-top: 12px;
}

.chart-container {
  width: 100%;
  height: 360px;
  background: #fff;
  border-radius: 6px;
}

.checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 24px;
  padding: 16px;
  background: #fafafa;
  border-radius: 6px;
}
</style>
