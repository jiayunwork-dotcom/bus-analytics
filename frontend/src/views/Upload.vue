<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">数据导入</div>
      
      <el-alert
        title="数据格式说明"
        type="info"
        :closable="false"
        style="margin-bottom: 20px;"
      >
        <p>请上传CSV格式文件，支持以下4类数据：</p>
        <ul style="margin: 8px 0 0 20px;">
          <li><strong>线路基础信息</strong>：线路编号、线路名、起讫站、全程公里数、站点数量、票价（可选：直线距离）</li>
          <li><strong>班次运行记录</strong>：线路编号、日期、班次号、实发时间、计划时间、车辆编号、方向（可选：驾驶员）</li>
          <li><strong>站点客流</strong>：线路编号、日期、班次号、站点序号、站点名、上客人数、下客人数、刷卡人数（可选：刷卡人ID）</li>
          <li><strong>车辆里程</strong>：车辆编号、日期、总里程、营运里程</li>
        </ul>
      </el-alert>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="线路基础信息" name="routes">
          <UploadZone data-type="routes" @uploaded="handleUploaded" />
        </el-tab-pane>
        <el-tab-pane label="班次运行记录" name="trips">
          <UploadZone data-type="trips" @uploaded="handleUploaded" />
        </el-tab-pane>
        <el-tab-pane label="站点客流" name="flows">
          <UploadZone data-type="flows" @uploaded="handleUploaded" />
        </el-tab-pane>
        <el-tab-pane label="车辆里程" name="mileages">
          <UploadZone data-type="mileages" @uploaded="handleUploaded" />
        </el-tab-pane>
      </el-tabs>
    </div>

    <div v-if="lastResult" class="section-card">
      <div class="section-title">导入结果</div>
      <el-result
        :icon="lastResult.success ? 'success' : 'warning'"
        :title="lastResult.success ? '导入完成' : '导入部分失败'"
        :sub-title="lastResult.message"
      />
      <div v-if="lastResult.errors && lastResult.errors.length" style="margin-top: 20px;">
        <div class="error-title">错误详情（前20条）：</div>
        <el-table :data="lastResult.errors.slice(0, 20)" border size="small">
          <el-table-column prop="row" label="行号" width="80" align="center" />
          <el-table-column label="缺失字段">
            <template #default="{ row }">
              <el-tag
                v-for="f in row.fields"
                :key="f"
                type="danger"
                size="small"
                style="margin-right: 4px;"
              >
                {{ f }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <div class="section-card">
      <div class="section-title">示例数据下载</div>
      <el-row :gutter="16">
        <el-col v-for="t in templates" :key="t.type" :span="6">
          <el-card shadow="hover" style="cursor: pointer;" @click="downloadTemplate(t.type)">
            <div style="text-align: center;">
              <el-icon :size="32" color="#409eff"><Document /></el-icon>
              <div style="margin-top: 8px; font-weight: 500;">{{ t.name }}</div>
              <div style="color: #909399; font-size: 12px; margin-top: 4px;">点击下载示例CSV</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import UploadZone from '../components/UploadZone.vue'

const activeTab = ref('routes')
const lastResult = ref(null)

const templates = [
  { type: 'routes', name: '线路基础信息' },
  { type: 'trips', name: '班次运行记录' },
  { type: 'flows', name: '站点客流' },
  { type: 'mileages', name: '车辆里程' },
]

const handleUploaded = (result) => {
  lastResult.value = result
}

const downloadTemplate = (type) => {
  const content = generateTemplate(type)
  const blob = new Blob(['\ufeff' + content], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${type}_template.csv`
  a.click()
  URL.revokeObjectURL(url)
}

const generateTemplate = (type) => {
  const templates = {
    routes: '线路编号,线路名,起讫站,全程公里数,站点数量,票价,直线距离\n1路,1路(火车站-科技园),火车站-科技园,15.5,28,2.0,12.3\n2路,2路(汽车站-大学城),汽车站-大学城,20.2,35,2.0,16.8',
    trips: '线路编号,日期,班次号,实发时间,计划时间,车辆编号,驾驶员,方向\n1路,2024-01-15,1001,07:00,07:00,粤B12345,张师傅,上行\n1路,2024-01-15,1002,07:12,07:10,粤B12345,张师傅,下行',
    flows: '线路编号,日期,班次号,站点序号,站点名,上客人数,下客人数,刷卡人数,刷卡人ID\n1路,2024-01-15,1001,1,火车站,35,0,25,C001\n1路,2024-01-15,1001,2,人民广场,18,12,15,C002',
    mileages: '车辆编号,日期,总里程,营运里程\n粤B12345,2024-01-15,320.5,285.0\n粤B67890,2024-01-15,298.3,265.5'
  }
  return templates[type] || ''
}
</script>

<style scoped>
.error-title {
  color: #f56c6c;
  font-weight: 600;
  margin-bottom: 12px;
}
</style>
