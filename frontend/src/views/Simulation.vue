<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">线路调整模拟</div>

      <el-row :gutter="24">
        <el-col :span="12">
          <div class="panel-title">模拟参数</div>

          <el-form :model="form" label-width="140px" label-position="left">
            <el-form-item label="目标线路">
              <el-select v-model="form.line_no" placeholder="选择线路" filterable style="width: 100%;" @change="onLineChange">
                <el-option v-for="l in lines" :key="l.line_no" :label="l.line_name" :value="l.line_no" />
              </el-select>
            </el-form-item>

            <el-form-item label="参考日期">
              <el-date-picker v-model="form.date" type="date" placeholder="选择参考日期" value-format="YYYY-MM-DD" style="width: 100%;" />
            </el-form-item>

            <el-form-item label="高峰发车间隔(分)">
              <el-input-number v-model="form.peak_interval" :min="3" :max="30" :step="1" :controls-position="'right'" style="width: 100%;" />
            </el-form-item>

            <el-form-item label="平峰发车间隔(分)">
              <el-input-number v-model="form.off_peak_interval" :min="3" :max="30" :step="1" :controls-position="'right'" style="width: 100%;" />
            </el-form-item>

            <el-form-item label="站点增减数">
              <el-input-number v-model="form.station_delta" :min="-5" :max="5" :step="1" :controls-position="'right'" style="width: 100%;" />
              <div class="form-tip">正数=新增站点，负数=裁撤站点（范围 -5 ~ 5）</div>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :icon="RefreshRight" @click="runSimulation" :loading="loading">运行模拟</el-button>
              <el-button @click="resetForm">重置参数</el-button>
            </el-form-item>
          </el-form>
        </el-col>

        <el-col :span="12">
          <div class="panel-title">调整前后对比预览</div>

          <el-row :gutter="12">
            <el-col :span="12">
              <div class="compare-card current">
                <div class="compare-card-title">当前值</div>
                <div class="compare-item" v-if="result">
                  <span class="compare-label">站点数</span>
                  <span class="compare-value">{{ result.orig_station_count }} 站</span>
                </div>
                <div class="compare-item" v-if="result">
                  <span class="compare-label">高峰间隔</span>
                  <span class="compare-value">{{ result.orig_peak_interval }} 分</span>
                </div>
                <div class="compare-item" v-if="result">
                  <span class="compare-label">平峰间隔</span>
                  <span class="compare-value">{{ result.orig_off_peak_interval }} 分</span>
                </div>
                <div class="compare-item" v-if="result">
                  <span class="compare-label">日总班次</span>
                  <span class="compare-value">{{ result.orig_total_trips }} 班</span>
                </div>
                <div class="compare-item" v-if="!result">
                  <span class="placeholder">请先运行模拟</span>
                </div>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="compare-card sim">
                <div class="compare-card-title">模拟值</div>
                <div class="compare-item" v-if="result">
                  <span class="compare-label">站点数</span>
                  <span class="compare-value" :class="deltaClass(result.new_station_count - result.orig_station_count)">
                    {{ result.new_station_count }} 站
                    <span v-if="result.new_station_count !== result.orig_station_count" class="delta">
                      ({{ result.new_station_count > result.orig_station_count ? '+' : '' }}{{ result.new_station_count - result.orig_station_count }})
                    </span>
                  </span>
                </div>
                <div class="compare-item" v-if="result">
                  <span class="compare-label">高峰间隔</span>
                  <span class="compare-value" :class="deltaClass(result.new_peak_interval - result.orig_peak_interval, true)">
                    {{ result.new_peak_interval }} 分
                    <span v-if="result.new_peak_interval !== result.orig_peak_interval" class="delta">
                      ({{ result.new_peak_interval > result.orig_peak_interval ? '+' : '' }}{{ result.new_peak_interval - result.orig_peak_interval }})
                    </span>
                  </span>
                </div>
                <div class="compare-item" v-if="result">
                  <span class="compare-label">平峰间隔</span>
                  <span class="compare-value" :class="deltaClass(result.new_off_peak_interval - result.orig_off_peak_interval, true)">
                    {{ result.new_off_peak_interval }} 分
                    <span v-if="result.new_off_peak_interval !== result.orig_off_peak_interval" class="delta">
                      ({{ result.new_off_peak_interval > result.orig_off_peak_interval ? '+' : '' }}{{ result.new_off_peak_interval - result.orig_off_peak_interval }})
                    </span>
                  </span>
                </div>
                <div class="compare-item" v-if="result">
                  <span class="compare-label">日总班次</span>
                  <span class="compare-value" :class="deltaClass(result.trips_delta)">
                    {{ result.new_total_trips }} 班
                    <span v-if="result.trips_delta !== 0" class="delta">
                      ({{ result.trips_delta > 0 ? '+' : '' }}{{ result.trips_delta }})
                    </span>
                  </span>
                </div>
                <div class="compare-item" v-if="!result">
                  <span class="placeholder">请先运行模拟</span>
                </div>
              </div>
            </el-col>
          </el-row>

          <div class="capacity-box" v-if="result">
            <el-row :gutter="12">
              <el-col :span="12">
                <div class="capacity-label">高峰运力变化</div>
                <div class="capacity-value" :class="deltaClass(result.peak_capacity_change)">
                  {{ result.peak_capacity_change > 0 ? '+' : '' }}{{ result.peak_capacity_change?.toFixed(1) }}%
                </div>
              </el-col>
              <el-col :span="12">
                <div class="capacity-label">可用车辆数</div>
                <div class="capacity-value">
                  {{ result.available_vehicles }} 辆
                </div>
              </el-col>
            </el-row>
            <el-alert v-if="result.capacity_warning" title="运力不足警告：高峰班次超过可用车辆数！" type="error" :closable="false" style="margin-top: 12px;" show-icon />
          </div>
        </el-col>
      </el-row>
    </div>

    <div v-if="result" class="section-card">
      <div class="section-title">目标线路 KPI 对比</div>
      <el-table :data="kpiRows" border stripe>
        <el-table-column prop="name" label="KPI 指标" width="180" align="center" />
        <el-table-column label="调整前" align="center">
          <template #default="{ row }">
            <span :class="row.origClass || ''">{{ row.orig }}</span>
          </template>
        </el-table-column>
        <el-table-column label="调整后" align="center">
          <template #default="{ row }">
            <span :class="row.newClass || ''">{{ row.new }}</span>
          </template>
        </el-table-column>
        <el-table-column label="变化量" width="200" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.delta" :type="row.tagType" size="small">
              {{ row.delta }}
            </el-tag>
            <span v-else class="no-change">无变化</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div v-if="result" class="section-card">
      <div class="section-title">
        相邻线路影响
        <el-tag v-if="adjacentOverloadCount > 0" type="danger" style="margin-left: 12px;">
          {{ adjacentOverloadCount }} 条线路存在过载风险
        </el-tag>
        <el-tag v-else-if="result.adjacent_impacts?.length" type="success" style="margin-left: 12px;">
          {{ result.adjacent_impacts.length }} 条线路受影响
        </el-tag>
      </div>

      <el-table v-if="result.adjacent_impacts?.length" :data="result.adjacent_impacts" border stripe>
        <el-table-column prop="line_no" label="线路编号" width="100" align="center" />
        <el-table-column prop="line_name" label="线路名称" width="180" />
        <el-table-column label="共享站点" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="st in row.shared_stations" :key="st" size="small" style="margin: 2px;">{{ st }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="原高峰满载率" width="140" align="center">
          <template #default="{ row }">{{ (row.orig_peak_load * 100).toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column label="新高峰满载率" width="140" align="center">
          <template #default="{ row }">
            <span :class="row.overload_risk ? 'metric-red' : ''">
              {{ (row.new_peak_load * 100).toFixed(1) }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column label="满载率增量" width="140" align="center">
          <template #default="{ row }">
            <el-tag :type="row.overload_risk ? 'danger' : (row.load_increment > 0 ? 'warning' : 'info')" size="small">
              +{{ (row.load_increment * 100).toFixed(2) }}%
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="风险状态" width="140" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.overload_risk" type="danger" effect="dark">相邻线路过载风险</el-tag>
            <el-tag v-else type="success" effect="plain">正常</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else description="当前调整不影响相邻线路（仅裁撤站点时才会产生相邻影响）" :image-size="100" />
    </div>

    <div v-if="result" class="section-card">
      <div class="section-title">满载率变化趋势（逐步裁撤站点模拟）</div>
      <div class="trend-desc">横轴：裁撤站点数（0~5），纵轴：高峰满载率</div>
      <div ref="trendChart" class="chart-container"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import api from '../api'
import * as echarts from 'echarts'
import { RefreshRight } from '@element-plus/icons-vue'

const lines = ref([])
const loading = ref(false)
const result = ref(null)
const trendChart = ref(null)

const form = ref({
  line_no: '',
  peak_interval: 8,
  off_peak_interval: 15,
  station_delta: 0,
  date: ''
})

const kpiRows = computed(() => {
  if (!result.value) return []
  const r = result.value
  const deltaTrips = r.trips_delta
  const deltaLoad = r.new_kpi.peak_load_factor - r.orig_kpi.peak_load_factor
  const deltaSpeed = r.new_kpi.operating_speed - r.orig_kpi.operating_speed
  const deltaIntensity = r.new_kpi.passenger_intensity - r.orig_kpi.passenger_intensity

  return [
    {
      name: '日总班次数',
      orig: r.orig_kpi.daily_trips + ' 班',
      new: r.new_kpi.daily_trips + ' 班',
      delta: deltaTrips !== 0 ? (deltaTrips > 0 ? '+' : '') + deltaTrips + ' 班' : '',
      tagType: deltaTrips > 0 ? 'success' : (deltaTrips < 0 ? 'warning' : 'info')
    },
    {
      name: '高峰满载率',
      orig: (r.orig_kpi.peak_load_factor * 100).toFixed(1) + '%',
      new: (r.new_kpi.peak_load_factor * 100).toFixed(1) + '%',
      newClass: r.new_kpi.peak_load_factor > 0.9 ? 'metric-red' : (r.new_kpi.peak_load_factor > 0.7 ? 'metric-yellow' : ''),
      delta: deltaLoad !== 0 ? (deltaLoad > 0 ? '+' : '') + (deltaLoad * 100).toFixed(2) + '%' : '',
      tagType: deltaLoad > 0.1 ? 'danger' : (deltaLoad > 0 ? 'warning' : (deltaLoad < 0 ? 'success' : 'info'))
    },
    {
      name: '营运速度',
      orig: r.orig_kpi.operating_speed?.toFixed(2) + ' km/h',
      new: r.new_kpi.operating_speed?.toFixed(2) + ' km/h',
      delta: deltaSpeed !== 0 ? (deltaSpeed > 0 ? '+' : '') + deltaSpeed.toFixed(2) + ' km/h' : '',
      tagType: deltaSpeed > 0 ? 'success' : (deltaSpeed < 0 ? 'warning' : 'info')
    },
    {
      name: '客运强度',
      orig: r.orig_kpi.passenger_intensity?.toFixed(2) + ' 人/km',
      new: r.new_kpi.passenger_intensity?.toFixed(2) + ' 人/km',
      delta: deltaIntensity !== 0 ? (deltaIntensity > 0 ? '+' : '') + deltaIntensity.toFixed(2) + ' 人/km' : '',
      tagType: deltaIntensity > 0 ? 'success' : (deltaIntensity < 0 ? 'warning' : 'info')
    }
  ]
})

const adjacentOverloadCount = computed(() => {
  if (!result.value?.adjacent_impacts) return 0
  return result.value.adjacent_impacts.filter(x => x.overload_risk).length
})

const deltaClass = (delta, inverse = false) => {
  if (delta === 0) return ''
  const good = inverse ? delta < 0 : delta > 0
  return good ? 'metric-green' : 'metric-red'
}

const onLineChange = async () => {
  result.value = null
}

const runSimulation = async () => {
  if (!form.value.line_no) return
  loading.value = true
  try {
    const payload = { ...form.value }
    if (!payload.date) {
      delete payload.date
    }
    result.value = await api.runLineSimulation(payload)
    await nextTick()
    renderTrendChart()
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.value.peak_interval = 8
  form.value.off_peak_interval = 15
  form.value.station_delta = 0
  result.value = null
}

const renderTrendChart = () => {
  if (!trendChart.value || !result.value?.removal_trend) return
  const chart = echarts.init(trendChart.value)
  const data = result.value.removal_trend
  chart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: params => {
        const p = params[0]
        return `裁撤 ${p.name} 个站点<br/>高峰满载率: ${(p.value * 100).toFixed(1)}%`
      }
    },
    grid: { left: 60, right: 40, top: 40, bottom: 50 },
    xAxis: {
      type: 'category',
      name: '裁撤站点数',
      nameLocation: 'middle',
      nameGap: 30,
      data: data.map(d => d.remove_count),
      axisLabel: { formatter: '{value} 个' }
    },
    yAxis: {
      type: 'value',
      name: '高峰满载率',
      min: 0,
      max: 1,
      axisLabel: { formatter: (val) => (val * 100).toFixed(0) + '%' },
      splitLine: { show: true }
    },
    markLine: {
      silent: true,
      data: [{ yAxis: 0.9, name: '过载警戒线', lineStyle: { color: '#f56c6c', type: 'dashed' } }]
    },
    series: [{
      name: '高峰满载率',
      type: 'line',
      smooth: true,
      data: data.map(d => d.peak_load_factor),
      symbol: 'circle',
      symbolSize: 8,
      itemStyle: { color: '#409eff' },
      lineStyle: { width: 3 },
      areaStyle: { opacity: 0.15, color: '#409eff' },
      label: {
        show: true,
        position: 'top',
        formatter: params => (params.value * 100).toFixed(0) + '%'
      },
      markLine: {
        silent: true,
        symbol: 'none',
        lineStyle: { color: '#f56c6c', type: 'dashed', width: 2 },
        data: [
          {
            yAxis: 0.9,
            label: { formatter: '过载警戒线 90%', position: 'end' }
          }
        ]
      }
    }]
  })
}

