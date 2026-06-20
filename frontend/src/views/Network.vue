<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">线网评价</div>
      <el-row :gutter="16" style="margin-bottom: 20px;">
        <el-col :span="6">
          <el-form-item label="城区面积(km²)">
            <el-input-number v-model="cityArea" :min="1" :max="100000" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="不重复线网里程(km)">
            <el-input-number v-model="uniqueLineKm" :min="0" :max="100000" />
          </el-form-item>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="loadMetrics">重新计算</el-button>
        </el-col>
      </el-row>
    </div>

    <el-row :gutter="20">
      <el-col :span="8">
        <div class="section-card gauge-card">
          <div class="section-title">站点覆盖率</div>
          <div ref="coverageGauge" style="height: 280px;"></div>
          <div class="gauge-info">
            <div>站点数：{{ metrics.station_count || 0 }} 个</div>
            <div>行业参考值：≥ 3 站/km²</div>
          </div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="section-card gauge-card">
          <div class="section-title">线路重复系数</div>
          <div ref="dupGauge" style="height: 280px;"></div>
          <div class="gauge-info">
            <div>总线路里程：{{ metrics.total_line_km?.toFixed(1) }} km</div>
            <div>不重复里程：{{ metrics.unique_line_km?.toFixed(1) }} km</div>
            <div>行业参考值：1.3 ~ 2.0</div>
          </div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="section-card gauge-card">
          <div class="section-title">非直线系数</div>
          <div ref="nslGauge" style="height: 280px;"></div>
          <div class="gauge-info">
            <div>当前值：{{ metrics.non_straight_line?.toFixed(2) }}</div>
            <div>行业参考值：≤ 1.4</div>
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import api from '../api'
import * as echarts from 'echarts'

const cityArea = ref(1000)
const uniqueLineKm = ref(500)
const metrics = ref({})
const coverageGauge = ref(null)
const dupGauge = ref(null)
const nslGauge = ref(null)

const loadMetrics = async () => {
  metrics.value = await api.getNetworkMetrics({
    city_area: cityArea.value,
    unique_line_km: uniqueLineKm.value
  })
  await nextTick()
  renderGauges()
}

const renderGauges = () => {
  if (coverageGauge.value) {
    echarts.init(coverageGauge.value).setOption({
      series: [{
        type: 'gauge',
        progress: { show: true, width: 18 },
        axisLine: { lineStyle: { width: 18 } },
        axisTick: { show: false },
        splitLine: { length: 15, lineStyle: { width: 2 } },
        axisLabel: { distance: 25, fontSize: 12 },
        pointer: { width: 6 },
        detail: { valueAnimation: true, formatter: '{value} 站/km²', fontSize: 18, offsetCenter: [0, '70%'] },
        data: [{ value: +(metrics.value.station_coverage || 0).toFixed(2) }],
        max: Math.max(10, +(metrics.value.station_coverage || 0) * 1.5)
      }]
    })
  }
  if (dupGauge.value) {
    echarts.init(dupGauge.value).setOption({
      series: [{
        type: 'gauge',
        min: 1,
        max: 4,
        progress: { show: true, width: 18 },
        axisLine: { lineStyle: { width: 18, color: [[0.325, '#67c23a'], [0.5, '#e6a23c'], [1, '#f56c6c']] } },
        axisTick: { show: false },
        splitLine: { length: 15, lineStyle: { width: 2 } },
        axisLabel: { distance: 25, fontSize: 12 },
        pointer: { width: 6 },
        detail: { valueAnimation: true, formatter: '{value}', fontSize: 18, offsetCenter: [0, '70%'] },
        data: [{ value: +(metrics.value.line_duplication || 0).toFixed(2) }]
      }]
    })
  }
  if (nslGauge.value) {
    echarts.init(nslGauge.value).setOption({
      series: [{
        type: 'gauge',
        min: 1,
        max: 2.5,
        progress: { show: true, width: 18 },
        axisLine: { lineStyle: { width: 18, color: [[0.4, '#67c23a'], [0.67, '#e6a23c'], [1, '#f56c6c']] } },
        axisTick: { show: false },
        splitLine: { length: 15, lineStyle: { width: 2 } },
        axisLabel: { distance: 25, fontSize: 12 },
        pointer: { width: 6 },
        detail: { valueAnimation: true, formatter: '{value}', fontSize: 18, offsetCenter: [0, '70%'] },
        data: [{ value: +(metrics.value.non_straight_line || 0).toFixed(2) }]
      }]
    })
  }
}

onMounted(loadMetrics)
</script>

<style scoped>
.gauge-card {
  text-align: center;
}
.gauge-info {
  margin-top: 8px;
  font-size: 13px;
  color: #606266;
  line-height: 1.8;
}
</style>
