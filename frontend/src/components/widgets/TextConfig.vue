<template>
  <div class="flex flex-col gap-5">
    
    <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
      <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('textWidget.config.textEditor') }}</h4>
      
      <!-- Unified Toolbar -->
      <div class="flex flex-wrap items-center gap-2 mb-2 bg-base-100 p-2 rounded border border-base-300 shadow-sm">
        
        <!-- Text Styling -->
        <div class="join border border-base-300">
          <button @click.prevent="formatText('bold')" class="btn btn-sm btn-ghost join-item px-3 font-bold" :title="$t('textWidget.config.bold')">B</button>
          <button @click.prevent="formatText('italic')" class="btn btn-sm btn-ghost join-item px-3 italic" :title="$t('textWidget.config.italic')">I</button>
          <button @click.prevent="formatText('underline')" class="btn btn-sm btn-ghost join-item px-3 underline" :title="$t('textWidget.config.underline')">U</button>
        </div>

        <!-- Lists -->
        <div class="join border border-base-300">
          <button @click.prevent="formatText('insertUnorderedList')" class="btn btn-sm btn-ghost join-item px-3" :title="$t('textWidget.config.bulletList')">• List</button>
          <button @click.prevent="formatText('insertOrderedList')" class="btn btn-sm btn-ghost join-item px-3" :title="$t('textWidget.config.numberedList')">1. List</button>
        </div>

        <div class="divider divider-horizontal mx-0"></div>

        <!-- Block Alignment -->
        <div class="join border border-base-300">
          <button @click.prevent="localConfig.textAlign = 'left'" :class="{ 'bg-base-300': localConfig.textAlign === 'left' }" class="btn btn-sm btn-ghost join-item px-2" :title="$t('textWidget.config.alignLeft')">
            <Icon icon="lucide:align-left" class="w-4 h-4" />
          </button>
          <button @click.prevent="localConfig.textAlign = 'center'" :class="{ 'bg-base-300': localConfig.textAlign === 'center' }" class="btn btn-sm btn-ghost join-item px-2" :title="$t('textWidget.config.alignCenter')">
            <Icon icon="lucide:align-center" class="w-4 h-4" />
          </button>
          <button @click.prevent="localConfig.textAlign = 'right'" :class="{ 'bg-base-300': localConfig.textAlign === 'right' }" class="btn btn-sm btn-ghost join-item px-2" :title="$t('textWidget.config.alignRight')">
            <Icon icon="lucide:align-right" class="w-4 h-4" />
          </button>
        </div>

        <div class="divider divider-horizontal mx-0"></div>

        <!-- Font Size & Color -->
        <div class="flex items-center gap-2 ml-auto">
          <div class="tooltip tooltip-bottom" :data-tip="$t('textWidget.config.baseFontSize')">
            <input type="number" v-model.number="localConfig.fontSize" class="input input-bordered input-sm w-16 text-center" placeholder="14" />
          </div>
          <div class="tooltip tooltip-bottom" :data-tip="$t('common.textColor')">
            <input type="color" v-model="localConfig.textColor" class="h-8 w-10 cursor-pointer rounded border border-base-300 p-0" />
          </div>
        </div>
      </div>

      <!-- Editable Content Area -->
      <div 
        ref="editorRef"
        contenteditable="true"
        @input="updateContent"
        @blur="updateContent"
        class="rich-text-container textarea textarea-bordered w-full min-h-[200px] max-h-[300px] overflow-y-auto bg-base-100 focus:outline-none focus:border-primary"
        :style="{ 
          textAlign: localConfig.textAlign, 
          fontSize: localConfig.fontSize + 'px', 
          color: localConfig.textColor 
        }"
      ></div>
    </div>

    <!-- Widget Options -->
    <div class="px-2">
      <label class="cursor-pointer flex items-center gap-4">
        <input type="checkbox" v-model="localConfig.showHeader" class="toggle toggle-primary toggle-sm" />
        <span class="label-text font-semibold">{{ $t('textWidget.config.showHeaderCard') }}</span>
      </label>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue'; 

const { t } = useI18n();

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(['update:modelValue']);
const editorRef = ref(null);

const localConfig = ref({
  content: props.modelValue.content || '',
  showHeader: props.modelValue.showHeader !== undefined ? props.modelValue.showHeader : true,
  textAlign: props.modelValue.textAlign || 'left',
  fontSize: props.modelValue.fontSize !== undefined ? props.modelValue.fontSize : 14,
  textColor: props.modelValue.textColor || '#3584e4'
});

const formatText = (command) => {
  document.execCommand(command, false, null);
  editorRef.value.focus();
  updateContent();
};

const updateContent = () => {
  if (editorRef.value) {
    localConfig.value.content = editorRef.value.innerHTML;
  }
};

watch(localConfig, (newVal) => {
  emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
  if (editorRef.value) {
    editorRef.value.innerHTML = localConfig.value.content || t('textWidget.config.placeholder');
  }
  emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>

<style scoped>
.rich-text-container :deep(b),
.rich-text-container :deep(strong) {
  font-weight: 700;
}

.rich-text-container :deep(i),
.rich-text-container :deep(em) {
  font-style: italic;
}

.rich-text-container :deep(u) {
  text-decoration: underline;
}

.rich-text-container :deep(ul) {
  list-style-type: disc;
  padding-left: 1.5rem;
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
}

.rich-text-container :deep(ol) {
  list-style-type: decimal;
  padding-left: 1.5rem;
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
}

.rich-text-container :deep(li) {
  margin-bottom: 0.25rem;
}
</style>