onMounted(async () => {
  lines.value = await api.getLines()
  const dr = await api.getDateRange()
  if (lines.value.length) {
    form.value.line_no = lines.value[0].line_no
  }
  if (dr.max_date) {
    form.value.date = dr.max_date
  }
})

watch(() => result.value, () => {
  nextTick(() => {
    if (result.value) {
      renderTrendChart()
    }
  })
})
</script>

<style scoped>
.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 10px;
  border-bottom: 1px dashed #ebeef5;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.compare-card {
  border-radius: 8px;
  padding: 20px;
  border: 1px solid #ebeef5;
}

.compare-card.current {
  background: linear-gradient(135deg, #f5f7fa 0%, #e8ecf1 100%);
}

.compare-card.sim {
  background: linear-gradient(135deg, #ecf5ff 0%, #d9ecff 100%);
  border-color: #a0cfff;
}

.compare-card-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 14px;
  padding-bottom: 10px;
  border-bottom: 1px dashed rgba(0, 0, 0, 0.08);
}

.compare-card.current .compare-card-title {
  color: #606266;
}

.compare-card.sim .compare-card-title {
  color: #409eff;
}

.compare-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  font-size: 13px;
}

.compare-label {
  color: #606266;
}

.compare-value {
  font-weight: 600;
  color: #303133;
  position: relative;
}

.compare-value .delta {
  font-size: 11px;
  font-weight: 500;
  margin-left: 4px;
}

.placeholder {
  color: #c0c4cc;
  font-size: 12px;
  text-align: center;
  display: block;
  padding: 20px 0;
}

.capacity-box {
  margin-top: 20px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #ebeef5;
}

.capacity-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
  text-align: center;
}

.capacity-value {
  font-size: 22px;
  font-weight: 700;
  color: #303133;
  text-align: center;
}

.trend-desc {
  font-size: 13px;
  color: #909399;
  margin-bottom: 12px;
  padding-left: 4px;
}

.no-change {
  color: #909399;
  font-size: 13px;
}
</style>
