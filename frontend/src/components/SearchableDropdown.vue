<template>
  <div class="relative w-full">
    <div :class="[
      'input input-bordered w-full flex flex-wrap gap-1.5 items-start h-auto min-h-[3rem] max-h-[120px] overflow-y-auto py-1.5 px-3 cursor-text transition-shadow',
      { 'input-error': error, 'outline outline-2 outline-offset-2 outline-base-content/20': isOpen }
    ]" @click="focusInput">
      <template v-if="multiple && Array.isArray(modelValue)">
        <span v-for="val in modelValue" :key="val" class="badge badge-primary badge-sm py-3 font-semibold gap-1 z-10">
          {{ getDisplayLabel(val) }}
          <svg xmlns="http://www.w3.org/2000/svg"
            class="h-3.5 w-3.5 cursor-pointer hover:text-base-100 transition-colors" fill="none" viewBox="0 0 24 24"
            stroke="currentColor" @click.stop="removeItem(val)">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </span>
      </template>

      <input ref="searchInput" type="text" v-model="searchQuery" @focus="isOpen = true" @blur="handleBlur"
        @input="isOpen = true" @keydown.enter.prevent="handleEnter" @keydown.backspace="handleBackspace"
        :placeholder="showPlaceholder ? (placeholder || $t('searchableDropdown.searchPlaceholder')) : ''"
        class="flex-1 min-w-[60px] bg-transparent outline-none border-none p-0 m-0 h-full text-sm" />
    </div>

    <ul v-show="isOpen"
      class="absolute z-50 w-full p-2 mt-1 shadow-xl bg-base-100 rounded-box max-h-60 overflow-y-auto border border-base-200">
      <div v-if="multiple && options.length > 0"
        class="flex justify-between gap-2 mb-2 sticky -top-2 bg-base-100 p-2 z-10 border-b border-base-200">
        <button type="button" class="btn btn-xs btn-outline btn-primary flex-1" @mousedown.prevent="selectAll">
          {{ $t('searchableDropdown.selectAll') }}
        </button>
        <button type="button" class="btn btn-xs btn-outline btn-error flex-1" @mousedown.prevent="clearAll">
          {{ $t('searchableDropdown.clearAll') }}
        </button>
      </div>
      <li v-for="option in filteredOptions" :key="option[valueKey]">
        <a :class="[
          'block px-4 py-2 cursor-pointer rounded-lg font-medium flex justify-between items-center transition-colors',
          isSelected(option) ? 'bg-primary/10 text-primary' : 'hover:bg-base-200'
        ]" @mousedown.prevent="selectOption(option)">
          {{ option[labelKey] }}
          <svg v-if="multiple && isSelected(option)" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none"
            viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
        </a>
      </li>
      <li v-if="filteredOptions.length === 0" class="px-4 py-2 text-base-content/50 text-sm">
        {{ $t('searchableDropdown.noResults', { query: searchQuery }) }}
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();

const props = defineProps({
  modelValue: {
    type: [String, Number, Array],
    default: null
  },
  multiple: {
    type: Boolean,
    default: false
  },
  options: {
    type: Array,
    default: () => []
  },
  labelKey: {
    type: String,
    default: 'name'
  },
  valueKey: {
    type: String,
    default: 'id'
  },
  placeholder: {
    type: String,
    default: ''
  },
  error: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['update:modelValue', 'blur']);

const searchQuery = ref('');
const isOpen = ref(false);
const searchInput = ref(null);

const focusInput = () => {
  searchInput.value?.focus();
  isOpen.value = true;
};

const isSelected = (option) => {
  if (props.multiple && Array.isArray(props.modelValue)) {
    return props.modelValue.includes(option[props.valueKey]);
  }
  return props.modelValue === option[props.valueKey];
};

const getDisplayLabel = (val) => {
  const opt = props.options.find(o => o[props.valueKey] === val);
  return opt ? opt[props.labelKey] : val;
};

const showPlaceholder = computed(() => {
  if (props.multiple && Array.isArray(props.modelValue) && props.modelValue.length > 0) {
    return false;
  }
  return true;
});

const filteredOptions = computed(() => {
  let filtered = [...props.options];

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();

    filtered = filtered.filter(opt =>
      String(opt[props.labelKey]).toLowerCase().includes(query)
    );

    filtered.sort((a, b) => {
      const strA = String(a[props.labelKey]).toLowerCase();
      const strB = String(b[props.labelKey]).toLowerCase();

      const aStarts = strA.startsWith(query);
      const bStarts = strB.startsWith(query);

      if (aStarts && !bStarts) return -1;
      if (!aStarts && bStarts) return 1;

      return strA.localeCompare(strB, undefined, { numeric: true, sensitivity: 'base' });
    });
  } else {
    filtered.sort((a, b) => {
      return String(a[props.labelKey]).localeCompare(String(b[props.labelKey]), undefined, { numeric: true, sensitivity: 'base' });
    });
  }

  return filtered;
});

const selectOption = (option) => {
  if (props.multiple) {
    let newValue = Array.isArray(props.modelValue) ? [...props.modelValue] : [];
    const val = option[props.valueKey];

    if (newValue.includes(val)) {
      newValue = newValue.filter(v => v !== val);
    } else {
      newValue.push(val);
    }

    emit('update:modelValue', newValue);
    searchQuery.value = '';

  } else {
    searchQuery.value = option[props.labelKey];
    isOpen.value = false;
    emit('update:modelValue', option[props.valueKey]);
  }

  emit('blur');
};

const handleEnter = () => {
  if (!isOpen.value) return;

  if (searchQuery.value.trim() === '') {
    if (!props.multiple) {
      emit('update:modelValue', null);
    }
    isOpen.value = false;
    emit('blur');

    searchInput.value?.blur();
    return;
  }

  if (filteredOptions.value.length > 0) {
    selectOption(filteredOptions.value[0]);
  }
};

const handleBackspace = () => {
  if (searchQuery.value.length > 0) return;

  if (props.multiple && Array.isArray(props.modelValue) && props.modelValue.length > 0) {
    const newValue = [...props.modelValue];
    newValue.pop();
    emit('update:modelValue', newValue);
  }
};

const removeItem = (val) => {
  if (props.multiple && Array.isArray(props.modelValue)) {
    const newValue = props.modelValue.filter(v => v !== val);
    emit('update:modelValue', newValue);
    emit('blur');
  }
};

const handleBlur = () => {
  isOpen.value = false;

  if (props.multiple) {
    searchQuery.value = '';
  } else {
    if (searchQuery.value.trim() === '') {
      emit('update:modelValue', null);
    } else {
      const selected = props.options.find(opt => opt[props.valueKey] === props.modelValue);
      searchQuery.value = selected ? selected[props.labelKey] : '';
    }
  }

  emit('blur');
};

const selectAll = () => {
  // Map over all available options and grab their value keys
  const allValues = props.options.map(opt => opt[props.valueKey]);
  emit('update:modelValue', allValues);
  searchQuery.value = '';
};

const clearAll = () => {
  // Empty the array
  emit('update:modelValue', []);
  searchQuery.value = '';
};

watch(
  () => [props.modelValue, props.options],
  () => {
    if (!isOpen.value) {
      if (props.multiple) {
        searchQuery.value = '';
      } else {
        const selected = props.options.find(opt => opt[props.valueKey] === props.modelValue);
        searchQuery.value = selected ? selected[props.labelKey] : '';
      }
    }
  },
  { immediate: true, deep: true }
);
</script>