<template>
  <div class="upload-zone">
    <el-upload
      drag
      :auto-upload="false"
      :limit="1"
      accept=".csv"
      :on-change="handleChange"
      :on-exceed="handleExceed"
    >
      <el-icon class="el-icon--upload"><upload-filled /></el-icon>
      <div class="el-upload__text">
        将CSV文件拖到此处，或<em>点击上传</em>
      </div>
      <template #tip>
        <div class="el-upload__tip">
          只能上传CSV文件，且不超过50MB
        </div>
      </template>
    </el-upload>
    <div v-if="selectedFile" class="selected-file">
      <el-icon color="#67c23a"><CircleCheck /></el-icon>
      <span>{{ selectedFile.name }}</span>
      <el-button type="primary" size="small" :loading="uploading" @click="doUpload" style="margin-left: 12px;">
        开始导入
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const props = defineProps(['dataType'])
const emit = defineEmits(['uploaded'])

const selectedFile = ref(null)
const uploading = ref(false)

const handleChange = (file) => {
  if (!file.name.endsWith('.csv')) {
    ElMessage.error('请上传CSV格式文件')
    return
  }
  selectedFile.value = file.raw
}

const handleExceed = () => {
  ElMessage.warning('只能上传一个文件')
}

const doUpload = async () => {
  if (!selectedFile.value) return
  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('data_type', props.dataType)
    formData.append('file', selectedFile.value)
    const result = await api.uploadData(formData)
    if (result.success) {
      ElMessage.success(result.message)
    } else {
      ElMessage.warning(result.message)
    }
    emit('uploaded', result)
  } finally {
    uploading.value = false
  }
}
</script>

<style scoped>
.selected-file {
  margin-top: 16px;
  padding: 12px 16px;
  background: #f0f9eb;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
