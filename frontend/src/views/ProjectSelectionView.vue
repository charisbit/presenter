<template>
  <div class="project-selection-view">
    <h1>プレゼンテーションを作成</h1>
    <div class="form-container">
      <div class="form-group">
        <label for="project">Backlogプロジェクトを選択</label>
        <select id="project" v-model="selectedProject">
          <option disabled value="">プロジェクトを選択してください</option>
          <option v-for="project in projects" :key="project.id" :value="project.id">
            {{ project.name }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label for="language">言語</label>
        <select id="language" v-model="selectedLanguage">
          <option value="ja">日本語</option>
          <option value="en">English</option>
        </select>
      </div>

      <button @click="generate" :disabled="!canGenerate" class="generate-btn">
        <span v-if="isGenerating">生成中...</span>
        <span v-else>プレゼンテーションを生成</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useSlidesStore } from '@/stores/slides';
import { projectApi } from '@/services/api';
import type { Project, SlideTheme } from '@/types';

const router = useRouter();
const slidesStore = useSlidesStore();

const projects = ref<Project[]>([]);
const selectedProject = ref('');
const fixedThemes: SlideTheme[] = [
  'project_overview',
  'project_progress',
  'issue_management',
  'risk_analysis',
  'team_collaboration',
  'document_management',
  'codebase_activity',
  'notifications',
  'predictive_analysis',
  'summary_plan',
];
const selectedLanguage = ref('ja');
const isGenerating = ref(false);

const canGenerate = computed(() => selectedProject.value);

onMounted(async () => {
  try {
    projects.value = await projectApi.getProjects();
  } catch (error) {
    console.error('Failed to fetch projects:', error);
  }
});

const generate = async () => {
  if (!canGenerate.value) return;

  isGenerating.value = true;
  try {
    const response = await slidesStore.generateSlides({
      projectId: selectedProject.value,
      themes: fixedThemes,
      language: selectedLanguage.value,
    });
    router.push(`/presentation/${response.slideId}`);
  } catch (error) {
    console.error('Failed to generate slides:', error);
    alert('スライドの生成に失敗しました。');
  } finally {
    isGenerating.value = false;
  }
};
</script>

<style scoped>
.project-selection-view {
  padding: 2rem;
  max-width: 800px;
  margin: 0 auto;
}

h1 {
  text-align: center;
  margin-bottom: 2rem;
}

.form-container {
  background: #f8f9fa;
  padding: 2rem;
  border-radius: 8px;
}

.form-group {
  margin-bottom: 1.5rem;
}

label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
}

select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  font-size: 1rem;
}

.generate-btn {
  width: 100%;
  padding: 1rem;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1.2rem;
  cursor: pointer;
  transition: background-color 0.3s;
}

.generate-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
}

.generate-btn:hover:not(:disabled) {
  background: #0056b3;
}
</style>