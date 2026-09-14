<template>
    <div class="flex flex-col gap-5">

        <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
            <h4 class="font-bold text-sm mb-3 text-base-content">{{ $t('visualizationWidget.config.imageSource') }}</h4>

            <div class="grid grid-cols-1 gap-4">

                <div class="form-control w-full">
                    <div class="label pb-1">
                        <span class="label-text font-semibold">{{ $t('visualizationWidget.config.uploadNewImage') }}</span>
                    </div>
                    <div class="flex items-center gap-2">
                        <input type="file" accept="image/png, image/jpeg, image/gif, image/webp"
                            class="file-input file-input-bordered file-input-sm w-full" :disabled="isUploading"
                            @change="handleImageUpload" />
                        <span v-if="isUploading"
                            class="loading loading-spinner loading-md text-primary shrink-0"></span>
                    </div>

                    <div v-if="uploadError" class="text-xs text-error mt-1 font-semibold">
                        {{ uploadError.message || $t('visualizationWidget.messages.uploadFailed') }}
                    </div>
                </div>

                <div class="divider text-xs opacity-50 my-0">{{ $t('visualizationWidget.config.orPasteUrl') }}</div>

                <label class="form-control w-full">
                    <div class="label pb-1">
                        <span class="label-text font-semibold">{{ $t('visualizationWidget.config.imageUrlLabel') }}</span>
                    </div>
                    <input type="text" v-model="localConfig.imageUrl"
                        class="input input-bordered input-sm w-full font-mono"
                        placeholder="https://example.com/floorplan.png" />
                </label>

                <label class="form-control w-full">
                    <div class="label pb-1"><span class="label-text font-semibold">{{ $t('visualizationWidget.config.imageFit') }}</span></div>
                    <select v-model="localConfig.imageFit" class="select select-bordered select-sm w-full">
                        <option value="contain">{{ $t('visualizationWidget.config.fitContain') }}</option>
                        <option value="cover">{{ $t('visualizationWidget.config.fitCover') }}</option>
                        <option value="fill">{{ $t('visualizationWidget.config.fitFill') }}</option>
                    </select>
                </label>

                <label class="cursor-pointer label justify-start gap-4 w-max mt-2">
                    <input type="checkbox" v-model="localConfig.showHeader" class="toggle toggle-primary toggle-sm" />
                    <span class="label-text font-semibold">{{ $t('visualizationWidget.config.showTopHeader') }}</span>
                </label>
            </div>
        </div>

        <div class="p-4 bg-base-200/50 rounded-box border border-base-200">
            <div class="flex justify-between items-center mb-3">
                <h4 class="font-bold text-sm text-base-content m-0">{{ $t('visualizationWidget.config.liveDataOverlay') }}</h4>
                <input type="checkbox" v-model="localConfig.showOverlay" class="toggle toggle-secondary toggle-sm" />
            </div>

            <div v-if="localConfig.showOverlay" class="grid grid-cols-2 gap-4 mt-2">
                <label class="form-control w-full">
                    <div class="label pb-1"><span class="label-text font-semibold">{{ $t('visualizationWidget.config.overlayPosition') }}</span></div>
                    <select v-model="localConfig.overlayPosition" class="select select-bordered select-sm w-full">
                        <option value="bottom-right">{{ $t('visualizationWidget.config.positions.bottomRight') }}</option>
                        <option value="bottom-left">{{ $t('visualizationWidget.config.positions.bottomLeft') }}</option>
                        <option value="top-right">{{ $t('visualizationWidget.config.positions.topRight') }}</option>
                        <option value="top-left">{{ $t('visualizationWidget.config.positions.topLeft') }}</option>
                    </select>
                </label>

                <label class="form-control w-full">
                    <div class="label pb-1"><span class="label-text font-semibold">{{ $t('visualizationWidget.config.bgColor') }}</span></div>
                    <input type="color" v-model="localConfig.overlayBgColor"
                        class="h-8 w-full cursor-pointer rounded border border-base-300 p-0" />
                </label>

                <label class="form-control w-full">
                    <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.textColor') }}</span></div>
                    <input type="color" v-model="localConfig.overlayTextColor"
                        class="h-8 w-full cursor-pointer rounded border border-base-300 p-0" />
                </label>

                <label class="form-control w-full">
                    <div class="label pb-1"><span class="label-text font-semibold">{{ $t('common.unitSuffix') }}</span></div>
                    <input type="text" v-model="localConfig.unit" class="input input-bordered input-sm w-full"
                        :placeholder="$t('common.unitPlaceholder')" />
                </label>

                <label class="form-control w-full sm:col-span-2">
                    <div class="label pb-1">
                        <span class="label-text font-semibold">{{ $t('common.decimalPlaces') }}</span>
                    </div>
                    <input type="number" v-model="localConfig.decimalPlaces" min="0" max="4"
                        class="input input-bordered input-sm w-full" />
                </label>
            </div>

            <div v-if="localConfig.showOverlay && activeDevices.length === 0"
                class="text-xs text-error font-semibold mt-3">
                {{ $t('visualizationWidget.config.noDevicesWarning') }}
            </div>
        </div>

    </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useMutation } from '@/composables/useMutation'; 

const { t } = useI18n();

const props = defineProps({
    modelValue: { type: Object, default: () => ({}) },
    selectedDeviceIds: { type: Array, default: () => [] }
});

const emit = defineEmits(['update:modelValue']);

const { isLoading: isUploading, error: uploadError, data: uploadData, execute: uploadApi } = useMutation();
const allowedMimeTypes = ['image/png', 'image/jpeg', 'image/gif', 'image/webp'];
const activeDevices = computed(() => props.selectedDeviceIds || []);

const localConfig = ref({
    imageUrl: props.modelValue.imageUrl || '',
    imageFit: props.modelValue.imageFit || 'contain',
    showHeader: props.modelValue.showHeader !== undefined ? props.modelValue.showHeader : true,
    showOverlay: props.modelValue.showOverlay !== undefined ? props.modelValue.showOverlay : true,
    overlayPosition: props.modelValue.overlayPosition || 'bottom-right',
    overlayBgColor: props.modelValue.overlayBgColor || '#ffffff',
    overlayTextColor: props.modelValue.overlayTextColor || '#334155',
    unit: props.modelValue.unit || '',
    decimalPlaces: props.modelValue.decimalPlaces !== undefined ? props.modelValue.decimalPlaces : 1
});

const handleImageUpload = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    // Check MIME type
    if (!allowedMimeTypes.includes(file.type)) {
        uploadError.value = { message: t('visualizationWidget.messages.fileNotAllowed') };
        event.target.value = ''; // Reset file input
        return;
    }

    // Optional: Add a max size check (e.g., 10MB)
    const maxSize = 10 * 1024 * 1024;
    if (file.size > maxSize) {
        uploadError.value = { message: t('visualizationWidget.messages.fileTooLarge') };
        event.target.value = '';
        return;
    }

    const formData = new FormData();
    formData.append('image', file);

    const success = await uploadApi('/file/image/upload', formData, 'POST', 'formData');

    if (success && uploadData.value?.data?.url) {
        localConfig.value.imageUrl = import.meta.env.VITE_API_BASE_URL + uploadData.value.data.url;
    }
};

watch(localConfig, (newVal) => {
    emit('update:modelValue', JSON.parse(JSON.stringify(newVal)));
}, { deep: true });

onMounted(() => {
    emit('update:modelValue', JSON.parse(JSON.stringify(localConfig.value)));
});
</script>